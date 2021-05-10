package main

//go:generate bash ./scripts/generate.sh

import (
	"compress/flate"
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/rate-limit-go"
	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/api"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelpers "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/middleware"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/session"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	influx "github.com/influxdata/influxdb1-client"
	"go.uber.org/zap"
)

type contextKey string

const (
	keyField = "key"

	ctxUserIDField    contextKey = "user_id"
	ctxUserLevelField contextKey = "user_level"
)

type Server struct {
}

func main() {

	err := config.Init(helpers.GetIP())
	log.InitZap(log.LogNameAPI)
	defer log.Flush()
	if err != nil {
		log.ErrS(err)
		return
	}

	session.Init()

	r := chi.NewRouter()
	r.Use(fixRequestURLMiddleware) // Needed for FindRoute()
	r.Use(chiMiddleware.Compress(flate.DefaultCompression))
	r.Use(middleware.RealIP)

	r.Get("/", rootHandler)
	r.Get("/health-check", healthCheckHandler)

	generated.HandlerWithOptions(Server{}, generated.ChiServerOptions{
		BaseRouter: r,
		Middlewares: []generated.MiddlewareFunc{
			// validateMiddlewear,
			rateLimitMiddlewear,
			authMiddlewear,
		},
	})

	r.NotFound(notFoundHandler)

	s := &http.Server{
		Addr:              "0.0.0.0:" + config.C.APIPort,
		Handler:           r,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	log.Info("Starting API on " + "http://" + s.Addr + "/")

	go func() {
		err = s.ListenAndServe()
		if err != nil {
			log.ErrS(err)
		}
	}()

	helpers.KeepAlive(
		mysql.Close,
		mongo.Close,
		memcache.Close,
	)
}

// Handlers
func rootHandler(w http.ResponseWriter, r *http.Request) {
	returnResponse(w, r, http.StatusOK, map[string]interface{}{
		"docs":    config.C.GlobalSteamDomain + "/api/globalsteam",
		"support": config.C.DiscordServerInviteURL,
	})
}

func rateLimitedHandler(w http.ResponseWriter, r *http.Request) {
	returnResponse(w, r, http.StatusTooManyRequests, generated.MessageResponse{Error: http.StatusText(http.StatusTooManyRequests)})
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	returnResponse(w, r, http.StatusNotFound, generated.MessageResponse{Error: "Invalid endpoint"})
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
}

// var validateOptions = &codegenMiddleware.Options{
// 	Options: openapi3filter.Options{
// 		MultiError:            true,
// 		IncludeResponseStatus: true,
// 		AuthenticationFunc: func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
// 			return nil // Skip auth check for now, done elseware
// 		},
// 	},
// }

// Middlewares
// func validateMiddlewear(next http.HandlerFunc) http.HandlerFunc {
// 	return codegenMiddleware.OapiRequestValidatorWithOptions(api.GetGlobalSteamResolved(), validateOptions)(next).ServeHTTP
// }

var apiKeyRegexp = regexp.MustCompile("^[A-Z0-9]{20}$")

func authMiddlewear(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// Check API key
		key := func() string {

			key := r.URL.Query().Get(keyField)
			if key != "" {
				w.Header().Set("Authed-With", "query")
				return key
			}

			key = strings.TrimPrefix(r.Header.Get(keyField), "Bearer ")
			if key != "" {
				w.Header().Set("Authed-With", "bearer")
				return key
			}

			key = session.Get(r, session.SessionUserAPIKey)
			if key != "" {
				w.Header().Set("Authed-With", "session")
				return key
			}

			return ""
		}()

		if key == "" {
			returnResponse(w, r, http.StatusUnauthorized, generated.MessageResponse{Error: "empty api key"})
			return
		}

		if !apiKeyRegexp.MatchString(key) {
			returnResponse(w, r, http.StatusUnauthorized, generated.MessageResponse{Error: "invalid api key: " + key})
			return
		}

		// Check user has access to api
		user, err := mysql.GetUserByAPIKey(key)
		if err == mysql.ErrRecordNotFound {
			returnResponse(w, r, http.StatusUnauthorized, generated.MessageResponse{Error: "invalid api key: " + key})
			return
		}
		if err != nil {
			returnResponse(w, r, http.StatusInternalServerError, err)
			return
		}

		route, _, err := api.GetRouter().FindRoute(r)
		if err != nil {
			log.Err("missing route", zap.Error(err), zap.String("method", r.Method), zap.String("url", r.URL.String()))
			notFoundHandler(w, r)
			return
		}
		if user.Level < mysql.UserLevel2 && !helpers.SliceHasString(api.TagPublic, route.Operation.Tags) {
			returnResponse(w, r, http.StatusUnauthorized, generated.MessageResponse{Error: "Invalid user level"})
			return
		}

		// Save user info to context
		r = r.WithContext(context.WithValue(r.Context(), ctxUserIDField, user.ID))
		r = r.WithContext(context.WithValue(r.Context(), ctxUserLevelField, user.Level))

		next.ServeHTTP(w, r)
	}
}

var (
	donatorLimiter = rate.New(time.Second*1, rate.WithBurst(10))
	publicLimiter  = rate.New(time.Second*5, rate.WithBurst(1))
)

func rateLimitMiddlewear(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		level, _ := r.Context().Value(ctxUserLevelField).(mysql.UserLevel)

		var limiters *rate.Limiters
		if level > mysql.UserLevelFree {
			limiters = donatorLimiter
		} else {
			limiters = publicLimiter
		}

		reservation := limiters.GetLimiter(r.RemoteAddr).Reserve()
		if reservation.Delay() > 0 {

			middleware.SetRateLimitHeaders(w, limiters, reservation)
			rateLimitedHandler(w, r)
			reservation.Cancel()
			return
		}

		next.ServeHTTP(w, r)
	}
}

func fixRequestURLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if config.IsLocal() {
			r.URL.Scheme = "http"
			r.URL.Host = r.Host
		} else {
			r.URL.Scheme = "https"
			r.URL.Host = "api.globalsteam.online"
		}

		next.ServeHTTP(w, r)
	})
}

//
func returnResponse(w http.ResponseWriter, r *http.Request, code int, i interface{}) {

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)

	if val, ok := i.(error); ok {
		i = generated.MessageResponse{Error: val.Error()}
	}

	err := json.NewEncoder(w).Encode(i)
	if err != nil {
		log.Err("encoding response", zap.Error(err))
	}

	if config.IsProd() {
		go func() {

			userID, _ := r.Context().Value(ctxUserIDField).(int)
			if userID == 1 {
				return
			}

			point := influx.Point{
				Measurement: string(influxHelpers.InfluxMeasurementAPICalls),
				Tags: map[string]string{
					"path":    r.URL.Path,
					"user_id": strconv.Itoa(userID),
					"code":    strconv.Itoa(code),
				},
				Fields: map[string]interface{}{
					"call": 1,
				},
				Time:      time.Now(),
				Precision: "s",
			}

			_, err := influxHelpers.InfluxWrite(influxHelpers.InfluxRetentionPolicyAllTime, point)
			if err != nil {
				log.Err("saving to influx", zap.Error(err))
			}
		}()
	}
}
