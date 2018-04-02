package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/Jleagle/go-helpers/logger"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/mysql"
)

const (
	maxPlayers = 4
)

func CoopHandler(w http.ResponseWriter, r *http.Request) {

	// Convert to ints
	var playerInts []int
	for _, v := range r.URL.Query()["p"] {
		i, err := strconv.Atoi(v)
		if err != nil {
			logger.Error(err)
		}
		playerInts = append(playerInts, i)
	}

	playerInts = helpers.Unique(playerInts)

	// Check for max number of players
	if len(playerInts) > maxPlayers {
		returnErrorTemplate(w, r, 404, "You can only compare games from up to "+strconv.Itoa(maxPlayers)+" people.")
		return
	}

	// Get playerStrings
	players, err := datastore.GetPlayersByIDs(playerInts)
	if err != nil {
		logger.Error(err)
	}

	// Make a map of all games the players have
	var allGames = map[int]int{}
	var gamesSlice []int
	var gamesSlices [][]int
	var wg sync.WaitGroup

	for i := 0; i < len(players); i++ {
		player := players[i]

		wg.Add(1)
		// todo, use go
		func(player datastore.Player) {
			gamesSlice = []int{}

			for _, vv := range player.GetGames() {
				allGames[vv.AppID] = vv.AppID
				gamesSlice = append(gamesSlice, vv.AppID)
			}

			gamesSlices = append(gamesSlices, gamesSlice)

			wg.Done()
		}(player)
	}
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
	gamesSlice = []int{}
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

	template := coopTemplate{}
	template.Fill(r, "Co-op")
	template.Players = players
	template.Games = templateGames

	returnTemplate(w, r, "coop", template)
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
