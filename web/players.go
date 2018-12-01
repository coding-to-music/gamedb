package web

import (
	"net/http"
	"path"
	"strconv"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
)

func playersRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", PlayersHandler)
	r.Post("/", PlayerIDHandler)
	r.Get("/ajax", PlayersAjaxHandler)
	r.Get("/{id:[0-9]+}", PlayerHandler)
	r.Get("/{id:[0-9]+}/ajax/games", PlayerGamesAjaxHandler)
	r.Get("/{id:[0-9]+}/ajax/update", PlayersUpdateAjaxHandler)
	r.Get("/{id:[0-9]+}/{slug}", PlayerHandler)
	return r
}

func PlayersHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := playersTemplate{}
	t.Fill(w, r, "Players")
	t.Description = "See where you come against the rest of the world"

	//
	var wg sync.WaitGroup

	// Get config
	wg.Add(1)
	go func() {

		defer wg.Done()

		config, err := db.GetConfig(db.ConfRanksUpdated)
		log.Log(err)

		t.Date = config.Value

	}()

	// Count players
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PlayersCount, err = db.CountPlayers()
		log.Log(err)

	}()

	// Count ranks
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.RanksCount, err = db.CountRanks()
		log.Log(err)

	}()

	// Wait
	wg.Wait()

	err := returnTemplate(w, r, "ranks", t)
	log.Log(err)
}

type playersTemplate struct {
	GlobalTemplate
	PlayersCount int
	RanksCount   int
	Date         string
}

func PlayerIDHandler(w http.ResponseWriter, r *http.Request) {

	post := r.PostFormValue("id")
	post = path.Base(post)

	// Check datastore
	dbPlayer, err := db.GetPlayerByName(post)
	if err != nil {

		if err != datastore.ErrNoSuchEntity {
			log.Log(err)
		}

		// Check steam
		id, err := helpers.GetSteam().GetID(post)
		if err != nil {

			if err != steam.ErrNoUserFound {
				log.Log(err)
			}

			returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Can't find user: " + post})
			return
		}

		http.Redirect(w, r, "/players/"+strconv.FormatInt(id, 10), 302)
		return
	}

	http.Redirect(w, r, "/players/"+strconv.FormatInt(dbPlayer.PlayerID, 10), 302)
}

func PlayersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	log.Log(err)

	//
	var wg sync.WaitGroup

	// Get ranks
	var ranksExtra []RankExtra

	wg.Add(1)
	go func() {

		defer wg.Done()

		client, ctx, err := db.GetDSClient()
		if err != nil {

			log.Log(err)
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

				log.Log(err)
				return
			}

			//q = q.Filter(column+" >", 0)

			var ranks []db.PlayerRank
			_, err := client.GetAll(ctx, q, &ranks)
			log.Log(err)

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
		log.Log(err)

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

	response.output(w)
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
