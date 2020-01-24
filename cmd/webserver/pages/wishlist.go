package pages

import (
	"math"
	"net/http"
	"sync"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

var wishlistFilter = bson.D{
	{"wishlist_count", bson.M{"$gt": 0}},
	{"wishlist_avg_position", bson.M{"$gt": 0}},
}

func WishlistsRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", wishlistsHandler)
	r.Get("/apps.json", wishlistAppsHandler)
	// r.Get("/tags.json", wishlistTagsHandler)
	return r
}

func wishlistsHandler(w http.ResponseWriter, r *http.Request) {

	t := wishlistsTemplate{}
	t.fill(w, r, "Wishlists", "Steam's most wishlisted games")

	returnTemplate(w, r, "wishlists", t)
}

type wishlistsTemplate struct {
	GlobalTemplate
}

func wishlistAppsHandler(w http.ResponseWriter, r *http.Request) {

	query := newDataTableQuery(r, true)

	filter2 := wishlistFilter
	search := query.getSearchString("search")
	if search != "" {
		filter2 = append(filter2, bson.E{Key: "$text", Value: bson.M{"$search": search}})
	}

	var wg sync.WaitGroup
	var countLock sync.Mutex
	var code = helpers.GetProductCC(r)

	// Count
	var apps []mongo.App
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		columns := map[string]string{
			"1": "wishlist_count",
			"2": "wishlist_avg_position",
			"3": "group_followers",
			"4": "prices." + string(code) + ".final",
			"5": "release_date_unix",
		}

		projection := bson.M{"_id": 1, "name": 1, "icon": 1, "wishlist_count": 1, "wishlist_avg_position": 1, "prices": 1, "group_followers": 1, "group_id": 1, "release_date_unix": 1, "release_state": 1}
		order := query.getOrderMongo(columns)
		offset := query.getOffset64()

		apps, err = mongo.GetApps(offset, 100, order, filter2, projection, nil)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get filtered count
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		countLock.Lock()
		filtered, err = mongo.CountDocuments(mongo.CollectionApps, filter2, 0)
		countLock.Unlock()
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get count
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		countLock.Lock()
		count, err = mongo.CountDocuments(mongo.CollectionApps, wishlistFilter, 86400)
		countLock.Unlock()
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Wait()

	//
	response := DataTablesResponse{}
	response.RecordsTotal = count
	response.RecordsFiltered = filtered
	response.Draw = query.Draw

	for _, app := range apps {

		avgPosition := math.Round(app.WishlistAvgPosition*100) / 100

		response.AddRow([]interface{}{
			app.ID,                          // 0
			app.GetName(),                   // 1
			app.GetIcon(),                   // 2
			app.GetPath(),                   // 3
			app.WishlistCount,               // 4
			avgPosition,                     // 5
			app.GetFollowers(),              // 6
			helpers.GetAppStoreLink(app.ID), // 7
			app.ReleaseDateUnix,             // 8
			app.GetReleaseDateNice(),        // 9
			app.GetPrice(code).GetFinal(),   // 10
		})
	}

	response.output(w, r)
}
