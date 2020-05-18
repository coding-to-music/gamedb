package pages

import (
	"encoding/json"
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/pages/api"
	"github.com/gamedb/gamedb/cmd/webserver/pages/api/generated"
	sessionHelpers "github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
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
	t.Swagger = api.Swagger
	t.Base = config.Config.GameDBDomain.Get() + "/api"
	t.Key = sessionHelpers.Get(r, sessionHelpers.SessionUserAPIKey)

	returnTemplate(w, r, "api", t)
}

type apiTemplate struct {
	GlobalTemplate
	Key     string
	Swagger *openapi3.Swagger
	Base    string
}

func (t apiTemplate) InputType(schema *openapi3.Schema) string {

	if schema.Type == "array" {
		return t.InputType(schema.Items.Value)
	}

	switch schema.Type {
	case "integer", "int32", "int64":
		return "number"
	case "boolean":
		return "checkbox"
	default:
		return "text"
	}
}

func (t apiTemplate) ParamType(schema *openapi3.Schema) string {

	switch schema.Type {
	case "integer":
		if schema.Format != "" {
			return schema.Format
		}
		return schema.Type
	case "array":
		return "" + t.ParamType(schema.Items.Value) + " (array)"
	default:
		return schema.Type
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
	return string(pretty.Pretty([]byte(t.renderSchema(schema.Value))))
}

func (t apiTemplate) renderSchema(schema *openapi3.Schema) (s string) {

	if len(schema.Properties) > 0 {

		// Object
		s += "{"
		for k, v := range schema.Properties {
			s += "\"" + k + "\":  " + t.renderSchema(v.Value) + ", "
		}
		s += "}"

	} else if schema.Items != nil && schema.Type == "array" {

		// Array
		s += "[" + t.renderSchema(schema.Items.Value) + "]"

	} else {

		// Property
		s += "\"" + schema.Type + "\""
	}

	return s
}

func openAPIJSONHandler(w http.ResponseWriter, r *http.Request) {

	b, err := json.MarshalIndent(api.Swagger, "", "  ")
	if err != nil {
		log.Err(err, r)
		return
	}

	_, err = w.Write(b)
	log.Err(err, r)
}
