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
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/logging"
	"github.com/gamedb/website/memcache"
	"github.com/gamedb/website/queue"
	"github.com/go-chi/chi"
)

func PlayerHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	var toasts []Toast

	idx, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Invalid Player ID: " + id, Error: err})
		return
	}

	if !db.IsValidPlayerID(idx) {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Invalid Player ID: " + id})
		return
	}

	// Find the player row
	player, err := db.GetPlayer(idx)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {

			data := errorTemplate{Code: 404, Message: "We haven't scanned this player yet, but we are looking now."}
			data.Toasts = []Toast{{Title: "Player added to scan queue!"}}
			returnErrorTemplate(w, r, data)
		} else {
			returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the player.", Error: err})
		}
		return
	}

	// Redirect to correct slug
	if r.URL.Path != player.GetPath() {
		http.Redirect(w, r, player.GetPath(), 302)
		return
	}

	// Queue profile for a refresh
	// "Profile queued for an update!"
	err = queuePlayer(r, player, db.PlayerUpdateAuto)
	err = helpers.IgnoreErrors(err, db.ErrUpdatingPlayerBot, db.ErrUpdatingPlayerTooSoon)
	logging.Error(err)

	var wg sync.WaitGroup

	// Get friends
	var friends []db.ProfileFriend
	wg.Add(1)
	go func(player db.Player) {

		var err error
		friends, err = player.GetFriends()
		if err != nil {

			logging.Error(err)

		} else {

			// Queue friends to be scanned
			err = player.ShouldUpdate(r, db.PlayerUpdateFriends)
			if err != nil {

				logging.Error(helpers.IgnoreErrors(err, db.ErrUpdatingPlayerBot, db.ErrUpdatingPlayerTooSoon))

			} else {

				for _, friend := range friends {

					f := db.Player{}
					f.PlayerID = friend.SteamID

					err = queuePlayer(r, f, db.PlayerUpdateAuto)
					err = helpers.IgnoreErrors(err, db.ErrUpdatingPlayerBot, db.ErrUpdatingPlayerTooSoon)
					logging.Error(err)
				}

				player.FriendsAddedAt = time.Now()

				err = player.Save() // todo, switch to update query so not to overwrite other player changes
				logging.Error(err)
			}
		}

		wg.Done()

	}(player)

	// Get ranks
	var ranks *db.PlayerRank
	wg.Add(1)
	go func(player db.Player) {

		var err error
		ranks, err = db.GetRank(player.PlayerID)
		if err != nil {
			if err != datastore.ErrNoSuchEntity {
				logging.Error(err)
			}
		}

		wg.Done()

	}(player)

	// Number of players
	var players int
	wg.Add(1)
	go func(player db.Player) {

		var err error
		players, err = db.CountPlayers()
		logging.Error(err)

		wg.Done()
	}(player)

	// Get badges
	var badges []db.ProfileBadge
	wg.Add(1)
	go func(player db.Player) {

		var err error
		badges, err = player.GetBadges()
		logging.Error(err)

		wg.Done()
	}(player)

	// Get recent games
	var recentGames []RecentlyPlayedGame
	wg.Add(1)
	go func(player db.Player) {

		response, err := player.GetRecentGames()
		if err != nil {

			logging.Error(err)

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
					game.Icon = db.DefaultAppIcon
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

		var err error
		bans, err = player.GetBans()
		logging.Error(err)

		wg.Done()
	}(player)

	// Get badge stats
	var badgeStats db.ProfileBadgeStats
	wg.Add(1)
	go func(player db.Player) {

		var err error
		badgeStats, err = player.GetBadgeStats()
		logging.Error(err)

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
	t.Toasts = toasts

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

func queuePlayer(r *http.Request, player db.Player, updateType db.UpdateType) (err error) {

	err = player.ShouldUpdate(r, updateType)
	if err == nil {

		err = queue.Produce(queue.QueueProfiles, []byte(strconv.FormatInt(player.PlayerID, 10)))
		if err == nil {

			err = memcache.SetItem(memcache.PlayerRefreshed(player.PlayerID))
			logging.Error(err)
		}
	}

	return err
}

func PlayerGamesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	playerID := chi.URLParam(r, "id")

	playerIDInt, err := strconv.ParseInt(playerID, 10, 64)
	if err != nil {
		logging.Error(err)
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

			logging.Error(err)

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

				logging.Error(err)

			} else {

				_, err := client.GetAll(ctx, q, &apps)
				logging.Error(err)
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
			logging.Error(err)
		} else {
			total = player.GamesCount
		}

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

func PlayersUpdateAjaxHandler(w http.ResponseWriter, r *http.Request) {

	var response PlayersUpdateResponse

	playerID := chi.URLParam(r, "id")

	idx, err := strconv.ParseInt(playerID, 10, 64)
	if err != nil || !db.IsValidPlayerID(idx) {

		response = PlayersUpdateResponse{Message: "Invalid Player ID", Success: false, Error: err.Error()}
		logging.Error(err)

	} else {

		player, err := db.GetPlayer(idx)
		if err != nil && err != datastore.ErrNoSuchEntity {

			response = PlayersUpdateResponse{Message: "Something has gone wrong", Success: false, Error: err.Error()}
			logging.Error(err)

		} else {

			err := player.ShouldUpdate(r, db.PlayerUpdateManual)
			if err != nil {

				response = PlayersUpdateResponse{Message: "Player is up to date", Success: false}
				logging.Error(err)

			} else {

				// All good
				if err != nil && err == datastore.ErrNoSuchEntity {
					response = PlayersUpdateResponse{Message: "Looking for new player!", Success: true, Error: err.Error()}
				} else {
					response = PlayersUpdateResponse{Message: "Updating player", Success: true}
				}

				err = queuePlayer(r, player, db.PlayerUpdateManual)
				if err != nil {

					response = PlayersUpdateResponse{Message: "Something has gone wrong", Success: false, Error: err.Error()}
					logging.Error(err)

				}
			}
		}
	}

	bytes, err := json.Marshal(response)
	logging.Error(err)
	if err == nil {
		w.Write(bytes)
	}
}

type PlayersUpdateResponse struct {
	Message string `json:"message"` // Browser notification
	Error   string `json:"error"`   // Console log
	Success bool   `json:"success"` // Red or green
}
