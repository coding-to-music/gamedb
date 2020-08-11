package main

import (
	"compress/flate"
	"fmt"
	"net/http"
	"time"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
)

func main() {

	r := chi.NewRouter()
	r.Use(chiMiddleware.RedirectSlashes)
	r.Use(chiMiddleware.NewCompressor(flate.DefaultCompression, "text/html", "text/css", "text/javascript", "application/json", "application/javascript").Handler)

	generated.HandlerFromMux(Server{}, r)

	// 404
	// r.NotFound(pages.Error404Handler)

	log.Info("Starting API on " + "http://" + config.FrontendPort())

	s := &http.Server{
		Addr:              config.FrontendPort(),
		Handler:           r,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	err := s.ListenAndServe()
	fmt.Println(err)
}
