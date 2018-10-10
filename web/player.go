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
	var friends []db.ProfileFriend
	wg.Add(1)
	go func(player db.Player) {

		friends, err = player.GetFriends()
		logger.Error(err)

		// Queue friends to be scanned
		if player.ShouldUpdateFriends() {

			for _, v := range friends {
				p, err := json.Marshal(queue.RabbitMessageProfile{
					PlayerID: v.SteamID,
					Time:     time.Now(),
				})
				if err != nil {
					logger.Error(err)
				} else {
					queue.Produce(queue.QueueProfiles, p)
				}
			}

			player.FriendsAddedAt = time.Now()

			err = player.Save() // todo, switch to update query so not to overwrite other player changes
			logger.Error(err)
		}

		wg.Done()

	}(player)

	// Get ranks
	var ranks *db.PlayerRank
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
	var badges []db.ProfileBadge
	wg.Add(1)
	go func(player db.Player) {

		badges, err = player.GetBadges()
		logger.Error(err)

		wg.Done()
	}(player)

	// Get recent games
	var recentGames []RecentlyPlayedGame
	wg.Add(1)
	go func(player db.Player) {

		response, err := player.GetRecentGames()
		if err != nil {

			logger.Error(err)

		} else {

			for _, v := range response {

				game := RecentlyPlayedGame{}
				game.AppID = v.AppID
				game.Name = v.Name
				game.Weeks = v.PlayTime2Weeks
				game.WeeksNice = helpers.GetTimeShort(v.PlayTime2Weeks, 2)
				game.AllTime = v.PlayTimeForever
				game.AllTimeNice = helpers.GetTimeShort(v.PlayTimeForever, 2)

				if v.ImgIconURL == "" {
					game.Icon = "/assets/img/no-app-image-square.jpg"
				} else {
					game.Icon = "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(v.AppID) + "/" + v.ImgIconURL + ".jpg"
				}

				recentGames = append(recentGames, game)
			}
		}

		wg.Done()
	}(player)

	// Get bans
	var bans steam.GetPlayerBanResponse
	wg.Add(1)
	go func(player db.Player) {

		bans, err = player.GetBans()
		logger.Error(err)

		wg.Done()
	}(player)

	// Get badge stats
	var badgeStats db.ProfileBadgeStats
	wg.Add(1)
	go func(player db.Player) {

		badgeStats, err = player.GetBadgeStats()
		logger.Error(err)

		wg.Done()
	}(player)

	// Wait
	wg.Wait()

	player.VanintyURL = helpers.TruncateString(player.VanintyURL, 14)

	// Template
	t := playerTemplate{}
	t.Fill(w, r, player.PersonaName)
	t.Player = player
	t.Friends = friends
	t.Apps = []db.PlayerApp{}
	t.Ranks = playerRanksTemplate{*ranks, players}
	t.Badges = badges
	t.BadgeStats = badgeStats
	t.RecentGames = recentGames
	t.Bans = bans

	returnTemplate(w, r, "player", t)
}

type playerTemplate struct {
	GlobalTemplate
	Player      db.Player
	Friends     []db.ProfileFriend
	Apps        []db.PlayerApp
	Ranks       playerRanksTemplate
	Badges      []db.ProfileBadge
	BadgeStats  db.ProfileBadgeStats
	RecentGames []RecentlyPlayedGame
	Bans        steam.GetPlayerBanResponse
}

// RecentlyPlayedGame
type RecentlyPlayedGame struct {
	AppID       int
	Icon        string
	Name        string
	Weeks       int
	WeeksNice   string
	AllTime     int
	AllTimeNice string
}

// playerRanksTemplate
type playerRanksTemplate struct {
	Ranks   db.PlayerRank
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
