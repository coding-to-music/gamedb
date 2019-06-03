package pages

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
)

func WishlistsRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", wishlistsHandler)
	r.Get("/apps.json", wishlistAppsHandler)
	r.Get("/tags.json", wishlistTagsHandler)
	return r
}

func wishlistsHandler(w http.ResponseWriter, r *http.Request) {

	t := wishlistsTemplate{}
	t.fill(w, r, "Wishlists", "")

	err := returnTemplate(w, r, "wishlists", t)
	log.Err(err, r)
}

type wishlistsTemplate struct {
	GlobalTemplate
}

func wishlistAppsHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	var wg sync.WaitGroup

	var apps []mongo.WishlistApp
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		apps, err = mongo.GetWishlistApps(query.getOffset64())
		log.Err(err, r)
	}()

	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionWishlistApps, nil)
		log.Err(err, r)
	}()

	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.FormatInt(count, 10)
	response.RecordsFiltered = response.RecordsTotal
	response.Draw = query.Draw

	for _, app := range apps {
		response.AddRow([]interface{}{
			app.AppID,
			app.AppName,
			app.GetAppPath(),
			app.Count,
			app.GetAppIcon(),
		})
	}

	response.output(w, r)
}

func wishlistTagsHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	var wg sync.WaitGroup

	var tags []mongo.WishlistTag
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		tags, err = mongo.GetWishlistTags()
		log.Err(err, r)
	}()

	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionWishlistTags, nil)
		log.Err(err, r)
	}()

	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.FormatInt(count, 10)
	response.RecordsFiltered = response.RecordsTotal
	response.Draw = query.Draw

	for _, tag := range tags {
		response.AddRow([]interface{}{
			tag.TagID,
			tag.TagName,
			tag.GetTagPath(),
			tag.Count,
		})
	}

	response.output(w, r)
}
