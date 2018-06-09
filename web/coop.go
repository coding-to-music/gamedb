package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
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
			logger.Error(err)
		}
		playerInts = append(playerInts, i)
	}

	playerInts = helpers.Unique64(playerInts)

	// Check for max number of players
	if len(playerInts) > maxPlayers {
		returnErrorTemplate(w, r, 404, "You can only compare games from up to "+strconv.Itoa(maxPlayers)+" people.")
		return
	}

	// Get players, one at a time so we can get the missing ones
	var players []datastore.Player
	var wg sync.WaitGroup
	for _, v := range playerInts {
		wg.Add(1)
		go func(id int64) {

			player, err := datastore.GetPlayer(id)
			if err != nil {
				if err != datastore.ErrNoSuchEntity {
					logger.Error(err)
					return
				}
			}

			player.Update(r.UserAgent()) // todo, handle errors

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
		go func(player datastore.Player) {

			var x []int
			resp, err := player.GetGames()
			if err != nil {
				logger.Error(err)
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

	games, err := mysql.GetApps(gamesSlice, []string{"id", "name", "icon", "platforms", "achievements", "tags"})
	if err != nil {
		logger.Error(err)
	}

	// Make visible tags
	// todo, just keep in memory?
	coopTags, err := mysql.GetTagsByID(mysql.GetCoopTags())
	if err != nil {
		logger.Error(err)
	}

	var coopTagsInts = map[int]string{}
	for _, v := range coopTags {
		coopTagsInts[v.ID] = v.GetName()
	}

	var templateGames []coopGameTemplate
	for _, v := range games {
		templateGames = append(templateGames, coopGameTemplate{
			Game: v,
			Tags: v.GetCoopTags(coopTagsInts),
		})
	}

	t := coopTemplate{}
	t.Fill(w, r, "Co-op")
	t.Players = players
	t.Games = templateGames

	returnTemplate(w, r, "coop", t)
	return
}

type coopTemplate struct {
	GlobalTemplate
	Players []datastore.Player
	Games   []coopGameTemplate
}

type coopGameTemplate struct {
	Game mysql.App
	Tags string
}
