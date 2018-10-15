package web

import (
	"net/http"
	"path"
	"strconv"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/dustin/go-humanize"
	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logging"
)

const (
	ranksLimit = 100
)

func PlayersHandler(w http.ResponseWriter, r *http.Request) {

	var wg sync.WaitGroup
	var err error

	// Get config
	var config db.Config
	wg.Add(1)
	go func() {

		config, err = db.GetConfig(db.ConfRanksUpdated)
		logging.Error(err)

		wg.Done()
	}()

	// Count players
	var playersCount int
	wg.Add(1)
	go func() {

		playersCount, err = db.CountPlayers()
		logging.Error(err)

		wg.Done()
	}()

	// Count ranks
	var ranksCount int
	wg.Add(1)
	go func() {

		ranksCount, err = db.CountRanks()
		logging.Error(err)

		wg.Done()
	}()

	// Wait
	wg.Wait()

	//
	t := playersTemplate{}
	t.Fill(w, r, "Players")
	t.PlayersCount = playersCount
	t.RanksCount = ranksCount
	t.Date = config.Value

	returnTemplate(w, r, "ranks", t)
	return
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

		if err != db.ErrNoSuchEntity {
			logging.Error(err)
		}

		// Check steam
		id, err := helpers.GetSteam().GetID(post)
		if err != nil {

			if err != steam.ErrNoUserFound {
				logging.Error(err)
			}

			returnErrorTemplate(w, r, 404, "Can't find user: "+post)
			return
		}

		http.Redirect(w, r, "/players/"+strconv.FormatInt(id, 10), 302)
		return
	}

	http.Redirect(w, r, "/players/"+strconv.FormatInt(dbPlayer.PlayerID, 10), 302)
	return
}

func PlayersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	query.FillFromURL(r.URL.Query())

	//
	var wg sync.WaitGroup

	// Get ranks
	var ranksExtra []RankExtra

	wg.Add(1)
	go func() {

		client, ctx, err := db.GetDSClient()
		if err != nil {

			logging.Error(err)

		} else {

			columns := map[string]string{
				"3": "level_rank",
				"4": "games_rank",
				"5": "badges_rank",
				"6": "play_time_rank",
				"7": "friends_rank",
			}

			q := datastore.NewQuery(db.KindPlayerRank).Limit(ranksLimit)

			column := query.GetOrderDS(columns, false)
			if column != "" {
				q, err = query.SetOrderOffsetDS(q, columns)
				if err != nil {

					logging.Error(err)

				} else {

					//q = q.Filter(column+" >", 0)

					var ranks []db.PlayerRank
					_, err := client.GetAll(ctx, q, &ranks)
					logging.Error(err)

					for _, v := range ranks {
						ranksExtra = append(ranksExtra, RankExtra{RankRow: v}.SetRank(column))
					}
				}
			}
		}

		wg.Done()
	}()

	// Get total
	var total int
	var err error
	wg.Add(1)
	go func() {

		total, err = db.CountRanks()
		logging.Error(err)

		wg.Done()
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

func (r RankExtra) SetRank(sort string) RankExtra {
	switch sort {
	case "badges_rank":
		r.Rank = humanize.Ordinal(r.RankRow.BadgesRank)
	case "friends_rank":
		r.Rank = humanize.Ordinal(r.RankRow.FriendsRank)
	case "games_rank":
		r.Rank = humanize.Ordinal(r.RankRow.GamesRank)
	case "level_rank", "":
		r.Rank = humanize.Ordinal(r.RankRow.LevelRank)
	case "play_time_rank":
		r.Rank = humanize.Ordinal(r.RankRow.PlayTimeRank)
	}

	return r
}

// Data array for datatables
func (r RankExtra) outputForJSON() (output []interface{}) {

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
