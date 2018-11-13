package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/logging"
)

const (
	maxPlayers = 4
)

func CoopHandler(w http.ResponseWriter, r *http.Request) {

	// Get player ints
	var playerInts []int64
	for _, v := range r.URL.Query()["p"] {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			logging.Error(err)
		}
		playerInts = append(playerInts, i)
	}

	playerInts = helpers.Unique64(playerInts)

	// Check for max number of players
	if len(playerInts) > maxPlayers {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "You can only compare games from up to " + strconv.Itoa(maxPlayers) + " people."})
		return
	}

	// Get players, one at a time so we can get the missing ones
	var players []db.Player
	var wg sync.WaitGroup
	for _, v := range playerInts {
		wg.Add(1)
		go func(id int64) {

			player, err := db.GetPlayer(id)
			if err != nil {
				if err != db.ErrNoSuchEntity {
					logging.Error(err)
					return
				}
			}

			err = queuePlayer(r, player, player.PlayerID, db.PlayerUpdateManual)
			if err != nil {
				logging.Error(err)
				return
			}

			players = append(players, player)

			wg.Done()
		}(v)
	}
	wg.Wait()

	// Make a map of all games the players have
	var allGames = map[int]int{}
	var gamesSlices [][]int

	for i := 0; i < len(players); i++ {
		player := players[i]

		wg.Add(1)
		go func(player db.Player) {

			var x []int
			resp, err := player.GetAllPlayerApps("app_name", 0)
			if err != nil {
				logging.Error(err)
				return
			}
			for _, vv := range resp {
				allGames[vv.AppID] = vv.AppID
				x = append(x, vv.AppID)
			}
			gamesSlices = append(gamesSlices, x)

			wg.Done()
		}(player)
	}

	// Wait
	wg.Wait()

	// Remove apps that are not in a users apps
	// Loop each game
	for allGameID := range allGames {

		var remove = false

		// Loop each user
		for _, gamesSlice := range gamesSlices {

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
	for _, v := range allGames {
		gamesSlice = append(gamesSlice, v)
	}

	games, err := db.GetAppsByID(gamesSlice, []string{"id", "name", "icon", "platforms", "achievements", "tags"})
	if err != nil {
		logging.Error(err)
	}

	// Make visible tags
	var templateGames []coopGameTemplate
	for _, v := range games {

		coopTags, err := v.GetCoopTags()
		logging.Error(err)

		templateGames = append(templateGames, coopGameTemplate{
			Game: v,
			Tags: coopTags,
		})
	}

	t := coopTemplate{}
	t.Fill(w, r, "Co-op")
	t.Players = players
	t.Games = templateGames
	t.Description = "Find a game to play with friends."

	returnTemplate(w, r, "coop", t)
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
