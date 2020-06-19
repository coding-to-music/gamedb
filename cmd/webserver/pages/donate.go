package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
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
	t.fill(w, r, "Donate", "Databases take up a tonne of resources. Help pay for the server costs or just buy me a beer.")
	t.Pages = map[int]int{
		sql.UserLevel0: sql.UserLevelLimit0,
		sql.UserLevel1: sql.UserLevelLimit1,
		sql.UserLevel2: sql.UserLevelLimit2,
		sql.UserLevel3: sql.UserLevelLimit3,
		sql.UserLevel4: sql.UserLevelLimit4,
	}

	returnTemplate(w, r, "donate", t)
}

type donateTemplate struct {
	GlobalTemplate
	Pages map[int]int
}

func topDonateHandler(w http.ResponseWriter, r *http.Request) {

	// Get donators
	donators, err := sql.TopDonators()
	if err != nil {
		log.Err(err)
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
		log.Err(err)
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
	donations, err := sql.LatestDonations()
	if err != nil {
		log.Err(err)
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
		log.Err(err)
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
