package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
)

func InfoRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", infoHandler)
	return r
}

func infoHandler(w http.ResponseWriter, r *http.Request) {

	t := GlobalTemplate{}
	t.fill(w, r, "Info", "Game DB Information")

	log.Debug(r.Header.Get("cf-ipcountry"))

	returnTemplate(w, r, "info", t)
}
