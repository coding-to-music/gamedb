package web

import (
	"net/http"
	"strconv"

	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/mongo"
	"github.com/gamedb/website/queue"
	"github.com/gamedb/website/sql"
)

const (
	maxPlayers = 10
)

func coopHandler(w http.ResponseWriter, r *http.Request) {

	t := coopTemplate{}
	t.fill(w, r, "Co-op", "Find a game to play with friends.")

	// Get player ints
	var playerInts []int64
	for _, v := range r.URL.Query()["p"] {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Err(err, r)
		}
		playerInts = append(playerInts, i)
	}

	playerInts = helpers.Unique64(playerInts)

	// Check for max number of players
	if len(playerInts) > maxPlayers {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "You can only compare games from up to " + strconv.Itoa(maxPlayers) + " people."})
		return
	}

	// Get players
	var err error
	t.Players, err = mongo.GetPlayersByIDs(playerInts)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Error: err})
		return
	}

	var foundPlayerIDs []int64
	for _, v := range t.Players {
		foundPlayerIDs = append(foundPlayerIDs, v.ID)
	}

	for _, v := range playerInts {

		// If we couldnt find player
		if !helpers.SliceHasInt64(foundPlayerIDs, v) {

			err = queue.ProducePlayer(v)
			if err != nil {
				log.Err(err, r)
			}

			t.addToast(Toast{Title: "Update", Message: "Player has been queued for an update"})
		}
	}

	// Make a map of all games the players have
	var allApps = map[int]bool{}
	var allAppsByPlayer = map[int64][]int{}

	playerApps, err := mongo.GetPlayerAppsByPlayers(foundPlayerIDs)
	for _, v := range playerApps {

		allApps[v.AppID] = true

		_, ok := allAppsByPlayer[v.PlayerID]
		if ok {

			allAppsByPlayer[v.PlayerID] = append(allAppsByPlayer[v.PlayerID], v.AppID)

		} else {

			allAppsByPlayer[v.PlayerID] = []int{v.AppID}

		}
	}

	// Remove apps that are not in a users apps
	for appID := range allApps {

		var remove = false

		// Loop each user
		for _, gamesSlice := range allAppsByPlayer {

			if !helpers.SliceHasInt(gamesSlice, appID) {
				remove = true
				break
			}
		}

		if remove {
			delete(allApps, appID)
		}
	}

	// Convert to slice
	var appsSlice []int
	for k := range allApps {
		appsSlice = append(appsSlice, k)
	}

	games, err := sql.GetAppsByID(appsSlice, []string{"id", "name", "icon", "platforms", "achievements", "tags"})
	if err != nil {
		log.Err(err, r)
	}

	// Make visible tags
	for _, v := range games {

		coopTags, err := v.GetCoopTags()
		log.Err(err, r)

		t.Games = append(t.Games, coopGameTemplate{
			Game: v,
			Tags: coopTags,
		})
	}

	err = returnTemplate(w, r, "coop", t)
	log.Err(err, r)
}

type coopTemplate struct {
	GlobalTemplate
	Players []mongo.Player
	Games   []coopGameTemplate
}

type coopGameTemplate struct {
	Game sql.App
	Tags string
}
