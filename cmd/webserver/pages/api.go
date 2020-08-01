package pages

import (
	"encoding/json"
	"net/http"
	"path"

	"github.com/gamedb/gamedb/cmd/webserver/api"
	"github.com/gamedb/gamedb/cmd/webserver/api/generated"
	sessionHelpers "github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/go-chi/chi"
)

func APIRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", apiHandler)
	r.Get("/gamedb", apiHandler)
	r.Get("/steam", apiHandler)
	r.Get("/gamedb.json", apiGamedbJSONHandler)
	r.Get("/steam.json", apiSteamJSONHandler)

	// Add generated handlers
	generated.HandlerFromMux(api.Server{}, r)

	return r
}

func apiHandler(w http.ResponseWriter, r *http.Request) {

	var spec = path.Base(r.URL.Path)
	if spec == "api" {
		http.Redirect(w, r, "/api/gamedb", http.StatusTemporaryRedirect)
		return
	}

	t := apiTemplate{}
	t.fill(w, r, "API Docs", "A list of API endpoints to access Steam data & Game DB data")
	t.Key = sessionHelpers.Get(r, sessionHelpers.SessionUserAPIKey)
	t.Spec = spec

	returnTemplate(w, r, "api", t)
}

type apiTemplate struct {
	globalTemplate
	Key  string
	Spec string
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

	var item = memcache.MemcacheAPISteam
	var b []byte

	err := memcache.GetSetInterface(item.Key, item.Expiration, &b, func() (interface{}, error) {
		return json.Marshal(api.GetSteamJSON())
	})

	if err != nil {
		log.Err(err)
		return
	}

	_, err = w.Write(b)
	log.Err(err, r)
}
