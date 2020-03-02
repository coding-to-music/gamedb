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
	"github.com/tidwall/pretty"
)

func APIRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", apiHandler)
	r.Get("/openapi.json", openAPIJSONHandler)

	// Add generated handlers
	generated.HandlerFromMux(api.Server{}, r)

	return r
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

func (t apiTemplate) InputType(typex string, f string) string {

	if typex == "array" {
		return t.InputType(f, f)
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

func (t apiTemplate) PathToSchema(path string, verb string) string {

	schema := &openapi3.SchemaRef{}
	x := api.Swagger.Paths[path]

	if verb == "GET" {
		schema = x.Get.Responses["200"].Value.Content["application/json"].Schema
	} else if verb == "POST" {
		schema = x.Post.Responses["200"].Value.Content["application/json"].Schema
	} else {
		return ""
	}

	// return t.renderSchema(schema)
	return string(pretty.Pretty([]byte(t.renderSchema(schema))))
}

func (t apiTemplate) renderSchema(schema *openapi3.SchemaRef) (s string) {

	if len(schema.Value.Properties) > 0 {

		// Object
		s += "{"
		for k, v := range schema.Value.Properties {
			s += "\"" + k + "\":  " + t.renderSchema(v) + ", "
		}
		s += "}"

	} else if schema.Value.Items != nil && schema.Value.Type == "array" {

		// Array
		s += "[" + t.renderSchema(schema.Value.Items) + "]"

	} else {

		// Property
		s += "\"" + schema.Value.Type + "\""
	}

	return s
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
