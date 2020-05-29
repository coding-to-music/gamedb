package pages

import (
	"net/http"
	"sync"
	"time"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

var upcomingFilter = bson.D{{"release_date_unix", bson.M{"$gte": time.Now().Add(time.Hour * 12 * -1).Unix()}}}

func upcomingRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", upcomingHandler)
	r.Get("/upcoming.json", upcomingAjaxHandler)
	return r
}

func upcomingHandler(w http.ResponseWriter, r *http.Request) {

	t := upcomingTemplate{}
	t.fill(w, r, "Upcoming", "Games with a release date in the future")

	returnTemplate(w, r, "upcoming", t)
}

type upcomingTemplate struct {
	GlobalTemplate
}

func upcomingAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	filter2 := upcomingFilter
	search := query.GetSearchString("search")
	if search != "" {
		filter2 = append(filter2, bson.E{Key: "$text", Value: bson.M{"$search": search}})
	}

	var wg sync.WaitGroup
	var countLock sync.Mutex

	// Count
	var apps []mongo.App
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		columns := map[string]string{
			"1": "group_followers",
			"3": "release_date_unix, group_followers desc",
		}

		projection := bson.M{"_id": 1, "name": 1, "icon": 1, "type": 1, "release_date_unix": 1, "group_id": 1, "group_followers": 1}
		order := query.GetOrderMongo(columns)
		order = append(order, bson.E{Key: "group_followers", Value: -1}, bson.E{Key: "name", Value: 1})
		offset := query.GetOffset64()

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
		count, err = mongo.CountDocuments(mongo.CollectionApps, upcomingFilter, 86400)
		countLock.Unlock()
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Wait()

	//
	var response = datatable.NewDataTablesResponse(r, query, count, filtered)
	for _, app := range apps {

		response.AddRow([]interface{}{
			app.ID,                          // 0
			app.GetName(),                   // 1
			app.GetIcon(),                   // 2
			app.GetPath(),                   // 3
			app.GetType(),                   // 4
			"",                              // 5
			app.GetReleaseDateNice(),        // 6
			app.GetFollowers(),              // 7
			helpers.GetAppStoreLink(app.ID), // 8
			app.ReleaseDateUnix,             // 9
			time.Unix(app.ReleaseDateUnix, 0).Format(helpers.DateYear), // 10
		})
	}

	returnJSON(w, r, response)
}
