package main

import (
	"compress/flate"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
)

var version string
var commits string

func main() {

	config.Init(version, commits, helpers.GetIP())
	log.Initialise(log.LogNameAPI)

	r := chi.NewRouter()
	r.Use(chiMiddleware.RedirectSlashes)
	r.Use(chiMiddleware.NewCompressor(flate.DefaultCompression, "text/html", "text/css", "text/javascript", "application/json", "application/javascript").Handler)
	r.Get("/", home)
	r.NotFound(error404)

	generated.HandlerFromMux(Server{}, r)

	log.Info("Starting API on " + "http://" + config.APIPort())

	s := &http.Server{
		Addr:              config.APIPort(),
		Handler:           r,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	err := s.ListenAndServe()
	log.Critical(err)
}

func home(w http.ResponseWriter, r *http.Request) {

	http.Redirect(w, r, config.Config.GameDBDomain.Get()+"/api/gamedb", http.StatusTemporaryRedirect)
}

func error404(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(404)

	b, err := json.Marshal(generated.MessageResponse{Message: "Invalid endpoint"})
	log.Err(err)

	_, err = w.Write(b)
	log.Err(err)
}
