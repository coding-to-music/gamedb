package web

import (
	"net/http"
	"path"
	"strconv"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
)

func playersRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", playersHandler)
	r.Post("/", playerIDHandler)
	r.Get("/ajax", playersAjaxHandler)
	r.Get("/{id:[0-9]+}", playerHandler)
	r.Get("/{id:[0-9]+}/ajax/games", playerGamesAjaxHandler)
	r.Get("/{id:[0-9]+}/ajax/update", playersUpdateAjaxHandler)
	r.Get("/{id:[0-9]+}/ajax/history", playersHistoryAjaxHandler)
	r.Get("/{id:[0-9]+}/{slug}", playerHandler)
	return r
}

func playersHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := playersTemplate{}
	t.Fill(w, r, "Players", "See where you come against the rest of the world.")

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
		t.PlayersCount, err = db.CountPlayers()
		log.Err(err, r)

	}()

	// Count ranks
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.RanksCount, err = db.CountRanks()
		log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	err := returnTemplate(w, r, "ranks", t)
	log.Err(err, r)
}

type playersTemplate struct {
	GlobalTemplate
	PlayersCount int
	RanksCount   int
	Date         string
}

func playerIDHandler(w http.ResponseWriter, r *http.Request) {

	post := r.PostFormValue("id")
	post = path.Base(post)

	id64, err := helpers.GetSteam().GetID(post)
	if err != nil {

		player, err := db.GetPlayerByName(post)
		if err != nil || player.PlayerID == 0 {

			returnErrorTemplate(w, r, errorTemplate{Code: 404, Title: "Can't find user: " + post, Message: "You can use your Steam ID or login to add your profile.", Error: err})
			return
		}

		http.Redirect(w, r, db.GetPlayerPath(player.PlayerID, player.PersonaName), 302)
		return
	}

	http.Redirect(w, r, "/players/"+strconv.FormatInt(id64, 10), 302)
}

func playersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	log.Err(err, r)

	//
	var wg sync.WaitGroup

	// Get ranks
	var ranksExtra []RankExtra

	wg.Add(1)
	go func() {

		defer wg.Done()

		client, ctx, err := db.GetDSClient()
		if err != nil {

			log.Err(err, r)
			return
		}

		columns := map[string]string{
			"3": "level_rank",
			"4": "games_rank",
			"5": "badges_rank",
			"6": "play_time_rank",
			"7": "friends_rank",
		}

		q := datastore.NewQuery(db.KindPlayerRank).Limit(100)

		column := query.GetOrderDS(columns, false)
		if column != "" {
			q, err = query.SetOrderOffsetDS(q, columns)
			if err != nil {

				log.Err(err, r)
				return
			}

			// q = q.Filter(column+" >", 0)

			var ranks []db.PlayerRank
			_, err := client.GetAll(ctx, q, &ranks)
			log.Err(err, r)

			for _, v := range ranks {

				rank := RankExtra{RankRow: v}

				switch column {
				case "badges_rank":
					rank.Rank = humanize.Ordinal(rank.RankRow.BadgesRank)
				case "friends_rank":
					rank.Rank = humanize.Ordinal(rank.RankRow.FriendsRank)
				case "games_rank":
					rank.Rank = humanize.Ordinal(rank.RankRow.GamesRank)
				case "level_rank", "":
					rank.Rank = humanize.Ordinal(rank.RankRow.LevelRank)
				case "play_time_rank":
					rank.Rank = humanize.Ordinal(rank.RankRow.PlayTimeRank)
				}

				ranksExtra = append(ranksExtra, rank)
			}
		}

	}()

	// Get total
	var total int
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = db.CountRanks()
		log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(total)
	response.RecordsFiltered = strconv.Itoa(total)
	response.Draw = query.Draw

	for _, v := range ranksExtra {

		response.AddRow(v.outputForJSON())
	}

	response.output(w, r)
}

type RankExtra struct {
	RankRow db.PlayerRank
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
