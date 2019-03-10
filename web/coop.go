package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/queue"
)

const (
	maxPlayers = 8
)

func coopHandler(w http.ResponseWriter, r *http.Request) {

	t := coopTemplate{}
	t.Fill(w, r, "Co-op", "Find a game to play with friends.")

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
	t.Players, err = db.GetPlayersByIDs(playerInts)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Error: err})
		return
	}

	var foundPlayerIDs []int64
	for _, v := range t.Players {
		foundPlayerIDs = append(foundPlayerIDs, v.PlayerID)
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
	var wg sync.WaitGroup
	var allGames = map[int]bool{}
	var allGamesByPlayer [][]int

	for _, player := range t.Players {

		wg.Add(1)
		go func(player db.Player) {

			defer wg.Done()

			playerApps, err := player.GetAppIDs()
			if err != nil {
				log.Err(err, r)
				return
			}

			var playerAppIDs []int
			for _, v := range playerApps {
				allGames[v] = true
				playerAppIDs = append(playerAppIDs, v)
			}
			allGamesByPlayer = append(allGamesByPlayer, playerAppIDs)

		}(player)
	}

	// Wait
	wg.Wait()

	// Remove apps that are not in a users apps
	// Loop each game
	for allGameID := range allGames {

		var remove = false

		// Loop each user
		for _, gamesSlice := range allGamesByPlayer {

			if !helpers.SliceHasInt(gamesSlice, allGameID) {
				remove = true
				break
			}
		}

		if remove {
			delete(allGames, allGameID)
		}
	}

	// Convert to slice
	var gamesSlice []int
	for k := range allGames {
		gamesSlice = append(gamesSlice, k)
	}

	games, err := db.GetAppsByID(gamesSlice, []string{"id", "name", "icon", "platforms", "achievements", "tags"})
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
	Players []db.Player
	Games   []coopGameTemplate
}

type coopGameTemplate struct {
	Game db.App
	Tags string
}
