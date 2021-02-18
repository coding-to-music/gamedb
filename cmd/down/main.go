package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func main() {

	err := config.Init(helpers.GetIP())
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
		log.Err("Missing environment variables")
		return
	}

	s := &http.Server{
		Addr:              "0.0.0.0:" + config.C.FrontendPort,
		Handler:           r,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	log.Info("Starting Down on " + "http://" + s.Addr)

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

func handler(w http.ResponseWriter, r *http.Request) {

	b, err := os.ReadFile("down.gohtml")
	if err != nil {
		log.ErrS(zap.Error(err))
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.ErrS(zap.Error(err))
		return
	}
}
