package pages

import (
	"net/http"

	"github.com/Jleagle/session-go/session"
	"github.com/gamedb/gamedb/cmd/webserver/pages/api"
	"github.com/gamedb/gamedb/cmd/webserver/pages/api/generated"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi"
)

func APIRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", apiHandler)
	r.Get("/openapi.json", openAPIJSONHandler)

	server := api.Server{}
	return generated.HandlerFromMux(server, r)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {

	t := apiTemplate{}
	t.fill(w, r, "API", "A list of API endpoints to access Steam data & Game DB data")
	t.Key, _ = session.Get(r, helpers.SessionUserAPIKey)
	t.Swagger = api.Swagger
	t.Base = config.Config.GameDBDomain.Get() + "/api"

	returnTemplate(w, r, "api", t)
}

type apiTemplate struct {
	GlobalTemplate
	Key     string
	Swagger *openapi3.Swagger
	Base    string
}

func (template apiTemplate) InputType(t string, f string) string {

	if t == "array" {
		return template.InputType(f, f)
	}

	switch f {
	case "integer", "int32", "int64":
		return "number"
	case "boolean":
		return "checkbox"
	default:
		return "text"
	}
}

func openAPIJSONHandler(w http.ResponseWriter, r *http.Request) {

	b, err := api.Swagger.MarshalJSON()
	if err != nil {
		log.Err(err)
		return
	}

	_, err = w.Write(b)
	log.Err(err)
}
