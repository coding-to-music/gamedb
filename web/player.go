package web

import (
	"encoding/json"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi"
	slugify "github.com/gosimple/slug"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/queue"
	"github.com/steam-authority/steam-authority/steam"
)

func PlayerHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	slug := chi.URLParam(r, "slug")

	idx, err := strconv.Atoi(id)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 404, err.Error())
		return
	}

	player, err := datastore.GetPlayer(idx)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 404, err.Error())
		return
	}

	errs := player.UpdateIfNeeded()
	if len(errs) > 0 {
		for _, v := range errs {

			logger.Error(err)

			// API is probably down
			if v.Error() == steam.ErrInvalidJson {
				returnErrorTemplate(w, r, 500, "Couldnt fetch player data, steam API may be down?")
				return
			}

			returnErrorTemplate(w, r, 500, err.Error())
			return
		}
	}

	// Redirect to correct slug
	correctSLug := slugify.Make(player.PersonaName)
	if slug != "" && slug != correctSLug {
		http.Redirect(w, r, "/players/"+id+"/"+correctSLug, 302)
		return
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func(player *datastore.Player) {

		// Queue friends
		if player.ShouldUpdateFriends() {

			for _, v := range player.Friends {
				vv, _ := strconv.Atoi(v.SteamID)
				p, _ := json.Marshal(queue.PlayerMessage{
					PlayerID: vv,
				})
				queue.Produce(queue.PlayerQueue, p)
			}

			player.FriendsAddedAt = time.Now()
			player.Save()
		}

		wg.Done()

	}(player)

	var friends []datastore.Player
	wg.Add(1)
	go func(player *datastore.Player) {

		// Make friend ID slice
		var friendsSlice []int
		for _, v := range player.Friends {
			s, _ := strconv.Atoi(v.SteamID)
			friendsSlice = append(friendsSlice, s)
		}

		// Get friends
		friends, err = datastore.GetPlayersByIDs(friendsSlice)
		if err != nil {
			logger.Error(err)
		}

		sort.Slice(friends, func(i, j int) bool {
			return friends[i].Level > friends[j].Level
		})

		wg.Done()

	}(player)

	var games = map[int]*playerAppTemplate{}
	wg.Add(1)
	go func(player *datastore.Player) {

		// todo, we should store everything the frontend needs on games field, then no need to query it from mysql below
		// Get game info from player
		var gamesSlice []int
		for _, v := range player.GetGames() {
			gamesSlice = append(gamesSlice, v.AppID)
			games[v.AppID] = &playerAppTemplate{
				Time:  v.PlaytimeForever,
				Price: 0,
				ID:    v.AppID,
				Name:  v.Name,
				Icon:  "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(v.AppID) + "/" + v.ImgIconURL + ".jpg",
			}
		}

		// Get game info from app
		gamesSql, err := mysql.GetApps(gamesSlice, []string{"id", "name", "price_initial", "icon"})
		if err != nil {
			logger.Error(err)
		}

		for _, v := range gamesSql {
			games[v.ID].ID = v.ID
			games[v.ID].Name = v.GetName()
			games[v.ID].Price = v.GetPriceInitial()
			games[v.ID].Icon = v.GetIcon()
		}

		wg.Done()

	}(player)

	var ranks *datastore.Rank
	wg.Add(1)
	go func(player *datastore.Player) {

		// Get ranks
		ranks, err = datastore.GetRank(player.PlayerID)
		if err != nil {
			if err != datastore.ErrNoSuchEntity {
				logger.Error(err)
			}
		}

		wg.Done()

	}(player)

	var players int
	wg.Add(1)
	go func(player *datastore.Player) {

		// Number of players
		players, err = datastore.CountPlayers()
		if err != nil {
			logger.Error(err)
		}

		wg.Done()
	}(player)

	// Wait
	wg.Wait()

	// Template
	template := playerTemplate{}
	template.Fill(r, player.PersonaName)
	template.Player = player
	template.Friends = friends
	template.Games = games
	template.Ranks = playerRanksTemplate{*ranks, players}

	returnTemplate(w, r, "player", template)
}

type playerTemplate struct {
	GlobalTemplate
	Player  *datastore.Player
	Friends []datastore.Player
	Games   map[int]*playerAppTemplate
	Ranks   playerRanksTemplate
}

type playerAppTemplate struct {
	ID    int
	Name  string
	Price float64
	Icon  string
	Time  int
}

func (g playerAppTemplate) GetTimeNice() string {

	return helpers.GetTimeShort(g.Time, 2)
}

func (g playerAppTemplate) GetPriceHour() string {

	x := g.Price / (float64(g.Time) / 60)
	if math.IsNaN(x) {
		x = 0
	}
	if math.IsInf(x, 0) {
		return "∞"
	}
	return strconv.FormatFloat(helpers.DollarsFloat(x), 'f', 2, 64)
}

func (g playerAppTemplate) GetPriceHourSort() string {

	return strings.Replace(g.GetPriceHour(), "∞", "1000000", 1)
}

// playerRanksTemplate
type playerRanksTemplate struct {
	Ranks   datastore.Rank
	Players int
}

func (p playerRanksTemplate) format(rank int) string {

	ord := humanize.Ordinal(rank)
	if ord == "0th" {
		return "-"
	}
	return ord
}

func (p playerRanksTemplate) GetLevel() string {
	return p.format(p.Ranks.LevelRank)
}

func (p playerRanksTemplate) GetGames() string {
	return p.format(p.Ranks.GamesRank)
}

func (p playerRanksTemplate) GetBadges() string {
	return p.format(p.Ranks.BadgesRank)
}

func (p playerRanksTemplate) GetTime() string {
	return p.format(p.Ranks.PlayTimeRank)
}

func (p playerRanksTemplate) GetFriends() string {
	return p.format(p.Ranks.FriendsRank)
}

func (p playerRanksTemplate) formatPercent(rank int) string {

	if rank == 0 {
		return ""
	}

	precision := 0
	if rank <= 10 {
		precision = 3
	} else if rank <= 100 {
		precision = 2
	} else if rank <= 1000 {
		precision = 1
	}

	percent := (float64(rank) / float64(p.Players)) * 100
	return strconv.FormatFloat(percent, 'f', precision, 64) + "%"

}

func (p playerRanksTemplate) GetLevelPercent() string {
	return p.formatPercent(p.Ranks.LevelRank)
}

func (p playerRanksTemplate) GetGamesPercent() string {
	return p.formatPercent(p.Ranks.GamesRank)
}

func (p playerRanksTemplate) GetBadgesPercent() string {
	return p.formatPercent(p.Ranks.BadgesRank)
}

func (p playerRanksTemplate) GetTimePercent() string {
	return p.formatPercent(p.Ranks.PlayTimeRank)
}

func (p playerRanksTemplate) GetFriendsPercent() string {
	return p.formatPercent(p.Ranks.FriendsRank)
}
