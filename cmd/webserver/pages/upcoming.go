package pages

import (
	"net/http"
	"sync"
	"time"

	"github.com/gamedb/gamedb/cmd/webserver/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

var upcomingFilter = bson.D{{"release_date_unix", bson.M{"$gte": time.Now().AddDate(0, 0, -1).Unix()}}}

func UpcomingRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", upcomingHandler)
	r.Get("/upcoming.json", upcomingAjaxHandler)
	return r
}

func upcomingHandler(w http.ResponseWriter, r *http.Request) {

	t := upcomingTemplate{}
	t.fill(w, r, "Upcoming", "The apps you have to look forward to!")

	var err error
	t.Apps, err = mongo.CountDocuments(mongo.CollectionApps, upcomingFilter, 86400)
	if err != nil {
		log.Err(err, r)
	}

	returnTemplate(w, r, "upcoming", t)
}

type upcomingTemplate struct {
	GlobalTemplate
	Apps int64
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

		// todo
		columns := map[string]string{
			// "1": "group_followers $dir, name ASC",
			// "4": "release_date_unix $dir, group_followers DESC, name ASC",
			"1": "group_followers",
			"4": "release_date_unix",
		}

		projection := bson.M{"_id": 1, "name": 1, "icon": 1, "type": 1, "prices": 1, "release_date_unix": 1, "group_id": 1, "group_followers": 1}
		order := query.GetOrderMongo(columns)
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
	var code = helpers.GetProductCC(r)
	var response = datatable.NewDataTablesResponse(r, query, count, filtered)
	for _, app := range apps {

		response.AddRow([]interface{}{
			app.ID,                          // 0
			app.GetName(),                   // 1
			app.GetIcon(),                   // 2
			app.GetPath(),                   // 3
			app.GetType(),                   // 4
			app.GetPrice(code).GetFinal(),   // 5
			app.GetReleaseDateNice(),        // 6
			app.GetFollowers(),              // 7
			helpers.GetAppStoreLink(app.ID), // 8
			app.ReleaseDateUnix,             // 9
			time.Unix(app.ReleaseDateUnix, 0).Format(helpers.DateYear), // 10
		})
	}

	returnJSON(w, r, response)
}
