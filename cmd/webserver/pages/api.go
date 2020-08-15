package pages

import (
	"encoding/json"
	"net/http"
	"path"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/api"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func APIRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", apiHandler)
	r.Get("/gamedb", apiHandler)
	r.Get("/steam", apiHandler)
	r.Get("/gamedb.json", apiGamedbJSONHandler)
	r.Get("/steam.json", apiSteamJSONHandler)

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
	t.Key = session.Get(r, session.SessionUserAPIKey)
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
		zap.S().Error(err)
		return
	}

	_, err = w.Write(b)
	zap.S().Error(err)
}

func apiSteamJSONHandler(w http.ResponseWriter, r *http.Request) {

	var item = memcache.MemcacheAPISteam
	var b []byte

	err := memcache.GetSetInterface(item.Key, item.Expiration, &b, func() (interface{}, error) {
		return json.Marshal(api.GetSteamJSON())
	})

	if err != nil {
		zap.S().Error(err)
		return
	}

	_, err = w.Write(b)
	zap.S().Error(err)
}
