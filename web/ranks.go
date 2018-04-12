package web

import (
	"errors"
	"net/http"
	"path"
	"strconv"

	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/steam"
)

func RanksHandler(w http.ResponseWriter, r *http.Request) {

	// Normalise the order
	var ranks []datastore.Rank
	var err error

	switch chi.URLParam(r, "id") {
	case "badges":
		ranks, err = datastore.GetRanksBy("badges_rank")

		for k := range ranks {
			ranks[k].Rank = humanize.Ordinal(ranks[k].BadgesRank)
		}
	case "friends":
		ranks, err = datastore.GetRanksBy("friends_rank")

		for k := range ranks {
			ranks[k].Rank = humanize.Ordinal(ranks[k].FriendsRank)
		}
	case "games":
		ranks, err = datastore.GetRanksBy("games_rank")

		for k := range ranks {
			ranks[k].Rank = humanize.Ordinal(ranks[k].GamesRank)
		}
	case "level", "":
		ranks, err = datastore.GetRanksBy("level_rank")

		for k := range ranks {
			ranks[k].Rank = humanize.Ordinal(ranks[k].LevelRank)
		}
	case "time":
		ranks, err = datastore.GetRanksBy("play_time_rank")

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
	if err != nil {
		logger.Error(err)
	}

	// Count ranks
	ranksCount, err := datastore.CountRanks()
	if err != nil {
		logger.Error(err)
	}

	template := playersTemplate{}
	template.Fill(r, "Ranks")
	template.Ranks = ranks
	template.PlayersCount = playersCount
	template.RanksCount = ranksCount

	returnTemplate(w, r, "ranks", template)
	return
}

type playersTemplate struct {
	GlobalTemplate
	Ranks        []datastore.Rank
	PlayersCount int
	RanksCount   int
}

func PlayerIDHandler(w http.ResponseWriter, r *http.Request) {

	post := r.PostFormValue("id")
	post = path.Base(post)

	// Check datastore
	dbPlayer, err := datastore.GetPlayerByName(post)
	if err != nil {

		if err != datastore.ErrorNotFound {
			logger.Error(err)
		}

		// Check steam
		id, err := steam.GetID(post)
		if err != nil {

			if err != steam.ErrNoUserFound {
				logger.Error(err)
			}

			returnErrorTemplate(w, r, 404, "Can't find user: "+post)
			return
		}

		http.Redirect(w, r, "/players/"+id, 302)
		return
	}

	http.Redirect(w, r, "/players/"+strconv.Itoa(dbPlayer.PlayerID), 302)
	return
}
