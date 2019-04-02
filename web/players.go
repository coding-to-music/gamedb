package web

import (
	"html/template"
	"net/http"
	"strconv"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/mongo"
	"github.com/gamedb/website/sql"
	"github.com/go-chi/chi"
)

func playersRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", playersHandler)
	// r.Post("/", playerIDHandler)
	r.Get("/players.json", playersAjaxHandler)
	r.Get("/{id:[0-9]+}", playerHandler)
	r.Get("/{id:[0-9]+}/games.json", playerGamesAjaxHandler)
	r.Get("/{id:[0-9]+}/update.json", playersUpdateAjaxHandler)
	r.Get("/{id:[0-9]+}/history.json", playersHistoryAjaxHandler)
	r.Get("/{id:[0-9]+}/{slug}", playerHandler)
	return r
}

func playersHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := playersTemplate{}

	//
	var wg sync.WaitGroup

	// Get config
	wg.Add(1)
	go func() {

		defer wg.Done()

		config, err := sql.GetConfig(sql.ConfRanksUpdated)
		log.Err(err, r)

		if err == nil {
			t.Date = config.Value
		}

	}()

	// Count players
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PlayersCount, err = mongo.CountPlayers()
		log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	t.fill(w, r, "Players", "See where you come against the rest of the world ("+template.HTML(humanize.Comma(t.PlayersCount))+" players).")

	err := returnTemplate(w, r, "players", t)
	log.Err(err, r)
}

type playersTemplate struct {
	GlobalTemplate
	PlayersCount int64
	Date         string
}

// func playerIDHandler(w http.ResponseWriter, r *http.Request) {
//
// 	post := r.PostFormValue("id")
// 	post = path.Base(post)
//
// 	id64, err := helpers.GetSteam().GetID(post)
// 	if err != nil {
//
// 		player, err := db.GetPlayerByName(post)
// 		if err != nil || player.PlayerID == 0 {
//
// 			returnErrorTemplate(w, r, errorTemplate{Code: 404, Title: "Can't find user: " + post, Message: "You can use your Steam ID or login to add your profile.", Error: err})
// 			return
// 		}
//
// 		http.Redirect(w, r, helpers.GetPlayerPath(player.PlayerID, player.PersonaName), 302)
// 		return
// 	}
//
// 	http.Redirect(w, r, "/players/"+strconv.FormatInt(id64, 10), 302)
// }

func playersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	//
	var wg sync.WaitGroup

	// Get players
	var playerRows []PlayerRow
	wg.Add(1)
	go func() {

		defer wg.Done()

		columns := map[string]string{
			"3": "level",
			"4": "games_count",
			"5": "badges_count",
			"6": "play_time",
			"7": "friends_count",
		}

		players, err := mongo.GetPlayers(query.getOffset64(), 100, query.getOrderMongo(columns, nil), mongo.M{
			"_id":          1,
			"persona_name": 1,
			"avatar":       1,
			"country_code": 1,
			//
			"badges_count":  1,
			"friends_count": 1,
			"games_count":   1,
			"level":         1,
			"play_time":     1,
			//
			"badges_rank":    1,
			"friends_rank":   1,
			"games_rank":     1,
			"level_rank":     1,
			"play_time_rank": 1,
		})
		if err != nil {
			log.Err(err)
			return
		}

		for _, v := range players {

			playerRow := PlayerRow{}
			playerRow.Player = v

			switch query.getOrderString(columns) {
			case "badges_count":
				playerRow.Rank = v.BadgesRank
			case "friends_count":
				playerRow.Rank = v.FriendsRank
			case "games_count":
				playerRow.Rank = v.GamesRank
			case "level", "":
				playerRow.Rank = v.LevelRank
			case "play_time":
				playerRow.Rank = v.PlayTimeRank
			}

			playerRows = append(playerRows, playerRow)
		}
	}()

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountPlayers()
		log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.FormatInt(total, 10)
	response.RecordsFiltered = response.RecordsTotal // todo, update if we filter players by name
	response.Draw = query.Draw

	for _, v := range playerRows {

		response.AddRow(v.Player.OutputForJSON(v.GetRank()))
	}

	response.output(w, r)
}

type PlayerRow struct {
	Player mongo.Player
	Rank   int
}

func (pr PlayerRow) GetRank() string {

	if pr.Rank == 0 {
		return "-"
	}

	return humanize.Ordinal(pr.Rank)
}
