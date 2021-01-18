package main

//go:generate bash ./scripts/generate.sh

import (
	"compress/flate"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelpers "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/middleware"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
	influx "github.com/influxdata/influxdb1-client"
)

const keyField = "key"

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
	r.Use(chiMiddleware.RedirectSlashes)
	r.Use(chiMiddleware.NewCompressor(flate.DefaultCompression, "text/html", "text/css", "text/javascript", "application/json", "application/javascript").Handler)
	r.Use(middleware.RealIP)
	r.Use(middleware.RateLimiterBlock(time.Second/2, 1, rateLimitedHandler))

	r.Get("/", homeHandler)
	r.Get("/health-check", healthCheckHandler)

	r.NotFound(errorHandler)

	generated.HandlerWithOptions(Server{}, generated.ChiServerOptions{
		BaseRouter: r,
		Middlewares: []generated.MiddlewareFunc{
			authMiddlewear,
		},
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
	)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {

	http.Redirect(w, r, config.C.GameDBDomain+"/api/gamedb", http.StatusTemporaryRedirect)
}

func errorHandler(w http.ResponseWriter, _ *http.Request) {

	w.WriteHeader(404)

	b, err := json.Marshal(generated.MessageResponse{Message: "Invalid endpoint"})
	if err != nil {
		log.ErrS(err)
	}

	_, err = w.Write(b)
	if err != nil {
		log.ErrS(err)
	}
}

func rateLimitedHandler(w http.ResponseWriter, _ *http.Request) {
	returnErrorResponse(w, http.StatusTooManyRequests, errors.New(http.StatusText(http.StatusTooManyRequests)))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
}

func authMiddlewear(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var err error

		// Check API key
		key := r.URL.Query().Get(keyField)
		if key == "" {
			key = strings.TrimLeft(r.Header.Get(keyField), "Bearer ")
			if key == "" {
				key = session.Get(r, session.SessionUserAPIKey)
			}
		}

		if key == "" {
			returnErrorResponse(w, http.StatusUnauthorized, errors.New("empty api key"))
			return
		}

		if !apiKeyRegexp.MatchString(key) {
			returnErrorResponse(w, http.StatusUnauthorized, errors.New("invalid api key: "+key))
			return
		}

		// Check user has access to api
		user, err := mysql.GetUserByAPIKey(key)
		if err == mysql.ErrRecordNotFound {

			returnErrorResponse(w, http.StatusUnauthorized, errors.New("invalid api key: "+key))
			return

		} else if err != nil {

			returnErrorResponse(w, http.StatusInternalServerError, err)
			return

		} else if user.Level < mysql.UserLevel2 {

			returnErrorResponse(w, http.StatusUnauthorized, errors.New("invalid user level"))
			return
		}

		go storeInInflux(map[string]string{
			"path":    r.URL.Path,
			"user_id": strconv.Itoa(user.ID),
		})

		next.ServeHTTP(w, r)
	})
}

func returnErrorResponse(w http.ResponseWriter, code int, err error) {

	returnResponse(w, code, generated.MessageResponse{Message: err.Error()})
}

func returnResponse(w http.ResponseWriter, code int, i interface{}) {

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(i)
	if err != nil {
		log.ErrS(err)
	}

	go storeInInflux(map[string]string{
		"code": strconv.Itoa(code),
	})
}

func storeInInflux(tags map[string]string) {

	if len(tags) > 0 {
		_, err := influxHelpers.InfluxWrite(influxHelpers.InfluxRetentionPolicyAllTime, influx.Point{
			Measurement: string(influxHelpers.InfluxMeasurementAPICalls),
			Tags:        tags,
			Fields: map[string]interface{}{
				"call": 1,
			},
			Time:      time.Now(),
			Precision: "s",
		})

		if err != nil {
			log.ErrS(err)
		}
	}
}
