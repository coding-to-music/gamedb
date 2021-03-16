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

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/api"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelpers "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/middleware"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
	influx "github.com/influxdata/influxdb1-client"
)

const (
	keyField = "key"

	ctxUserIDField    = "user_id"
	ctxUserLevelField = "user_level"
)

var apiKeyRegexp = regexp.MustCompile("^[A-Z0-9]{20}$")

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
	r.Use(chiMiddleware.Compress(flate.DefaultCompression))
	r.Use(middleware.RealIP)
	r.Use(middleware.RateLimiterBlock(time.Second, 1, rateLimitedHandler))
	// r.Use(codegenMiddleware.OapiRequestValidatorWithOptions(api.SwaggerGameDB, &codegenMiddleware.Options{Options: openapi3filter.Options{MultiError: true}}))

	r.Get("/health-check", healthCheckHandler)

	r.NotFound(notFoundHandler)

	generated.HandlerWithOptions(Server{}, generated.ChiServerOptions{
		BaseRouter:  r,
		Middlewares: []generated.MiddlewareFunc{authMiddlewear},
	})

	s := &http.Server{
		Addr:              "0.0.0.0:" + config.C.APIPort,
		Handler:           r,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	if config.IsLocal() {
		s.Addr = "localhost:" + config.C.APIPort
	}

	log.Info("Starting API on " + "http://" + s.Addr + "/games")

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

func rateLimitedHandler(w http.ResponseWriter, r *http.Request) {
	returnResponse(w, r, http.StatusTooManyRequests, generated.MessageResponse{Error: http.StatusText(http.StatusTooManyRequests)})
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {

	w.WriteHeader(404)

	b, err := json.Marshal(generated.MessageResponse{Error: "Invalid endpoint"})
	if err != nil {
		log.ErrS(err)
	}

	_, err = w.Write(b)
	if err != nil {
		log.ErrS(err)
	}
}

var router *openapi3filter.Router

func authMiddlewear(next http.HandlerFunc) http.HandlerFunc {

	if router == nil {

		s := &*api.GetGlobalSteam()

		err := openapi3.NewSwaggerLoader().ResolveRefsIn(s, nil)
		if err != nil {
			log.ErrS(err)
		}

		router = openapi3filter.NewRouter().WithSwagger(s)
	}

	return func(w http.ResponseWriter, r *http.Request) {

		// Check API key
		key := r.URL.Query().Get(keyField)
		if key == "" {
			key = strings.TrimLeft(r.Header.Get(keyField), "Bearer ")
			if key == "" {
				key = session.Get(r, session.SessionUserAPIKey)
			}
		}

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

		r.URL.Host = r.Host
		r.URL.Scheme = "https"

		route, _, err := router.FindRoute(r.Method, r.URL)
		if err != nil {
			returnResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		if user.Level < mysql.UserLevel2 && !helpers.SliceHasString(api.TagFree, route.Operation.Tags) {
			returnResponse(w, r, http.StatusUnauthorized, generated.MessageResponse{Error: "invalid user level"})
			return
		}

		// Save user ID to context
		r = r.WithContext(context.WithValue(r.Context(), ctxUserIDField, user.ID))
		r = r.WithContext(context.WithValue(r.Context(), ctxUserLevelField, user.Level))

		next.ServeHTTP(w, r)
	}
}

func returnResponse(w http.ResponseWriter, r *http.Request, code int, i interface{}) {

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)

	if val, ok := i.(error); ok {
		i = generated.MessageResponse{Error: val.Error()}
	}

	err := json.NewEncoder(w).Encode(i)
	if err != nil {
		log.ErrS(err)
	}

	if config.IsProd() {
		go func() {

			userID, _ := r.Context().Value(ctxUserIDField).(int)

			_, err := influxHelpers.InfluxWrite(influxHelpers.InfluxRetentionPolicyAllTime, influx.Point{
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
			})

			if err != nil {
				log.ErrS(err)
			}
		}()
	}
}
