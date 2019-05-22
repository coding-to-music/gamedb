package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
)

func APIRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", apiHandler)
	r.Get("/app", apiApp)
	r.Get("/package", apiPackage)
	r.Get("/bundle", apiBundle)
	r.Get("/group", apiGroup)
	r.Get("/player", apiPlayer)
	return r
}

var (
	apiKey = apiCallParam{Name: "key", Type: "string"}
	start  = apiCallParam{Name: "start", Type: "uint"}
	length = apiCallParam{Name: "length", Type: "uint"}
)

func apiHandler(w http.ResponseWriter, r *http.Request) {

	t := apiTemplate{}
	t.fill(w, r, "API", "")

	t.Calls = []apiCall{
		{
			Title: "New releases",
			Params: []apiCallParam{
				apiKey,
			},
		},
		{
			Title: "Bundles",
			Params: []apiCallParam{
				apiKey,
			},
		},
		{
			Title: "Packages",
			Params: []apiCallParam{
				apiKey,
			},
		},
		{
			Title: "Changes",
			Params: []apiCallParam{
				apiKey,
			},
		},
		{
			Title: "Trending apps",
			Params: []apiCallParam{
				apiKey,
			},
		},
		{
			Title: "News",
			Params: []apiCallParam{
				apiKey,
			},
		},
		{
			Title: "Trending",
			Params: []apiCallParam{
				apiKey,
			},
		},
		{
			Title: "Apps",
			Params: []apiCallParam{
				apiKey,
			},
		},
		{
			Title: "Players",
			Params: []apiCallParam{
				apiKey,
			},
		},
		{
			Title: "Prices",
			Params: []apiCallParam{
				{
					Name: "limit",
					Type: "int",
				},
				{
					Name: "start",
					Type: "int",
				},
			},
		},
		{
			Title: "Product Keys",
			Params: []apiCallParam{
				apiKey,
			},
		},
		{
			Title: "App's top players",
			Params: []apiCallParam{
				apiKey,
			},
		},
		{
			Title: "App's Reviews",
			Params: []apiCallParam{
				apiKey,
			},
		},
		{
			Title: "Player's games",
			Params: []apiCallParam{
				apiKey,
			},
		},
		{
			Title: "Player's history",
			Params: []apiCallParam{
				apiKey,
			},
		},
		{
			Title: "Retrieve an app",
			Path:  "app",
			Params: []apiCallParam{
				apiKey,
				{
					Name: "id",
					Type: "int",
				},
			},
		},
		{
			Title: "Retrieve a package",
			Path:  "package",
			Params: []apiCallParam{
				apiKey,
				{
					Name: "id",
					Type: "int",
				},
			},
		},
	}

	err := returnTemplate(w, r, "api", t)
	log.Err(err, r)
}

type apiTemplate struct {
	GlobalTemplate
	Calls []apiCall
}

type apiCall struct {
	Title  string
	Path   string
	Params []apiCallParam
}

type apiCallParam struct {
	Name string
	Type string
}

func apiApp(w http.ResponseWriter, r *http.Request) {

}

func apiPackage(w http.ResponseWriter, r *http.Request) {

}

func apiBundle(w http.ResponseWriter, r *http.Request) {

}

func apiPlayer(w http.ResponseWriter, r *http.Request) {

}

func apiGroup(w http.ResponseWriter, r *http.Request) {

}
