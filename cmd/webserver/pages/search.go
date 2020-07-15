package pages

import (
	"net/http"

	elastic "github.com/gamedb/gamedb/pkg/elastic-search"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/go-chi/chi"
)

func SearchRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", searchHandler)
	r.Get("/search.json", searchAjaxHandler)
	return r
}

var cats = []string{
	elastic.GlobalTypeAchievement,
	elastic.GlobalTypeApp,
	elastic.GlobalTypeArticle,
	elastic.GlobalTypeGroup,
	elastic.GlobalTypePlayer,
}

func searchHandler(w http.ResponseWriter, r *http.Request) {

	t := searchTemplate{}
	t.fill(w, r, "Search", "Search")
	t.Search = r.URL.Query().Get("s")
	t.Category = r.URL.Query().Get("c")
	t.Categories = []helpers.Tuple{
		{Key: elastic.GlobalTypeAchievement, Value: "Achievements"},
		{Key: elastic.GlobalTypeApp, Value: "Apps"},
		{Key: elastic.GlobalTypeArticle, Value: "Articles"},
		{Key: elastic.GlobalTypeGroup, Value: "Groups"},
		{Key: elastic.GlobalTypePlayer, Value: "Players"},
	}

	if !helpers.SliceHasString(t.Category, cats) {
		t.Category = elastic.GlobalTypePlayer
	}

	returnTemplate(w, r, "search", t)
}

type searchTemplate struct {
	GlobalTemplate
	Search     string
	Category   string
	Categories []helpers.Tuple
}

func searchAjaxHandler(w http.ResponseWriter, r *http.Request) {

	// var search = r.URL.Query().Get("s")
	// var category = r.URL.Query().Get("c")
	//
	// if !helpers.SliceHasString(category, cats) {
	// 	category = elastic.GlobalTypePlayer
	// }
	//
	// items, aggregations, filtered, err := elastic.SearchGlobal(10, 0, search)
	// if err != nil {
	// 	log.Err(err, r)
	// }
	//
	// //
	// var query = datatable.NewDataTableQuery(r, false)
	// var response = datatable.NewDataTablesResponse(r, query, filtered, filtered, aggregations)
	// for _, item := range items {
	//
	// 	response.AddRow([]interface{}{
	// 		item.GetName(), // 0
	// 		item.GetIcon(), // 1
	// 		item.GetPath(), // 2
	// 	})
	// }
	//
	// returnJSON(w, r, response)
}
