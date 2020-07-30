package pages

import (
	"encoding/json"
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/pages/api"
	"github.com/gamedb/gamedb/cmd/webserver/pages/api/generated"
	sessionHelpers "github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
)

func APIRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", apiHandler)
	r.Get("/gamedb", apiGamedbHandler)
	r.Get("/gamedb.json", apiGamedbJSONHandler)
	r.Get("/steam", apiSteamHandler)
	r.Get("/steam.json", apiSteamJSONHandler)

	// Add generated handlers
	generated.HandlerFromMux(api.Server{}, r)

	return r
}

func apiHandler(w http.ResponseWriter, r *http.Request) {

	t := apiTemplate{}
	t.fill(w, r, "API Docs", "A list of API endpoints to access Steam data & Game DB data")
	t.Key = sessionHelpers.Get(r, sessionHelpers.SessionUserAPIKey)

	returnTemplate(w, r, "api", t)
}

type apiTemplate struct {
	globalTemplate
	Key string
}

func apiGamedbHandler(w http.ResponseWriter, r *http.Request) {

	t := apiFrameTemplate{}
	t.JSON = "/api/gamedb.json"

	returnTemplate(w, r, "api_frame", t)
}

func apiSteamHandler(w http.ResponseWriter, r *http.Request) {

	t := apiFrameTemplate{}
	t.JSON = "/api/steam.json"

	returnTemplate(w, r, "api_frame", t)
}

type apiFrameTemplate struct {
	globalTemplate
	JSON string
}

func apiGamedbJSONHandler(w http.ResponseWriter, r *http.Request) {

	b, err := json.Marshal(api.SwaggerGameDB)
	if err != nil {
		log.Err(err, r)
		return
	}

	_, err = w.Write(b)
	log.Err(err, r)
}

func apiSteamJSONHandler(w http.ResponseWriter, r *http.Request) {

	b, err := json.Marshal(api.SwaggerSteam)
	if err != nil {
		log.Err(err, r)
		return
	}

	_, err = w.Write(b)
	log.Err(err, r)
}
