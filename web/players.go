package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	t.fill(w, r, "Players", "See where you come against the rest of the world.")

	//
	var wg sync.WaitGroup

	// Get config
	wg.Add(1)
	go func() {

		defer wg.Done()

		config, err := db.GetConfig(db.ConfRanksUpdated)
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

	// Count ranks
	wg.Add(1)
	go func() {

		defer wg.Done()

		// var err error
		// t.RanksCount, err = db.CountRanks()
		// log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	err := returnTemplate(w, r, "players", t)
	log.Err(err, r)
}

type playersTemplate struct {
	GlobalTemplate
	PlayersCount int64
	RanksCount   int
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

	// Get ranks
	var playersExtra []RankExtra
	wg.Add(1)
	go func() {

		defer wg.Done()

		columns := map[string]string{
			"3": "level_rank",
			"4": "games_rank",
			"5": "badges_rank",
			"6": "play_time_rank",
			"7": "friends_rank",
		}

		ops := options.Find().SetSort(query.getOrderMongo(columns, "level_rank", -1))
		players, err := mongo.GetPlayers(query.getOffset64(), ops)
		if err != nil {
			log.Err(err)
			return
		}

		for _, v := range players {
			playersExtra = append(playersExtra, RankExtra{
				RankRow: mongo.PlayerRank{
					PersonaName: v.PersonaName,
				},
			})
		}

		// _, err := client.GetAll(ctx, q, &ranks)
		// log.Err(err, r)
		//
		// for _, v := range ranks {
		//
		// 	rank := RankExtra{RankRow: v}
		//
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
		//
		// 	ranksExtra = append(ranksExtra, rank)
		// }

	}()

	// Get total
	var total int
	wg.Add(1)
	go func() {

		defer wg.Done()

		// var err error
		// total, err = db.CountRanks()
		// log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(total)
	response.RecordsFiltered = strconv.Itoa(total)
	response.Draw = query.Draw

	for _, v := range playersExtra {

		response.AddRow(v.outputForJSON())
	}

	response.output(w, r)
}

type RankExtra struct {
	RankRow mongo.PlayerRank
	Rank    string
}

// Data array for datatables
func (r *RankExtra) outputForJSON() (output []interface{}) {

	return []interface{}{
		r.Rank,
		strconv.FormatInt(r.RankRow.PlayerID, 10),
		r.RankRow.PersonaName,
		r.RankRow.GetAvatar(),
		r.RankRow.GetAvatar2(),
		r.RankRow.Level,
		r.RankRow.Games,
		r.RankRow.Badges,
		r.RankRow.GetTimeShort(),
		r.RankRow.GetTimeLong(),
		r.RankRow.Friends,
		r.RankRow.GetFlag(),
		r.RankRow.GetCountry(),
	}
}
