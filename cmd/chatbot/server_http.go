package main

import (
	"compress/flate"
	"net/http"
	"time"

	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/middleware"
	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
)

func slashCommandServer() error {

	r := chi.NewRouter()
	r.Use(chiMiddleware.RedirectSlashes)
	r.Use(chiMiddleware.NewCompressor(flate.DefaultCompression, "text/html", "text/css", "text/javascript", "application/json", "application/javascript").Handler)
	r.Use(middleware.RealIP)

	r.Get("/health-check", healthCheckHandler)

	for _, c := range chatbot.CommandRegister {

		r.Get("/"+c.ID(), func(w http.ResponseWriter, r *http.Request) {

			_, err := w.Write([]byte("success"))
			if err != nil {
				log.ErrS(err)
			}
		})
	}

	r.NotFound(errorHandler)

	s := &http.Server{
		Addr:              "0.0.0.0:" + config.C.ChatbotPort,
		Handler:           r,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	log.Info("Starting Chatbot webserver on http://" + s.Addr + "/")

	go func() {
		err := s.ListenAndServe() // Blocks
		if err != nil {
			log.ErrS(err)
		}
	}()

	return nil
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")

	_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.ErrS(err)
	}
}

func errorHandler(w http.ResponseWriter, _ *http.Request) {

	w.WriteHeader(404)

	_, err := w.Write([]byte("404"))
	if err != nil {
		log.ErrS(err)
	}
}
