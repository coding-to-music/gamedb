package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

func DonateRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", donateHandler)
	r.Get("/top.json", topDonateHandler)
	r.Get("/latest.json", latestDonateHandler)
	return r
}

func donateHandler(w http.ResponseWriter, r *http.Request) {

	t := donateTemplate{}
	t.fill(w, r, "Support the Game DB open source project", "Databases take up a tonne of resources. Help pay for the server costs or just buy me a beer 🍻")
	t.Pages = map[int]int{
		mysql.UserLevel0: mysql.UserLevelLimit0,
		mysql.UserLevel1: mysql.UserLevelLimit1,
		mysql.UserLevel2: mysql.UserLevelLimit2,
		mysql.UserLevel3: mysql.UserLevelLimit3,
		mysql.UserLevel4: mysql.UserLevelLimit4,
	}

	returnTemplate(w, r, "donate", t)
}

type donateTemplate struct {
	globalTemplate
	Pages map[int]int
}

func topDonateHandler(w http.ResponseWriter, r *http.Request) {

	// Get donators
	donators, err := mysql.TopDonators()
	if err != nil {
		zap.S().Error(err)
		return
	}

	// Get players
	var playerIDs []int64
	for _, v := range donators {
		playerIDs = append(playerIDs, v.PlayerID)
	}

	var playersMap = map[int64]mongo.Player{}
	players, err := mongo.GetPlayersByID(playerIDs, bson.M{"_id": 1, "persona_name": 1, "avatar": 1, "country_code": 1})
	if err != nil {
		zap.S().Error(err)
		return
	}

	for _, v := range players {
		playersMap[v.ID] = v
	}

	// Response
	var query = datatable.NewDataTableQuery(r, false)
	var response = datatable.NewDataTablesResponse(r, query, 0, 0, nil)
	for _, donator := range donators {

		var player mongo.Player
		if val, ok := playersMap[donator.PlayerID]; ok {
			player = val
		} else {
			player = mongo.Player{
				PersonaName: "Anonymous",
			}
		}

		// Also update below
		response.AddRow([]interface{}{
			player.ID,          // 0
			player.GetName(),   // 1
			player.GetPath(),   // 2
			player.GetAvatar(), // 3
			player.GetFlag(),   // 4
			donator.Format(),   // 5
		})
	}

	returnJSON(w, r, response)
}

func latestDonateHandler(w http.ResponseWriter, r *http.Request) {

	// Get donations
	donations, err := mysql.LatestDonations()
	if err != nil {
		zap.S().Error(err)
		return
	}

	// Get players
	var playerIDs []int64
	for _, v := range donations {
		playerIDs = append(playerIDs, v.PlayerID)
	}

	var playersMap = map[int64]mongo.Player{}
	players, err := mongo.GetPlayersByID(playerIDs, bson.M{"_id": 1, "persona_name": 1, "avatar": 1, "country_code": 1})
	if err != nil {
		zap.S().Error(err)
		return
	}

	for _, v := range players {
		playersMap[v.ID] = v
	}

	// Response
	var query = datatable.NewDataTableQuery(r, false)
	var response = datatable.NewDataTablesResponse(r, query, 0, 0, nil)
	for _, donation := range donations {

		var player mongo.Player
		if val, ok := playersMap[donation.PlayerID]; ok {
			player = val
		} else {
			player = mongo.Player{
				PersonaName: "Anonymous",
			}
		}

		// Also update above
		response.AddRow([]interface{}{
			player.ID,          // 0
			player.GetName(),   // 1
			player.GetPath(),   // 2
			player.GetAvatar(), // 3
			player.GetFlag(),   // 4
			donation.Format(),  // 5
		})
	}

	returnJSON(w, r, response)
}
