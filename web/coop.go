package web

import (
	"net/http"
	"strconv"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
)

const (
	maxPlayers = 4
)

func coopHandler(w http.ResponseWriter, r *http.Request) {

	// Get player ints
	var playerInts []int64
	for _, v := range r.URL.Query()["p"] {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Log(err)
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

			defer wg.Done()

			player, err := db.GetPlayer(id)
			if err != nil {
				if err != datastore.ErrNoSuchEntity {
					log.Log(err)
					return
				}
			}

			err = queuePlayer(r, player, db.PlayerUpdateManual)
			if err != nil {
				log.Log(err)
				return
			}

			players = append(players, player)

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

			defer wg.Done()

			var x []int
			resp, err := player.GetAllPlayerApps("app_name", 0)
			if err != nil {
				log.Log(err)
				return
			}
			for _, vv := range resp {
				allGames[vv.AppID] = vv.AppID
				x = append(x, vv.AppID)
			}
			gamesSlices = append(gamesSlices, x)

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
		log.Log(err)
	}

	// Make visible tags
	var templateGames []coopGameTemplate
	for _, v := range games {

		coopTags, err := v.GetCoopTags()
		log.Log(err)

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

	err = returnTemplate(w, r, "coop", t)
	log.Log(err)
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
