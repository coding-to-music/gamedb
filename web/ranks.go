package web

import (
	"errors"
	"net/http"
	"path"
	"strconv"

	"github.com/Jleagle/steam-go/steam"
	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/steami"
)

const (
	ranksLimit = 100
)

func RanksHandler(w http.ResponseWriter, r *http.Request) {

	// Get page number
	page, err := strconv.Atoi(r.URL.Query().Get("p"))
	if err != nil {
		page = 1
	}

	// Get config
	config, err := mysql.GetConfig(mysql.ConfRanksUpdated)
	logger.Error(err)

	//
	var ranks []datastore.Rank

	switch chi.URLParam(r, "id") {
	case "badges":
		ranks, err = datastore.GetRanksBy("badges_rank", ranksLimit, page)

		for k := range ranks {
			ranks[k].Rank = humanize.Ordinal(ranks[k].BadgesRank)
		}
	case "friends":
		ranks, err = datastore.GetRanksBy("friends_rank", ranksLimit, page)

		for k := range ranks {
			ranks[k].Rank = humanize.Ordinal(ranks[k].FriendsRank)
		}
	case "games":
		ranks, err = datastore.GetRanksBy("games_rank", ranksLimit, page)

		for k := range ranks {
			ranks[k].Rank = humanize.Ordinal(ranks[k].GamesRank)
		}
	case "level", "":
		ranks, err = datastore.GetRanksBy("level_rank", ranksLimit, page)

		for k := range ranks {
			ranks[k].Rank = humanize.Ordinal(ranks[k].LevelRank)
		}
	case "time":
		ranks, err = datastore.GetRanksBy("play_time_rank", ranksLimit, page)

		for k := range ranks {
			ranks[k].Rank = humanize.Ordinal(ranks[k].PlayTimeRank)
		}
	default:
		err = errors.New("incorrect sort")
	}

	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 404, err.Error())
		return
	}

	// Count players
	playersCount, err := datastore.CountPlayers()
	logger.Error(err)

	// Count ranks
	ranksCount, err := datastore.CountRanks()
	logger.Error(err)

	t := playersTemplate{}
	t.Fill(w, r, "Ranks")
	t.Ranks = ranks
	t.PlayersCount = playersCount
	t.RanksCount = ranksCount
	t.Date = config.Value
	t.Pagination = Pagination{
		path:  "/players?p=",
		page:  page,
		limit: ranksLimit,
		total: ranksCount,
	}

	returnTemplate(w, r, "ranks", t)
	return
}

type playersTemplate struct {
	GlobalTemplate
	Ranks        []datastore.Rank
	PlayersCount int
	RanksCount   int
	Pagination   Pagination
	Date         string
}

func PlayerIDHandler(w http.ResponseWriter, r *http.Request) {

	post := r.PostFormValue("id")
	post = path.Base(post)

	// Check datastore
	dbPlayer, err := datastore.GetPlayerByName(post)
	if err != nil {

		if err != datastore.ErrNoSuchEntity {
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
