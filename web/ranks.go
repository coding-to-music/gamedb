package web

import (
	"net/http"
	"path"
	"strconv"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/steami"
)

const (
	ranksLimit = 100
)

type RankExtra struct {
	RankRow db.Rank
	Rank    string
}

func PlayersHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := db.GetConfig(db.ConfRanksUpdated)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 404, err.Error())
		return
	}

	// Count players
	playersCount, err := db.CountPlayers()
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 404, err.Error())
		return
	}

	// Count ranks
	ranksCount, err := db.CountRanks()
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 404, err.Error())
		return
	}

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
			logger.Error(err)
		}

		// Check steam
		id, err := steami.Steam().GetID(post)
		if err != nil {

			if err != steam.ErrNoUserFound {
				logger.Error(err)
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

			logger.Error(err)

		} else {

			columns := map[string]string{
				"3": "level_rank",
				"4": "games_rank",
				"5": "badges_rank",
				"6": "play_time_rank",
				"7": "friends_rank",
			}

			q := datastore.NewQuery(db.KindRank).Limit(ranksLimit)

			sort := query.GetOrderDS(columns)
			if sort != "" {
				query.SetOrderOffsetDS(q, columns)
			}

			q = q.Filter("player_id >", 0)

			q, err = query.SetOrderOffsetDS(q, columns)
			if err != nil {

				logger.Error(err)

			} else {
				var ranks []db.Rank
				_, err := client.GetAll(ctx, q, &ranks)
				logger.Error(err)
			}
		}

		switch chi.URLParam(r, "id") {
		case "badges":
			ranks, err := db.GetRanksBy("badges_rank", ranksLimit, page)
			logger.Error(err)
			for k, v := range ranks {

				ranksExtra = append(ranksExtra, RankExtra{
					RankRow: v,
					Rank:    humanize.Ordinal(ranks[k].BadgesRank),
				})
			}
		case "friends":
			ranks, err := db.GetRanksBy("friends_rank", ranksLimit, page)

			for k := range ranks {
				ranks[k].Rank = humanize.Ordinal(ranks[k].FriendsRank)
			}
		case "games":
			ranks, err := db.GetRanksBy("games_rank", ranksLimit, page)

			for k := range ranks {
				ranks[k].Rank = humanize.Ordinal(ranks[k].GamesRank)
			}
		case "level", "":
			ranks, err := db.GetRanksBy("level_rank", ranksLimit, page)

			for k := range ranks {
				ranks[k].Rank = humanize.Ordinal(ranks[k].LevelRank)
			}
		case "time":
			ranks, err := db.GetRanksBy("play_time_rank", ranksLimit, page)

			for k := range ranks {
				ranks[k].Rank = humanize.Ordinal(ranks[k].PlayTimeRank)
			}
		default:
			err = errors.New("incorrect sort")
		}

		wg.Done()
	}()

	// Get total
	var total int
	var err error
	wg.Add(1)
	go func() {

		total, err = db.CountRanks()
		logger.Error(err)

		wg.Done()
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(total)
	response.RecordsFiltered = strconv.Itoa(total)
	response.Draw = query.Draw

	for _, v := range ranksExtra {

		response.AddRow([]interface{}{
			v.RankRow.PlayerID,
			v.Rank,
			v.RankRow.GetAvatar(),
			v.RankRow.PersonaName,
			v.RankRow.GetAvatar2(),
			v.RankRow.Level,
			v.RankRow.Games,
			v.RankRow.Badges,
			v.RankRow.PlayTime,
			v.RankRow.Friends,
		})
	}

	response.Output(w)
}
