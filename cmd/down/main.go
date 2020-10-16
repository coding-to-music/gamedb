package main

import (
	"net/http"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

var (
	version string
	commits string
)

func main() {

	err := config.Init(version, commits, helpers.GetIP())
	log.InitZap(log.LogNameDown)
	defer log.Flush()
	if err != nil {
		log.ErrS(err)
		return
	}

	r := chi.NewRouter()
	r.Get("/", handler)

	// Serve
	if config.C.FrontendPort == "" {
		log.ErrS("Missing environment variables")
		return
	}

	s := &http.Server{
		Addr:              "0.0.0.0:" + config.C.FrontendPort,
		Handler:           r,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	log.Info("Starting Down on " + "http://" + s.Addr)

	err = s.ListenAndServe()
	if err != nil {
		log.ErrS(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Down for maintenance"))
	if err != nil {
		log.ErrS(zap.Error(err))
	}
}
