package pages

import (
	"net/http"
	"sync"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func appsDLCRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", dlcHandler)
	r.Get("/dlc.json", appsDLCAjaxHandler)
	return r
}

func dlcHandler(w http.ResponseWriter, r *http.Request) {

	t := appsDLCTemplate{}
	t.fill(w, r, "dlc", "DLC", "")

	returnTemplate(w, r, t)
}

type appsDLCTemplate struct {
	globalTemplate
}

func appsDLCAjaxHandler(w http.ResponseWriter, r *http.Request) {

	var wg sync.WaitGroup
	var query = datatable.NewDataTableQuery(r, false)

	var playerApps []mongo.PlayerApp
	var playerAppsOwned = map[int]int{}
	var count int64

	playerID := session.GetPlayerIDFromSesion(r)
	if playerID > 0 {

		filter := bson.D{
			{"player_id", playerID}, // Needed for count
			{"app_dlc_count", bson.M{"$gt": 0}},
		}

		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error
			var columns = map[string]string{
				"0": "app_name",
				"1": "app_dlc_count",
			}

			playerApps, err = mongo.GetPlayerAppsByPlayer(playerID, query.GetOffset64(), 100, query.GetOrderMongo(columns), nil, filter)
			if err != nil {
				log.ErrS(err)
				return
			}

			// Get all DLC for page
			var appIDs []int
			var appIDs2 bson.A
			for _, v := range playerApps {
				appIDs = append(appIDs, v.AppID)
				appIDs2 = append(appIDs2, v.AppID)
			}

			dlcs, err := mongo.GetDLCForApps(appIDs, 0, 0, nil, nil, bson.M{"app_id": 1, "dlc_id": 1})
			if err != nil {
				log.ErrS(err)
				return
			}

			// Get owned DLC
			var dlcsMap = map[int]int{}
			var dlcAppIDs bson.A
			for _, v := range dlcs {
				dlcAppIDs = append(dlcAppIDs, v.DLCID)
				dlcsMap[v.DLCID] = v.AppID
			}

			filter = bson.D{{"app_id", bson.M{"$in": dlcAppIDs}}}

			owned, err := mongo.GetPlayerAppsByPlayer(playerID, 0, 0, nil, bson.M{"_id": 1, "app_id": 1}, filter)
			if err != nil {
				log.ErrS(err)
				return
			}

			for _, v := range owned {
				playerAppsOwned[dlcsMap[v.AppID]]++
			}
		}()

		// Get count
		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error
			count, err = mongo.CountDocuments(mongo.CollectionPlayerApps, filter, 0)
			if err != nil {
				log.ErrS(err)
			}
		}()

		// Wait
		wg.Wait()
	}

	var response = datatable.NewDataTablesResponse(r, query, count, count, nil)
	for _, playerApp := range playerApps {

		response.AddRow([]interface{}{
			playerApp.AppID,                  // 0
			playerApp.AppName,                // 1
			playerApp.GetIcon(),              // 2
			playerApp.GetPath() + "#dlc",     // 3
			playerApp.AppDLCCount,            // 4
			playerAppsOwned[playerApp.AppID], // 5
			playerApp.GetStoreLink(),         // 6
		})
	}

	returnJSON(w, r, response)
}
