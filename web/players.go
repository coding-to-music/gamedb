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
	var players []mongo.Player
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

		var err error
		players, err = mongo.GetPlayers(query.getOffset64(), 100, query.getOrderMongo(columns, "level", -1), mongo.M{})
		if err != nil {
			log.Err(err)
			return
		}

		// 	switch column {
		// 	case "badges_rank":
		// 		rank.Rank = humanize.Ordinal(rank.RankRow.BadgesRank)
		// 	case "friends_rank":
		// 		rank.Rank = humanize.Ordinal(rank.RankRow.FriendsRank)
		// 	case "games_rank":
		// 		rank.Rank = humanize.Ordinal(rank.RankRow.GamesRank)
		// 	case "level_rank", "":
		// 		rank.Rank = humanize.Ordinal(rank.RankRow.LevelRank)
		// 	case "play_time_rank":
		// 		rank.Rank = humanize.Ordinal(rank.RankRow.PlayTimeRank)
		// 	}

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

	for _, v := range players {

		response.AddRow(v.OutputForJSON())
	}

	response.output(w, r)
}
