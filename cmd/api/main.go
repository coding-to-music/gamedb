package main

//go:generate bash ./scripts/generate.sh

import (
	"compress/flate"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/middleware"
	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
)

var version string
var commits string

func main() {

	err := config.Init(version, commits, helpers.GetIP())
	log.InitZap(log.LogNameAPI)
	defer log.Flush()
	if err != nil {
		log.ErrS(err)
		return
	}

	session.InitSession()

	r := chi.NewRouter()
	r.Use(chiMiddleware.RedirectSlashes)
	r.Use(chiMiddleware.NewCompressor(flate.DefaultCompression, "text/html", "text/css", "text/javascript", "application/json", "application/javascript").Handler)
	r.Use(middleware.MiddlewareRealIP)
	r.Use(middleware.RateLimiterBlock(time.Second/2, 1, false))

	r.Get("/", homeHandler)
	r.Get("/health-check", healthCheckHandler)

	r.NotFound(errorHandler)

	generated.HandlerFromMux(Server{}, r)

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

	err = s.ListenAndServe()
	if err != nil {
		log.ErrS(err)
	}
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

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")

	_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.ErrS(err)
	}
}
