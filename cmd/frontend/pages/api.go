package pages

import (
	"encoding/json"
	"net/http"
	"path"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/api"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/go-chi/chi"
	"gopkg.in/yaml.v2"
)

func APIRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", apiHandler)
	r.Get("/gamedb", apiHandler)
	r.Get("/gamedb.json", apiGamedbJSONHandler)
	r.Get("/gamedb.yaml", apiGamedbYAMLHandler)
	r.Get("/steam", apiHandler)
	r.Get("/steam.json", apiSteamJSONHandler)
	r.Get("/steam.yaml", apiSteamYAMLHandler)

	return r
}

func apiHandler(w http.ResponseWriter, r *http.Request) {

	var spec = path.Base(r.URL.Path)
	if spec == "api" {
		http.Redirect(w, r, "/api/gamedb", http.StatusTemporaryRedirect)
		return
	}

	t := apiTemplate{}
	t.fill(w, r, "api", "API Docs", "A list of API endpoints to access Steam & Global Steam data")
	t.Key = session.Get(r, session.SessionUserAPIKey)
	t.Spec = spec

	returnTemplate(w, r, t)
}

type apiTemplate struct {
	globalTemplate
	Key  string
	Spec string
}

func apiGamedbJSONHandler(w http.ResponseWriter, r *http.Request) {

	returnJSON(w, r, api.SwaggerGameDB)
}

func apiGamedbYAMLHandler(w http.ResponseWriter, r *http.Request) {

	returnYAML(w, r, api.SwaggerGameDB)
}

func apiSteamJSONHandler(w http.ResponseWriter, r *http.Request) {

	var b []byte
	var callback = func() (interface{}, error) {
		return json.Marshal(api.GetSteam())
	}

	err := memcache.GetSetInterface(memcache.ItemAPISteam, &b, callback)
	if err != nil {
		log.ErrS(err)
		return
	}

	returnJSON(w, r, b)
}

func apiSteamYAMLHandler(w http.ResponseWriter, r *http.Request) {

	var b []byte
	var callback = func() (interface{}, error) {
		return yaml.Marshal(api.GetSteam())
	}

	err := memcache.GetSetInterface(memcache.ItemAPISteam, &b, callback)
	if err != nil {
		log.ErrS(err)
		return
	}

	returnYAML(w, r, b)
}
