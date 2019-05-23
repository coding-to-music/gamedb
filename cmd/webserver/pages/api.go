package pages

import (
	"net/http"
	"regexp"

	"github.com/gamedb/gamedb/pkg/helpers"
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
	paramAPIKey = apiCallParam{Name: "key", Type: "string"}
	paramStart  = apiCallParam{Name: "start", Type: "int"}
	paramLength = apiCallParam{Name: "length", Type: "int"}
	paramID     = apiCallParam{Name: "length", Type: "int"}
)

func apiHandler(w http.ResponseWriter, r *http.Request) {

	t := apiTemplate{}
	t.fill(w, r, "API", "A list of API endpoints to access Steam data & Game DB data")

	t.Calls = []apiCall{
		{
			Title: "App",
			Path:  "app",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "App - Players",
			Path:  "app-players",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "App - Price Changes",
			Path:  "app-prices",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "App - Reviews",
			Path:  "app-reviews",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "Apps",
			Path:  "apps",
			Params: []apiCallParam{
				paramAPIKey,
			},
		},
		{
			Title: "Apps - New releases",
			Path:  "new-releases",
			Params: []apiCallParam{
				paramAPIKey,
			},
		},
		{
			Title: "Apps - Trending",
			Path:  "trending-apps",
			Params: []apiCallParam{
				paramAPIKey,
			},
		},
		{
			Title: "Apps - Keys",
			Path:  "app-keys",
			Params: []apiCallParam{
				paramAPIKey,
			},
		},
		{
			Title: "Article",
			Path:  "article",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "Articles",
			Path:  "articles",
			Params: []apiCallParam{
				paramAPIKey,
			},
		},
		{
			Title: "Bundle",
			Path:  "bundle",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "Bundles",
			Path:  "bundles",
			Params: []apiCallParam{
				paramAPIKey,
			},
		},
		{
			Title: "Change",
			Path:  "change",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "Changes",
			Path:  "changes",
			Params: []apiCallParam{
				paramAPIKey,
			},
		},
		{
			Title: "Group",
			Path:  "group",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "Groups",
			Path:  "groups",
			Params: []apiCallParam{
				paramAPIKey,
			},
		},
		{
			Title: "Package",
			Path:  "package",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "Packages",
			Path:  "packages",
			Params: []apiCallParam{
				paramAPIKey,
			},
		},
		{
			Title: "Player",
			Path:  "player",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "Player - Games",
			Path:  "player-apps",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "Player - History",
			Path:  "player-history",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "Players",
			Path:  "players",
			Params: []apiCallParam{
				paramAPIKey,
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

func (c apiCall) Hashtag() string {
	return regexp.MustCompile("[^a-zA-Z0-9]+").ReplaceAllString(c.Title, "")
}

type apiCallParam struct {
	Name string
	Type string
}

func (p apiCallParam) InputType() string {
	if helpers.SliceHasString([]string{"int", "uint"}, p.Type) {
		return "number"
	}
	return "text"
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
