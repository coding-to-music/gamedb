package pages

//
// import (
// 	"net/http"
// 	"sync"
//
// 	"github.com/gamedb/gamedb/pkg/helpers"
// 	"github.com/gamedb/gamedb/pkg/log"
// 	"github.com/gamedb/gamedb/pkg/mongo"
// 	"github.com/gamedb/gamedb/pkg/sql"
// 	"github.com/go-chi/chi"
// )
//
// func WishlistsRouter() http.Handler {
// 	r := chi.NewRouter()
// 	r.Get("/", wishlistsHandler)
// 	r.Get("/apps.json", wishlistAppsHandler)
// 	// r.Get("/tags.json", wishlistTagsHandler)
// 	return r
// }
//
// func wishlistsHandler(w http.ResponseWriter, r *http.Request) {
//
// 	t := wishlistsTemplate{}
// 	t.fill(w, r, "Wishlists", "Steam's most wishlisted games")
//
// 	returnTemplate(w, r, "wishlists", t)
// }
//
// type wishlistsTemplate struct {
// 	GlobalTemplate
// }
//
// func wishlistAppsHandler(w http.ResponseWriter, r *http.Request) {
//
// 	query := dataTablesQuery{}
// 	err := query.fillFromURL(r.URL.Query())
// 	log.Err(err, r)
//
// 	var wg sync.WaitGroup
//
// 	var apps []mongo.WishlistApp
// 	wg.Add(1)
// 	go func() {
//
// 		defer wg.Done()
//
// 		var err error
// 		apps, err = mongo.GetWishlistApps(query.getOffset64())
// 		log.Err(err, r)
// 	}()
//
// 	var count int64
// 	wg.Add(1)
// 	go func() {
//
// 		defer wg.Done()
//
// 		var err error
// 		count, err = mongo.CountDocuments(mongo.CollectionWishlistApps, nil, 0)
// 		log.Err(err, r)
// 	}()
//
// 	wg.Wait()
//
// 	response := DataTablesResponse{}
// 	response.RecordsTotal = count
// 	response.RecordsFiltered = count
// 	response.Draw = query.Draw
//
// 	for _, app := range apps {
// 		response.AddRow([]interface{}{
// 			app.AppID,
// 			app.AppName,
// 			app.GetAppPath(),
// 			app.Count,
// 			app.GetAppIcon(),
// 		})
// 	}
//
// 	response.output(w, r)
//
//
//
//
//
// 	query := dataTablesQuery{}
// 	err := query.fillFromURL(r.URL.Query())
// 	log.Err(err, r)
//
// 	//
// 	var code = helpers.GetProductCC(r)
// 	var wg sync.WaitGroup
//
// 	// Get apps
// 	var packages []sql.Package
//
// 	wg.Add(1)
// 	go func(r *http.Request) {
//
// 		defer wg.Done()
//
// 		db, err := sql.GetMySQLClient()
// 		if err != nil {
// 			log.Err(err, r)
// 			return
// 		}
//
// 		db = db.Model(&sql.Package{})
// 		db = db.Select([]string{"id", "name", "apps_count", "change_number_date", "prices", "icon"})
//
// 		sortCols := map[string]string{
// 			"1": "JSON_EXTRACT(prices, \"$." + string(code) + ".final\")",
// 			"2": "JSON_EXTRACT(prices, \"$." + string(code) + ".discount_percent\")",
// 			"3": "apps_count",
// 			"4": "change_number_date",
// 		}
// 		db = query.setOrderOffsetGorm(db, sortCols, "4")
//
// 		db = db.Limit(100)
// 		db = db.Find(&packages)
//
// 		log.Err(db.Error)
//
// 	}(r)
//
// 	// Get total
// 	var count int
// 	wg.Add(1)
// 	go func() {
//
// 		defer wg.Done()
//
// 		var err error
// 		count, err = sql.CountPackages()
// 		log.Err(err, r)
//
// 	}()
//
// 	// Wait
// 	wg.Wait()
//
// 	response := DataTablesResponse{}
// 	response.RecordsTotal = int64(count)
// 	response.RecordsFiltered = int64(count)
// 	response.Draw = query.Draw
//
// 	for _, v := range packages {
// 		response.AddRow(v.OutputForJSON(code))
// 	}
//
// 	response.output(w, r)
// }
//
// func wishlistTagsHandler(w http.ResponseWriter, r *http.Request) {
//
// 	query := dataTablesQuery{}
// 	err := query.fillFromURL(r.URL.Query())
// 	log.Err(err, r)
//
// 	var wg sync.WaitGroup
//
// 	var tags []mongo.WishlistTag
// 	wg.Add(1)
// 	go func() {
//
// 		defer wg.Done()
//
// 		var err error
// 		tags, err = mongo.GetWishlistTags()
// 		log.Err(err, r)
// 	}()
//
// 	var count int64
// 	wg.Add(1)
// 	go func() {
//
// 		defer wg.Done()
//
// 		var err error
// 		count, err = mongo.CountDocuments(mongo.CollectionWishlistTags, nil, 0)
// 		log.Err(err, r)
// 	}()
//
// 	wg.Wait()
//
// 	response := DataTablesResponse{}
// 	response.RecordsTotal = count
// 	response.RecordsFiltered = count
// 	response.Draw = query.Draw
//
// 	for _, tag := range tags {
// 		response.AddRow([]interface{}{
// 			tag.TagID,
// 			tag.TagName,
// 			tag.GetTagPath(),
// 			tag.Count,
// 		})
// 	}
//
// 	response.output(w, r)
// }
