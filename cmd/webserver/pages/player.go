package pages

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

func PlayerRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", playerHandler)
	r.Get("/games.json", playerGamesAjaxHandler)
	r.Get("/recent.json", playerRecentAjaxHandler)
	r.Get("/friends.json", playerFriendsAjaxHandler)
	r.Get("/badges.json", playerBadgesAjaxHandler)
	r.Get("/update.json", playersUpdateAjaxHandler)
	r.Get("/history.json", playersHistoryAjaxHandler)
	r.Get("/{slug}", playerHandler)
	return r
}

func playerHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	idx, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid Player ID: " + id, Error: err})
		return
	}

	if !helpers.IsValidPlayerID(idx) {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid Player ID: " + id})
		return
	}

	var code = helpers.GetProductCC(r)

	// Find the player row
	player, err := mongo.GetPlayer(idx)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			err = queue.ProducePlayer(idx)
			log.Err(err, r)

			// Template
			tm := playerMissingTemplate{}
			tm.fill(w, r, "Looking for player!", "")
			tm.addAssetHighCharts()
			tm.addToast(Toast{Title: "Update", Message: "Player has been queued for an update"})
			tm.Player = player
			tm.DefaultAvatar = helpers.DefaultPlayerAvatar

			err = returnTemplate(w, r, "player_missing", tm)
			log.Err(err, r)

		} else {
			returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the player.", Error: err})
		}
		return
	}

	//
	var wg sync.WaitGroup

	// Number of players
	var players int64
	wg.Add(1)
	go func(player mongo.Player) {

		defer wg.Done()

		var err error
		players, err = mongo.CountPlayers()
		log.Err(err, r)

	}(player)

	// Get bans
	var bans mongo.PlayerBans
	wg.Add(1)
	go func(player mongo.Player) {

		defer wg.Done()

		var err error
		bans, err = player.GetBans()
		log.Err(err, r)

		if err == nil {
			bans.EconomyBan = strings.Title(bans.EconomyBan)
		}

	}(player)

	// Get badge stats
	var badgeStats mongo.ProfileBadgeStats
	wg.Add(1)
	go func(player mongo.Player) {

		defer wg.Done()

		var err error
		badgeStats, err = player.GetBadgeStats()
		log.Err(err, r)

	}(player)

	// Get groups
	var groups []mongo.Group
	wg.Add(1)
	go func(player mongo.Player) {

		defer wg.Done()

		var err error
		groups, err = mongo.GetGroupsByID(player.Groups, mongo.M{"_id": 1, "name": 1, "members": 1, "icon": 1, "type": 1, "url": 1})
		log.Err(err, r)

		sort.Slice(groups, func(i, j int) bool {
			return groups[i].GetName() > groups[j].GetName()
		})

		for k, v := range groups {
			groups[k].Icon = v.GetIcon()
		}

	}(player)

	// Get wishlist
	var wishlist []playerWishlistItem
	wg.Add(1)
	go func(player mongo.Player) {

		defer wg.Done()

		wishlistApps, err := sql.GetAppsByID(player.Wishlist, []string{"id", "name", "icon", "release_state", "release_date_unix", "prices"})
		log.Err(err)
		if err == nil {
			for _, v := range wishlistApps {

				wishlist = append(wishlist, playerWishlistItem{
					App:   v,
					Price: v.GetPrice(code),
				})
			}
		}

	}(player)

	// Get background app
	var backgroundApp sql.App
	if player.BackgroundAppID > 0 {
		wg.Add(1)
		go func(player mongo.Player) {

			defer wg.Done()

			var err error
			backgroundApp, err = sql.GetApp(player.BackgroundAppID, []string{"id", "name", "background"})
			err = helpers.IgnoreErrors(err, sql.ErrInvalidAppID)
			log.Err(err)

		}(player)
	}

	// Wait
	wg.Wait()

	player.VanintyURL = helpers.TruncateString(player.VanintyURL, 14, "...")

	// Game stats
	gameStats, err := player.GetGameStats(code)
	log.Err(err, r)

	// Make banners
	banners := make(map[string][]string)
	var primary []string

	if player.ID == 76561197960287930 {
		primary = append(primary, "This profile belongs to Gabe Newell, the Co-founder of Valve")
	}

	if len(primary) > 0 {
		banners["primary"] = primary
	}

	t := playerTemplate{}

	if player.NeedsUpdate(mongo.PlayerUpdateAuto) && !helpers.IsBot(r.UserAgent()) {

		err = queue.ProducePlayer(player.ID)

		if err != nil {
			log.Err(err, r)
		} else {
			t.addToast(Toast{Title: "Update", Message: "Player has been queued for an update"})
		}
	}

	// Template
	t.setBackground(backgroundApp, true, false)
	t.fill(w, r, player.PersonaName, "")
	t.addAssetHighCharts()

	t.Apps = []mongo.PlayerApp{}
	t.Badges = player.GetSpecialBadges()
	t.BadgeStats = badgeStats
	t.Banners = banners
	t.Bans = bans
	t.Canonical = player.GetPath()
	t.DefaultAvatar = helpers.DefaultPlayerAvatar
	t.GameStats = gameStats
	t.Groups = groups
	t.Player = player
	t.Wishlist = wishlist

	err = returnTemplate(w, r, "player", t)
	log.Err(err, r)
}

