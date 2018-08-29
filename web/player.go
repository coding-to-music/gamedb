package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/queue"
)

func PlayerHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	idx, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		returnErrorTemplate(w, r, 404, "Invalid Player ID: "+id)
		return
	}

	if !datastore.IsValidPlayerID(idx) {
		returnErrorTemplate(w, r, 404, "Invalid Player ID: "+id)
		return
	}

	player, err := datastore.GetPlayer(idx)
	if err != nil {
		if err != datastore.ErrNoSuchEntity {
			logger.Error(err)
			returnErrorTemplate(w, r, 500, err.Error())
			return
		}
	}

	errs := player.Update(r.UserAgent())
	if len(errs) > 0 {
		for _, v := range errs {
			logger.Error(v)
		}

		// API is probably down, todo
		//for _, v := range errs {
		//	if v.Error() == steami.Steam().ErrInvalidJson {
		//		returnErrorTemplate(w, r, 500, "Couldnt fetch player data, steam API may be down?")
		//		return
		//	}
		//}

		for _, v := range errs {
			returnErrorTemplate(w, r, 500, v.Error())
			return
		}
	}

	// Redirect to correct slug
	if r.URL.Path != player.GetPath() {
		http.Redirect(w, r, player.GetPath(), 302)
		return
	}

	var wg sync.WaitGroup

	// Get friends
	var friends = map[int64]datastore.Player{}
	wg.Add(1)
	go func(player datastore.Player) {

		resp, err := player.GetFriends()
		if err != nil {
			logger.Error(err)
			return
		}

		// Queue friends to be scanned
		if player.ShouldUpdateFriends() {

			for _, v := range resp.Friends {
				p, _ := json.Marshal(queue.RabbitMessageProfile{
					PlayerID: v.SteamID,
					Time:     time.Now(),
				})
				queue.Produce(queue.QueueProfiles, p)
			}

			player.FriendsAddedAt = time.Now()

			err = player.Save()
			logger.Error(err)
		}

		// Make friend ID slice & map
		var friendsSlice []int64
		for _, v := range resp.Friends {
			friendsSlice = append(friendsSlice, v.SteamID)
			friends[v.SteamID] = datastore.Player{PlayerID: v.SteamID}
		}

		// Get friends from DS
		friendsResp, err := datastore.GetPlayersByIDs(friendsSlice)
		logger.Error(err)

		// Fill in the map
		for _, v := range friendsResp {
			if v.PlayerID != 0 {
				friends[v.PlayerID] = v
			}
		}

		wg.Done()

	}(player)

	// Get games
	var games = map[int]*playerAppTemplate{}
	var gameStats = playerAppStatsTemplate{}
	wg.Add(1)
	go func(player datastore.Player) {

		// todo, we should store everything the frontend needs on games field, then no need to query it from mysql below

		resp, err := player.GetGames()
		if err != nil {
			logger.Error(err)
			return
		}

		// Get game info from player
		var gamesSlice []int
		for _, v := range resp {
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
		gamesSql, err := mysql.GetApps(gamesSlice, []string{"id", "name", "price_final", "icon"})
		logger.Error(err)

		for _, v := range gamesSql {

			games[v.ID].ID = v.ID
			games[v.ID].Name = v.GetName()
			games[v.ID].Price = v.GetPriceFinal()
			games[v.ID].Icon = v.GetIcon()

			// Game stats
			gameStats.All.Fill(games[v.ID])
			if games[v.ID].Time > 0 {
				gameStats.Played.Fill(games[v.ID])
			}
		}

		wg.Done()

	}(player)

	// Get ranks
	var ranks *datastore.Rank
	wg.Add(1)
	go func(player datastore.Player) {

		ranks, err = datastore.GetRank(player.PlayerID)
		if err != nil {
			if err != datastore.ErrNoSuchEntity {
				logger.Error(err)
			}
		}

		wg.Done()

	}(player)

	// Number of players
	var players int
	wg.Add(1)
	go func(player datastore.Player) {

		players, err = datastore.CountPlayers()
		logger.Error(err)

		wg.Done()
	}(player)

	// Get badges
	var badges steam.BadgesInfo
	wg.Add(1)
	go func(player datastore.Player) {

		resp, err := player.GetBadges()
		if err != nil {
			logger.Error(err)
			return
		}

		badges = resp.Response

		wg.Done()
	}(player)

	// Get recent games
	var recentGames []steam.RecentlyPlayedGame
	wg.Add(1)
	go func(player datastore.Player) {

		recentGames, err = player.GetRecentGames()
		if err != nil {
			logger.Error(err)
			return
		}

		wg.Done()
	}(player)

	// Get bans
	var bans steam.GetPlayerBanResponse
	wg.Add(1)
	go func(player datastore.Player) {

		bans, err = player.GetBans()
		if err != nil {
			logger.Error(err)
			return
		}

		wg.Done()
	}(player)

	// Wait
	wg.Wait()

	// Template
	t := playerTemplate{}
	t.Fill(w, r, player.PersonaName)
	t.Player = player
	t.Friends = friends
	t.Games = games
	t.Ranks = playerRanksTemplate{*ranks, players}
	t.GameStats = gameStats
	t.Badges = badges
	t.RecentGames = recentGames
	t.Bans = bans

	returnTemplate(w, r, "player", t)
}

type playerTemplate struct {
	GlobalTemplate
	Player      datastore.Player
	Friends     map[int64]datastore.Player
	Games       map[int]*playerAppTemplate
	GameStats   playerAppStatsTemplate
	Ranks       playerRanksTemplate
	Badges      steam.BadgesInfo
	RecentGames []steam.RecentlyPlayedGame
	Bans        steam.GetPlayerBanResponse
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

func (g playerAppTemplate) GetPriceHour() float64 {

	if g.Price == 0 {
		return 0
	}

	if g.Time == 0 {
		return -1
	}

	return g.Price / (float64(g.Time) / 60)
}

func (g playerAppTemplate) GetPriceHourNice() string {

	x := g.GetPriceHour()
	if x == -1 {
		return "∞"
	}

	return strconv.FormatFloat(helpers.DollarsFloat(x), 'f', 2, 64)
}

func (g playerAppTemplate) GetPriceHourSort() string {

	x := g.GetPriceHour()
	if x == -1 {
		return "1000000"
	}

	return strconv.FormatFloat(helpers.DollarsFloat(x), 'f', 2, 64)
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

type playerAppStatsTemplate struct {
	Played playerAppStatsInnerTemplate
	All    playerAppStatsInnerTemplate
}

type playerAppStatsInnerTemplate struct {
	count     int
	price     float64
	priceHour float64
	time      int
}

func (p *playerAppStatsInnerTemplate) Fill(app *playerAppTemplate) {

	p.count++
	p.price = p.price + app.Price
	p.priceHour = p.priceHour + app.GetPriceHour()
	p.time = p.time + app.Time
}

func (p playerAppStatsInnerTemplate) GetAveragePrice() float64 {
	return helpers.DollarsFloat(p.price / float64(p.count))
}

func (p playerAppStatsInnerTemplate) GetTotalPrice() float64 {
	return helpers.DollarsFloat(p.price)
}

func (p playerAppStatsInnerTemplate) GetAveragePriceHour() float64 {
	return helpers.DollarsFloat(p.priceHour / float64(p.count))
}
func (p playerAppStatsInnerTemplate) GetAverageTime() string {
	return helpers.GetTimeShort(int(float64(p.time)/float64(p.count)), 2)
}

func (p playerAppStatsInnerTemplate) GetTotalTime() string {
	return helpers.GetTimeShort(p.time, 2)
}
