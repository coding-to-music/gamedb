package pages

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/website/pkg"
	"github.com/go-chi/chi"
)

func newsRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", newsHandler)
	r.Get("/news.json", newsAjaxHandler)
	return r
}

func newsHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour)

	t := newsTemplate{}
	t.fill(w, r, "News", "All the news from all the games, all in one place.")

	apps, err := pkg.PopularApps()
	log.Err(err, r)

	var appIDs []int
	for _, app := range apps {
		appIDs = append(appIDs, app.ID)
	}

	t.Articles, err = pkg.GetArticlesByApps(appIDs, 0, time.Now().AddDate(0, 0, -7))
	log.Err(err, r)

	t.Count, err = pkg.CountDocuments(pkg.CollectionAppArticles, nil)
	log.Err(err, r)

	err = returnTemplate(w, r, "news", t)
	log.Err(err, r)
}

type newsTemplate struct {
	GlobalTemplate
	Articles []pkg.Article
	Count    int64
}

func newsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{"draw", "start"})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*1)

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	var wg sync.WaitGroup

	var articles []pkg.Article
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		articles, err = pkg.GetArticles(query.getOffset64())
		log.Err(err)

		for k, v := range articles {
			articles[k].Contents = pkg.BBCodeCompiler.Compile(v.Contents)
		}
	}()

	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = pkg.CountDocuments(pkg.CollectionAppArticles, nil)
		log.Err(err)
	}()

	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.FormatInt(count, 10)
	response.RecordsFiltered = response.RecordsTotal
	response.Draw = query.Draw

	for _, v := range articles {
		response.AddRow(v.OutputForJSON())
	}

	response.output(w, r)
}
