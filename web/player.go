package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/queue"
)

func PlayerHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	idx, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		returnErrorTemplate(w, r, 404, "Invalid Player ID: "+id)
		return
	}

	if !db.IsValidPlayerID(idx) {
		returnErrorTemplate(w, r, 404, "Invalid Player ID: "+id)
		return
	}

	// Find the player row
	player, err := db.GetPlayer(idx)
	if err != nil {
		if err != db.ErrNoSuchEntity {
			logger.Error(err)
			returnErrorTemplate(w, r, 500, err.Error())
			return
		}
	}

	// Redirect to correct slug
	if r.URL.Path != player.GetPath() {
		http.Redirect(w, r, player.GetPath(), 302)
		return
	}

	// Update player if needed
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

	var wg sync.WaitGroup

	// Get friends
	var friends = map[int64]db.Player{}
	wg.Add(1)
	go func(player db.Player) {

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
			friends[v.SteamID] = db.Player{PlayerID: v.SteamID}
		}

		// Get friends from DS
		friendsResp, err := db.GetPlayersByIDs(friendsSlice)
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
	//var apps []db.PlayerApp
	//var gameStats = playerAppStatsTemplate{}
	//wg.Add(1)
	//go func(player db.Player) {
	//
	//	apps, err = player.LoadApps()
	//	if err != nil {
	//		logger.Error(err)
	//		return
	//	}
	//
	//	// Make game stats
	//	// todo, do these stats where we save the apps
	//	for _, v := range apps {
	//
	//		gameStats.All.AddApp(v)
	//		if v.AppTime > 0 {
	//			gameStats.Played.AddApp(v)
	//		}
	//	}
	//
	//	wg.Done()
	//
	//}(player)

	// Get ranks
	var ranks *db.Rank
	wg.Add(1)
	go func(player db.Player) {

		ranks, err = db.GetRank(player.PlayerID)
		if err != nil {
			if err != db.ErrNoSuchEntity {
				logger.Error(err)
			}
		}

		wg.Done()

	}(player)

	// Number of players
	var players int
	wg.Add(1)
	go func(player db.Player) {

		players, err = db.CountPlayers()
		logger.Error(err)

		wg.Done()
	}(player)

	// Get badges
	var badges steam.BadgesInfo
	wg.Add(1)
	go func(player db.Player) {

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
	go func(player db.Player) {

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
	go func(player db.Player) {

		bans, err = player.GetBans()
		if err != nil {
			logger.Error(err)
			return
		}

		wg.Done()
	}(player)

	// Wait
	wg.Wait()

	player.VanintyURL = helpers.TruncateString(player.VanintyURL, 15)

	// Template
	t := playerTemplate{}
	t.Fill(w, r, player.PersonaName)
	t.Player = player
	t.Friends = friends
	t.Apps = []db.PlayerApp{}
	t.Ranks = playerRanksTemplate{*ranks, players}
	t.GameStats = playerAppStatsTemplate{}
	t.Badges = badges
	t.RecentGames = recentGames
	t.Bans = bans

	returnTemplate(w, r, "player", t)
}

type playerTemplate struct {
	GlobalTemplate
	Player      db.Player
	Friends     map[int64]db.Player
	Apps        []db.PlayerApp
	GameStats   playerAppStatsTemplate
	Ranks       playerRanksTemplate
	Badges      steam.BadgesInfo
	RecentGames []steam.RecentlyPlayedGame
	Bans        steam.GetPlayerBanResponse
}

// playerRanksTemplate
type playerRanksTemplate struct {
	Ranks   db.Rank
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

// playerAppStatsTemplate
type playerAppStatsTemplate struct {
	Played playerAppStatsInnerTemplate
	All    playerAppStatsInnerTemplate
}

type playerAppStatsInnerTemplate struct {
	count     int
	price     int
	priceHour float64
	time      int
}

func (p *playerAppStatsInnerTemplate) AddApp(app db.PlayerApp) {

	p.count++
	p.price = p.price + app.AppPrice
	p.priceHour = p.priceHour + app.AppPriceHour
	p.time = p.time + app.AppTime
}

func (p playerAppStatsInnerTemplate) GetAveragePrice() float64 {
	return helpers.DollarsFloat(float64(p.price) / float64(p.count))
}

func (p playerAppStatsInnerTemplate) GetTotalPrice() float64 {
	return helpers.DollarsFloat(float64(p.price))
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

func PlayerGamesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	playerID := chi.URLParam(r, "id")

	playerIDInt, err := strconv.ParseInt(playerID, 10, 64)
	if err != nil {
		logger.Error(err)
		return
	}

	query := DataTablesQuery{}
	query.FillFromURL(r.URL.Query())

	//
	var wg sync.WaitGroup

	// Get apps
	var apps []db.PlayerApp

	wg.Add(1)
	go func() {

		client, ctx, err := db.GetDSClient()
		if err != nil {

			logger.Error(err)

		} else {

			columns := map[string]string{
				"0": "app_name",
				"1": "app_price",
				"2": "app_time",
				"3": "app_price_hour",
			}

			q := datastore.NewQuery(db.KindPlayerApp).Filter("player_id =", playerIDInt).Limit(100)
			q, err = query.SetOrderOffsetDS(q, columns)
			if err != nil {

				logger.Error(err)

			} else {

				_, err := client.GetAll(ctx, q, &apps)
				logger.Error(err)
			}
		}

		wg.Done()
	}()

	// Get total
	var total int
	wg.Add(1)
	go func() {

		player, err := db.GetPlayer(playerIDInt)
		if err != nil {
			logger.Error(err)
		}

		total = player.GamesCount

		//q := datastore.NewQuery(db.KindPlayerApp).Filter("player_id =", playerIDInt)
		//total, err = client.Count(ctx, q)

		wg.Done()
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(total)
	response.RecordsFiltered = strconv.Itoa(total)
	response.Draw = query.Draw

	for _, v := range apps {

		response.AddRow([]interface{}{
			v.AppID,
			v.AppName,
			v.GetIcon(),
			v.AppTime,
			v.GetTimeNice(),
			helpers.CentsInt(v.AppPrice),
			v.GetPriceHourFormatted(),
		})
	}

	response.output(w)

}