type playerTemplate struct {
	GlobalTemplate
	Apps          []mongo.PlayerApp
	Badges        []mongo.PlayerBadge
	BadgeStats    mongo.ProfileBadgeStats
	Banners       map[string][]string
	Bans          mongo.PlayerBans
	DefaultAvatar string
	GameStats     mongo.PlayerAppStatsTemplate
	Groups        []mongo.Group
	Player        mongo.Player
	Wishlist      []playerWishlistItem
	// Ranks       playerRanksTemplate
}

type playerMissingTemplate struct {
	GlobalTemplate
	Player        mongo.Player
	DefaultAvatar string
}

type playerWishlistItem struct {
	App   sql.App
	Price sql.ProductPrice
}

// // playerRanksTemplate
// type playerRanksTemplate struct {
// 	Ranks   mongo.PlayerRank
// 	Players int64
// }
//
// // func (p playerRanksTemplate) format(rank int) string {
// //
// // 	ord := humanize.Ordinal(rank)
// // 	if ord == "0th" {
// // 		return "-"
// // 	}
// // 	return ord
// // }
// //
// // func (p playerRanksTemplate) GetLevel() string {
// // 	return p.format(p.Ranks.LevelRank)
// // }
// //
// // func (p playerRanksTemplate) GetGames() string {
// // 	return p.format(p.Ranks.GamesRank)
// // }
// //
// // func (p playerRanksTemplate) GetBadges() string {
// // 	return p.format(p.Ranks.BadgesRank)
// // }
// //
// // func (p playerRanksTemplate) GetTime() string {
// // 	return p.format(p.Ranks.PlayTimeRank)
// // }
// //
// // func (p playerRanksTemplate) GetFriends() string {
// // 	return p.format(p.Ranks.FriendsRank)
// // }
// //
// // func (p playerRanksTemplate) formatPercent(rank int) string {
// //
// // 	if rank == 0 {
// // 		return ""
// // 	}
// //
// // 	precision := 0
// // 	if rank <= 10 {
// // 		precision = 3
// // 	} else if rank <= 100 {
// // 		precision = 2
// // 	} else if rank <= 1000 {
// // 		precision = 1
// // 	}
// //
// // 	percent := (float64(rank) / float64(p.Players)) * 100
// // 	return helpers.FloatToString(percent, precision) + "%"
// //
// // }
// //
// // func (p playerRanksTemplate) GetLevelPercent() string {
// // 	return p.formatPercent(p.Ranks.LevelRank)
// // }
// //
// // func (p playerRanksTemplate) GetGamesPercent() string {
// // 	return p.formatPercent(p.Ranks.GamesRank)
// // }
// //
// // func (p playerRanksTemplate) GetBadgesPercent() string {
// // 	return p.formatPercent(p.Ranks.BadgesRank)
// // }
// //
// // func (p playerRanksTemplate) GetTimePercent() string {
// // 	return p.formatPercent(p.Ranks.PlayTimeRank)
// // }
// //
// // func (p playerRanksTemplate) GetFriendsPercent() string {
// // 	return p.formatPercent(p.Ranks.FriendsRank)
// // }

func playerGamesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	playerID := chi.URLParam(r, "id")

	playerIDInt, err := strconv.ParseInt(playerID, 10, 64)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	query.limit(r)

	code := helpers.GetProductCC(r)

	//
	var wg sync.WaitGroup

	// Get apps
	var playerApps []mongo.PlayerApp
	wg.Add(1)
	go func() {

		defer wg.Done()

		columns := map[string]string{
			"0": "app_name",
			"1": "app_prices",
			"2": "app_time",
			"3": "app_prices_hour",
		}

		colEdit := func(col string) string {
			if col == "app_prices" || col == "app_prices_hour" {
				col = col + "." + string(code)
			}
			return col
		}

		var err error
		playerApps, err = mongo.GetPlayerApps(playerIDInt, query.getOffset64(), 100, query.getOrderMongo(columns, colEdit))
		log.Err(err)
	}()

	// Get total
	var total int
	wg.Add(1)
	go func() {

		defer wg.Done()

		player, err := mongo.GetPlayer(playerIDInt)
		if err != nil {
			log.Err(err, r)
			return
		}

		total = player.GamesCount

	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = int64(total)
	response.RecordsFiltered = int64(total)
	response.Draw = query.Draw
	response.limit(r)

	for _, pa := range playerApps {
		response.AddRow([]interface{}{
			pa.AppID,
			pa.AppName,
			pa.GetIcon(),
			pa.AppTime,
			pa.GetTimeNice(),
			pa.GetPriceFormatted(code),
			pa.GetPriceHourFormatted(code),
			pa.GetPath(),
		})
	}

	response.output(w, r)

}

func playerRecentAjaxHandler(w http.ResponseWriter, r *http.Request) {

	playerID := chi.URLParam(r, "id")

	playerIDInt, err := strconv.ParseInt(playerID, 10, 64)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	query.limit(r)

	//
	var wg sync.WaitGroup

	// Get apps
	var apps []mongo.PlayerRecentApp
	wg.Add(1)
	go func() {

		defer wg.Done()

		columns := map[string]string{
			"0": "name",
			"1": "playtime_2_weeks",
			"2": "playtime_forever",
		}

		var err error
		apps, err = mongo.GetRecentApps(playerIDInt, query.getOffset64(), 100, query.getOrderMongo(columns, nil))
		log.Err(err)
	}()

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountRecent(playerIDInt)
		if err != nil {
			log.Err(err, r)
			return
		}
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = total
	response.RecordsFiltered = total
	response.Draw = query.Draw
	response.limit(r)

	for _, app := range apps {
		response.AddRow([]interface{}{
			app.AppID,                               // 0
			helpers.GetAppIcon(app.AppID, app.Icon), // 1
			app.AppName,                             // 2
			helpers.GetTimeShort(app.PlayTime2Weeks, 2),  // 3
			helpers.GetTimeShort(app.PlayTimeForever, 2), // 4
			helpers.GetAppPath(app.AppID, app.AppName),   // 5
		})
	}

	response.output(w, r)

}

func playerFriendsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	playerID := chi.URLParam(r, "id")

	playerIDInt, err := strconv.ParseInt(playerID, 10, 64)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	query.limit(r)

	//
	var wg sync.WaitGroup

	// Get apps
	var friends []mongo.PlayerFriend
	wg.Add(1)
	go func() {

		defer wg.Done()

		columns := map[string]string{
			"1": "level",
			"2": "games",
			"3": "logged_off",
			"4": "since",
		}

		var err error
		friends, err = mongo.GetFriends(playerIDInt, query.getOffset64(), 100, query.getOrderMongo(columns, nil))
		log.Err(err)
	}()

	// Get total
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountFriends(playerIDInt)
		if err != nil {
			log.Err(err, r)
			return
		}
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = count
	response.RecordsFiltered = count
	response.Draw = query.Draw
	response.limit(r)

	for _, friend := range friends {
		response.AddRow([]interface{}{
			strconv.FormatInt(friend.PlayerID, 10), // 0
			friend.GetPath(),                       // 1
			friend.Avatar,                          // 2
			friend.GetName(),                       // 3
			friend.GetLevel(),                      // 4
			friend.Scanned(),                       // 5
			friend.Games,                           // 6
			friend.GetLoggedOff(),                  // 7
			friend.GetFriendSince(),                // 8
		})
	}

	response.output(w, r)
}

func playerBadgesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id: "+id, r)
		return
	}

	idx, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	if err != nil {
		log.Err(err)
		return
	}

	// Make filter
	var filter = mongo.M{"app_id": mongo.M{"$gt": 0}, "player_id": idx}

	//
	var wg sync.WaitGroup

	// Get badges
	var badges []mongo.PlayerBadge
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		badges, err = mongo.GetPlayerEventBadges(query.getOffset64(), filter)
		if err != nil {
			log.Err(err, r)
		}
	}(r)

	// Get total
	var total int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionPlayerBadges, filter, 0)
		if err != nil {
			log.Err(err, r)
		}
	}(r)

	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = total
	response.RecordsFiltered = total
	response.Draw = query.Draw

	for _, badge := range badges {
		response.AddRow([]interface{}{
			badge.AppID,        // 0
			badge.AppName,      // 1
			badge.GetAppPath(), // 2
			badge.BadgeCompletionTime.Format("2006-01-02 15:04:05"), // 3
			badge.BadgeFoil,     // 4
			badge.BadgeIcon,     // 5
			badge.BadgeLevel,    // 6
			badge.BadgeScarcity, // 7
			badge.BadgeXP,       // 8
		})
	}

	response.output(w, r)
}

func playersUpdateAjaxHandler(w http.ResponseWriter, r *http.Request) {

	message, err, success := func(r *http.Request) (string, error, bool) {

		// todo, csrf

		idx, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			return "Invalid Player ID", err, false
		}

		if !helpers.IsValidPlayerID(idx) {
			return "Invalid Player ID", err, false
		}

		var message string

		player, err := mongo.GetPlayer(idx)
		if err == nil {
			message = "Updating player!"
		} else if err == mongo.ErrNoDocuments {
			message = "Looking for new player!"
		} else {
			log.Err(err, r)
			return "Error looking for player", err, false
		}

		updateType := mongo.PlayerUpdateManual
		if isAdmin(r) {
			message = "Admin update!"
			updateType = mongo.PlayerUpdateAdmin
		}

		if !player.NeedsUpdate(updateType) {
			return "Player can't be updated yet", nil, false
		}

		err = queue.ProducePlayer(player.ID)
		if err != nil {
			log.Err(err, r)
			return "Something has gone wrong", err, false
		}

		return message, err, true
	}(r)

	var response = PlayersUpdateResponse{
		Success: success,
		Toast:   message,
		Log:     err,
	}

	err = returnJSON(w, r, response)
	log.Err(err, r)
}

type PlayersUpdateResponse struct {
	Success bool   `json:"success"` // Red or green
	Toast   string `json:"toast"`   // Browser notification
	Log     error  `json:"log"`     // Console log
}

func playersHistoryAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id", r)
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect(`mean("level")`, "mean_level")
	builder.AddSelect(`mean("games")`, "mean_games")
	builder.AddSelect(`mean("badges")`, "mean_badges")
	builder.AddSelect(`mean("playtime")`, "mean_playtime")
	builder.AddSelect(`mean("friends")`, "mean_friends")
	builder.AddSelect(`mean("level_rank")`, "mean_level_rank")
	builder.AddSelect(`mean("games_rank")`, "mean_games_rank")
	builder.AddSelect(`mean("badges_rank")`, "mean_badges_rank")
	builder.AddSelect(`mean("playtime_rank")`, "mean_playtime_rank")
	builder.AddSelect(`mean("friends_rank")`, "mean_friends_rank")
	builder.SetFrom(helpers.InfluxGameDB, helpers.InfluxRetentionPolicyAllTime.String(), helpers.InfluxMeasurementPlayers.String())
	builder.AddWhere("player_id", "=", id)
	builder.AddWhere("time", ">", "now()-365d")
	builder.AddGroupByTime("1d")
	builder.SetFillNone()

	resp, err := helpers.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc helpers.HighChartsJson

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = helpers.InfluxResponseToHighCharts(resp.Results[0].Series[0])
	}

	err = returnJSON(w, r, hc)
	log.Err(err, r)
}
