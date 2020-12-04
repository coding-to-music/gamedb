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
	r.Get("/steam", apiHandler)
	r.Get("/gamedb.json", apiGamedbJSONHandler)
	r.Get("/gamedb.yaml", apiGamedbYAMLHandler)
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
	t.fill(w, r, "api", "API Docs", "A list of API endpoints to access Steam data & Game DB data")
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

	b, err := json.Marshal(api.SwaggerGameDB)
	if err != nil {
		log.ErrS(err)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.ErrS(err)
	}
}

func apiGamedbYAMLHandler(w http.ResponseWriter, r *http.Request) {

	b, err := yaml.Marshal(api.SwaggerGameDB)
	if err != nil {
		log.ErrS(err)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.ErrS(err)
	}
}

func apiSteamJSONHandler(w http.ResponseWriter, r *http.Request) {

	var b []byte
	var callback = func() (interface{}, error) {
		return json.Marshal(api.GetSteam())
	}

	err := memcache.GetSetInterface(memcache.MemcacheAPISteam, &b, callback)
	if err != nil {
		log.ErrS(err)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.ErrS(err)
	}
}

func apiSteamYAMLHandler(w http.ResponseWriter, r *http.Request) {

	var b []byte
	var callback = func() (interface{}, error) {
		return yaml.Marshal(api.GetSteam())
	}

	err := memcache.GetSetInterface(memcache.MemcacheAPISteam, &b, callback)
	if err != nil {
		log.ErrS(err)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.ErrS(err)
	}
}
