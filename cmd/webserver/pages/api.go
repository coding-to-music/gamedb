package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/pages/api"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi"
)

var (
	endpoints = []api.APICall{
		{
			Title: "App - Players",
			Path:  "app-players",
			Params: []api.APICallParam{
				api.ParamAPIKey,
				api.ParamLimit,
				api.ParamPage,
				api.ParamID,
			},
		},
		{
			Title: "App - Price Changes",
			Path:  "app-prices",
			Params: []api.APICallParam{
				api.ParamAPIKey,
				api.ParamPage,
				api.ParamLimit,
				api.ParamID,
			},
		},
		{
			Title: "Apps",
			Path:  "apps",
			Params: []api.APICallParam{
				api.ParamAPIKey,
				api.ParamPage,
				api.ParamLimit,
				api.ParamID,
				api.ParamPlayers,
				api.ParamScore,
				api.ParamCategory,
				api.ParamReleaseDate,
				api.ParamTrending,
			},
			Handler: ApiAppsHandler,
		},
		{
			Title: "Articles",
			Path:  "articles",
			Params: []api.APICallParam{
				api.ParamAPIKey,
				api.ParamPage,
				api.ParamLimit,
			},
		},
		{
			Title: "Bundles",
			Path:  "bundles",
			Params: []api.APICallParam{
				api.ParamAPIKey,
				api.ParamPage,
				api.ParamLimit,
			},
		},
		{
			Title: "Changes",
			Path:  "changes",
			Params: []api.APICallParam{
				api.ParamAPIKey,
				api.ParamPage,
				api.ParamLimit,
			},
		},
		{
			Title: "Groups",
			Path:  "groups",
			Params: []api.APICallParam{
				api.ParamAPIKey,
				api.ParamPage,
				api.ParamLimit,
			},
		},
		{
			Title: "Packages",
			Path:  "packages",
			Params: []api.APICallParam{
				api.ParamAPIKey,
				api.ParamPage,
				api.ParamLimit,
			},
		},
		{
			Title: "Player - Badges",
			Path:  "player-badges",
			Params: []api.APICallParam{
				api.ParamAPIKey,
				api.ParamID,
				api.ParamPage,
				api.ParamLimit,
			},
		},
		{
			Title: "Player - Games",
			Path:  "player-apps",
			Params: []api.APICallParam{
				api.ParamAPIKey,
				api.ParamID,
				api.ParamPage,
				api.ParamLimit,
			},
		},
		{
			Title: "Player - History",
			Path:  "player-history",
			Params: []api.APICallParam{
				api.ParamAPIKey,
				api.ParamID,
				api.ParamPage,
				api.ParamLimit,
			},
		},
		{
			Title:  "Player - Update",
			Path:   "player-update",
			Params: []api.APICallParam{},
		},
		{
			Title: "Players",
			Path:  "players",
			Params: []api.APICallParam{
				api.ParamAPIKey,
				api.ParamPage,
				api.ParamLimit,
			},
		},
		{
			Title: "Stats - Categories",
			Path:  "steam-stats",
			Params: []api.APICallParam{
				api.ParamAPIKey,
			},
		},
		{
			Title: "Stats - Genres",
			Path:  "steam-stats",
			Params: []api.APICallParam{
				api.ParamAPIKey,
			},
		},
		{
			Title: "Stats - Publishers",
			Path:  "steam-stats",
			Params: []api.APICallParam{
				api.ParamAPIKey,
			},
		},
		{
			Title: "Stats - Steam",
			Path:  "steam-stats",
			Params: []api.APICallParam{
				api.ParamAPIKey,
			},
		},
		{
			Title: "Stats - Tags",
			Path:  "steam-stats",
			Params: []api.APICallParam{
				api.ParamAPIKey,
			},
		},
	}
)

func APIRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", apiHandler)
	r.Get("/swagger.json", apiSwaggerHandler)

	for _, v := range endpoints {
		if v.Handler != nil {
			r.Get(v.GetPath(), v.Handler)
		}
	}

	return r
}

func apiHandler(w http.ResponseWriter, r *http.Request) {

	t := apiTemplate{}
	t.fill(w, r, "API", "A list of API endpoints to access Steam data & Game DB data")
	t.Calls = endpoints

	err := returnTemplate(w, r, "api", t)
	log.Err(err, r)
}

type apiTemplate struct {
	GlobalTemplate
	Calls []api.APICall
}

func apiSwaggerHandler(w http.ResponseWriter, r *http.Request) {

	swagger := openapi3.Swagger{
		OpenAPI: "3.0",
		Info: openapi3.Info{
			Title: "Steam DB API",
			Contact: &openapi3.Contact{
				URL: "https://gamedb.online/contact",
			},
		},
		Servers: []*openapi3.Server{
			{URL: "https://gamedb.online/api"},
		},
		Paths: openapi3.Paths{
			"/prefix/{pathArg}/suffix": &openapi3.PathItem{
				Post: &openapi3.Operation{
					// Parameters: openapi3.Parameters{
					// 	{
					// 		Value: &openapi3.Parameter{
					// 			In:     "query",
					// 			Name:   "pathArg",
					// 			Schema: openapi3.NewStringSchema().WithMaxLength(2).NewRef(),
					// 		},
					// 	},
					// 	{
					// 		Value: &openapi3.Parameter{
					// 			In:     "query",
					// 			Name:   "queryArg",
					// 			Schema: openapi3.NewStringSchema().WithMaxLength(2).NewRef(),
					// 		},
					// 	},
					// },
				},
			},
		},
	}

	b, err := swagger.MarshalJSON()
	log.Err(err)

	_, err = w.Write(b)
	log.Err(err)
}

func ApiAppsHandler(w http.ResponseWriter, r *http.Request) {

	call, err := api.NewAPICall(r)

	db, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err)
		return
	}

	db = db.Select([]string{"id", "name", "tags", "genres", "developers", "categories", "prices"})
	db, err = call.SetSQLLimitOffset(db)
	if err != nil {
		log.Err(err)
		return
	}

	var apps []sql.App
	db = db.Find(&apps)
	if db.Error != nil {
		log.Err(db.Error)
		return
	}

	//noinspection GoPreferNilSlice
	var apiApps = []api.ApiApp{}

	for _, v := range apps {
		apiApp := api.ApiApp{}
		err = apiApp.Fill(v)
		log.Err(err)

		apiApps = append(apiApps, apiApp)
	}

	err = returnJSON(w, r, apiApps)
	log.Err(err)
}

func apiPackagesHandler(w http.ResponseWriter, r *http.Request) {

}

func apiBundlesHandler(w http.ResponseWriter, r *http.Request) {

}

func apiPlayersHandler(w http.ResponseWriter, r *http.Request) {

}

func apiGroupsHandler(w http.ResponseWriter, r *http.Request) {

}
