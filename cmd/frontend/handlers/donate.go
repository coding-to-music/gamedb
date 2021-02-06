package handlers

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func DonateRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", donateHandler)
	r.Get("/latest.json", latestDonateHandler)
	r.Get("/top.json", topDonateHandler)
	return r
}

func donateHandler(w http.ResponseWriter, r *http.Request) {

	t := donateTemplate{}
	t.fill(w, r, "donate", "Support the Global Steam open source project", `Our database takes up <a href="/stats/gamedb">tonnes</a> of resources. Help pay for the server costs or just buy me a beer 🍻`)
	t.Pages = map[mysql.UserLevel]int{
		mysql.UserLevelGuest: mysql.UserLevelLimitGuest,
		mysql.UserLevelFree:  mysql.UserLevelLimitFree,
		mysql.UserLevel1:     mysql.UserLevelLimit1,
		mysql.UserLevel2:     mysql.UserLevelLimit2,
		mysql.UserLevel3:     mysql.UserLevelLimit3,
	}

	returnTemplate(w, r, t)
}

type donateTemplate struct {
	globalTemplate
	Pages map[mysql.UserLevel]int
}

func topDonateHandler(w http.ResponseWriter, r *http.Request) {

	// Get donators
	donators, err := mysql.TopDonators()
	if err != nil {
		log.ErrS(err)
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
		log.ErrS(err)
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
		log.ErrS(err)
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
		log.ErrS(err)
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
