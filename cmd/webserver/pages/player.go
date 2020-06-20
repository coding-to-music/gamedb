package pages

import (
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"sync"

	"github.com/Jleagle/influxql"
	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/session-go/session"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/middleware"
	sessionHelpers "github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/i18n"
	"github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
	"github.com/justinas/nosurf"
	"github.com/memcachier/mc"
	"go.mongodb.org/mongo-driver/bson"
)

func PlayerRouter() http.Handler {

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Use(middleware.MiddlewareCSRF)
		r.Get("/", playerHandler)
		r.Get("/{slug}", playerHandler)
		r.Get("/update.json", playersUpdateAjaxHandler)
	})

	r.Get("/add-friends", playerAddFriendsHandler)
	r.Get("/badges.json", playerBadgesAjaxHandler)
	r.Get("/friends.json", playerFriendsAjaxHandler)
	r.Get("/achievements.json", playerAchievementsAjaxHandler)
	// r.Get("/achievements-summary.json", playerAchievementsSummaryAjaxHandler)
	r.Get("/games.json", playerGamesAjaxHandler)
	r.Get("/groups.json", playerGroupsAjaxHandler)
	r.Get("/history.json", playersHistoryAjaxHandler)
	r.Get("/recent.json", playerRecentAjaxHandler)
	r.Get("/wishlist.json", playerWishlistAppsAjaxHandler)
	return r
}

func playerHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Err(err, r)
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid Player ID"})
		return
	}

	id, err = helpers.IsValidPlayerID(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid Player ID"})
		return
	}

	// Find the player row
	player, err := mongo.GetPlayer(id)
	if err == mongo.ErrNoDocuments {

		ua := r.UserAgent()
		err = queue.ProducePlayer(queue.PlayerMessage{ID: id, UserAgent: &ua})
		if err == nil {
			log.Info(log.LogNameTriggerUpdate, r, "new", ua)
		}
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue, queue.ErrIsBot)
		if err != nil {
			log.Err(err, r)
		}

		// Template
		tm := playerMissingTemplate{}
		tm.fill(w, r, "Looking for player!", "")
		tm.addAssetHighCharts()
		tm.addToast(Toast{Title: "Update", Message: "Player has been queued for an update"})
		tm.Player = player
		tm.DefaultAvatar = helpers.DefaultPlayerAvatar

		p := rabbit.Payload{}
		p.Preset(rabbit.RangeOneMinute)

		q, err := helpers.RabbitClient.GetQueue(queue.QueuePlayers, p)
		if err != nil {
			log.Err(err, r)
		} else {
			tm.Queue = q.Messages
		}

		returnTemplate(w, r, "player_missing", tm)
		return

	} else if err != nil {

		log.Err(err, r)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the player."})
		return
	}

	var code = sessionHelpers.GetProductCC(r)
	player.GameStats.All.ProductCC = code
	player.GameStats.Played.ProductCC = code

	//
	var wg sync.WaitGroup

	// Number of players
	var players int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		players, err = mongo.CountDocuments(mongo.CollectionPlayers, nil, 0)
		if err != nil {
			log.Err(err, r)
		}
	}()

	var playersContinent int64
	if player.ContinentCode != "" {

		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error
			playersContinent, err = mongo.CountDocuments(mongo.CollectionPlayers, bson.D{{"continent_code", player.ContinentCode}}, 60*60*24*7)
			if err != nil {
				log.Err(err, r)
			}
		}()
	}

	var playersCountry int64
	if player.CountryCode != "" {

		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error
			playersCountry, err = mongo.CountDocuments(mongo.CollectionPlayers, bson.D{{"country_code", player.CountryCode}}, 60*60*24*7)
			if err != nil {
				log.Err(err, r)
			}
		}()
	}

	var playersState int64
	if player.StateCode != "" {

		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error
			playersState, err = mongo.CountDocuments(mongo.CollectionPlayers, bson.D{{"country_code", player.CountryCode}, {"status_code", player.StateCode}}, 60*60*24*7)
			if err != nil {
				log.Err(err, r)
			}
		}()
	}

	// Get background app
	var backgroundApp mongo.App
	if player.BackgroundAppID > 0 {
		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error
			backgroundApp, err = mongo.GetApp(player.BackgroundAppID)
			err = helpers.IgnoreErrors(err, mongo.ErrInvalidAppID)
			if err == mongo.ErrNoDocuments {
				err = queue.ProduceSteam(queue.SteamMessage{AppIDs: []int{player.BackgroundAppID}})
				log.Err(err, r, player.BackgroundAppID)
			} else if err != nil {
				log.Err(err, r, player.BackgroundAppID)
			}
		}()
	}

	// Type counts
	var typeCounts = map[string]int64{}
	wg.Add(1)
	go func() {

		defer wg.Done()

		var code = sessionHelpers.GetProductCC(r)

		counts, err := mongo.GetAppsGroupedByType(code)
		if err != nil {
			log.Err(err, r)
			return
		}

		for _, v := range counts {
			typeCounts[v.Format()] = v.Count
		}
	}()

	// Check if in queue
	var inQueue bool
	wg.Add(1)
	go func() {

		defer wg.Done()

		item := memcache.MemcachePlayerInQueue(player.ID)
		_, err = memcache.Get(item.Key)

		inQueue = err == nil

		if err != nil {
			err = helpers.IgnoreErrors(err, mc.ErrNotFound)
			log.Err(err, r)
		}
	}()

	// Get user
	var user sql.User
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		user, err = sql.GetUserByKey("steam_id", player.ID, 0)
		err = helpers.IgnoreErrors(err, sql.ErrRecordNotFound)
		if err != nil {
			log.Err(err, r)
			return
		}
	}()

	// Wait
	wg.Wait()

	// Make banners
	banners := make(map[string][]string)
	var primary []string

	if player.ID == 76561197960287930 {
		primary = append(primary, "This profile belongs to Gabe Newell, the Co-founder of Valve")
	}

	if len(primary) > 0 {
		banners["primary"] = primary
	}

	//
	t := playerTemplate{}

	// Add to Rabbit
	if player.NeedsUpdate(mongo.PlayerUpdateAuto) {

		ua := r.UserAgent()
		err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID, UserAgent: &ua})
		if err == nil {
			log.Info(log.LogNameTriggerUpdate, r, ua)
			t.addToast(Toast{Title: "Update", Message: "Player has been queued for an update"})
		}
		err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
		if err != nil {
			log.Err(err, r)
		}
	}

	// Get ranks
	ranks := []mongo.RankMetric{mongo.RankKeyLevel, mongo.RankKeyBadges, mongo.RankKeyFriends, mongo.RankKeyComments, mongo.RankKeyGames, mongo.RankKeyPlaytime}

	for _, v := range ranks {
		if position, ok := player.Ranks[string(v)]; ok {
			t.Ranks = append(t.Ranks, playerRankTemplate{
				Players:  players,
				List:     "Globally",
				Metric:   v,
				Position: position,
			})
		}
		if position, ok := player.Ranks[string(v)+"_continent-"+player.ContinentCode]; ok {
			t.Ranks = append(t.Ranks, playerRankTemplate{
				Players:  playersContinent,
				List:     "In the continent",
				Metric:   v,
				Position: position,
			})
		}
		if position, ok := player.Ranks[string(v)+"_country-"+player.CountryCode]; ok {
			t.Ranks = append(t.Ranks, playerRankTemplate{
				Players:  playersCountry,
				List:     "In the country",
				Metric:   v,
				Position: position,
			})
		}
		if position, ok := player.Ranks[string(v)+"_state-"+player.StateCode]; ok {
			t.Ranks = append(t.Ranks, playerRankTemplate{
				Players:  playersState,
				List:     "In the state",
				Metric:   v,
				Position: position,
			})
		}
	}

	sort.Slice(t.Ranks, func(i, j int) bool {
		return t.Ranks[i].Position < t.Ranks[j].Position
	})

	// Template
	t.setBackground(backgroundApp, true, false)
	t.fill(w, r, player.GetName(), "")
	t.addAssetHighCharts()
	t.IncludeSocialJS = true

	t.Banners = banners
	t.Canonical = player.GetPath()
	t.CSRF = nosurf.Token(r)
	t.DefaultAvatar = helpers.DefaultPlayerAvatar
	t.Player = player
	t.Types = typeCounts
	t.InQueue = inQueue
	t.User = user

	returnTemplate(w, r, "player", t)
}

type playerTemplate struct {
	GlobalTemplate
	Banners       map[string][]string
	CSRF          string
	DefaultAvatar string
	Player        mongo.Player
	Ranks         []playerRankTemplate
	Types         map[string]int64
	InQueue       bool
	User          sql.User
}

func (pt playerTemplate) TypePercent(typex string) string {

	if _, ok := pt.Player.GamesByType[typex]; !ok {
		return "0%"
	}

	if _, ok := pt.Types[typex]; !ok {
		return "0%"
	}

	f := float64(pt.Player.GamesByType[typex]) / float64(pt.Types[typex]) * 100

	return helpers.FloatToString(f, 2) + "%"
}

type playerMissingTemplate struct {
	GlobalTemplate
	Player        mongo.Player
	DefaultAvatar string
	Queue         int
}

type playerRankTemplate struct {
	List     string
	Metric   mongo.RankMetric
	Position int
	Players  int64
}

func (pr playerRankTemplate) Rank() string {
	return helpers.OrdinalComma(pr.Position)
}

func (pr playerRankTemplate) GetPlayers() string {
	return humanize.FormatFloat("#,###.", float64(pr.Players))
}

func (pr playerRankTemplate) Percentile() string {

	p := float64(pr.Position) / float64(pr.Players) * 100

	if p < 1 {
		return helpers.FloatToString(p, 2)
	} else if p < 10 {
		return helpers.FloatToString(p, 1)
	} else {
		return helpers.FloatToString(p, 0)
	}
}

func playerAddFriendsHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	defer func() {
		sessionHelpers.Save(w, r)
		http.Redirect(w, r, "/players/"+id+"#friends", http.StatusFound)
	}()

	idx, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return
	}

	if !sessionHelpers.IsAdmin(r) {

		user, err := getUserFromSession(r)
		if err != nil {
			log.Err(err, r)
			return
		}

		if user.SteamID.String != id {
			err = session.SetFlash(r, sessionHelpers.SessionBad, "Invalid user")
			log.Err(err)
			return
		}

		if user.Level <= sql.UserLevel2 {
			err = session.SetFlash(r, sessionHelpers.SessionBad, "Invalid user level")
			log.Err(err)
			return
		}
	}

	//
	var friendIDs []int64
	var friendIDsMap = map[int64]bool{}

	friends, err := mongo.GetFriends(idx, 0, 0, nil)
	log.Err(err)
	for _, v := range friends {
		friendIDs = append(friendIDs, v.FriendID)
		friendIDsMap[v.FriendID] = true
	}

	// Remove players we already have
	players, err := mongo.GetPlayersByID(friendIDs, bson.M{"_id": 1})
	log.Err(err)
	for _, v := range players {
		delete(friendIDsMap, v.ID)
	}

	// Queue the rest
	for friendID := range friendIDsMap {

		ua := r.UserAgent()
		err = queue.ProducePlayer(queue.PlayerMessage{ID: friendID, UserAgent: &ua})
		if err == nil {
			log.Info(log.LogNameTriggerUpdate, r, ua)
		}
		err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
		if err != nil {
			log.Err(err)
		}
	}

	err = session.SetFlash(r, sessionHelpers.SessionGood, strconv.Itoa(len(friendIDsMap))+" friends queued")
	log.Err(err)
}

func playerGamesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Err(err, r)
		return
	}

	var query = datatable.NewDataTableQuery(r, false)
	var code = sessionHelpers.GetProductCC(r)

	// Make filter
	var filter = bson.D{{"player_id", id}}
	var filter2 = filter

	var search = query.GetSearchString("player-games-search")
	if search != "" {
		quoted := regexp.QuoteMeta(search)
		filter2 = append(filter2, bson.E{Key: "app_name", Value: bson.M{"$regex": quoted, "$options": "i"}})
	}

	//
	var wg sync.WaitGroup

	// Get apps
	var playerApps []mongo.PlayerApp
	wg.Add(1)
	go func() {

		defer wg.Done()

		columns := map[string]string{
			"0": "app_name",
			"1": "app_prices" + "." + string(code),
			"2": "app_time",
			"3": "app_prices_hour" + "." + string(code),
			"4": "app_achievements_percent, app_achievements_have desc",
		}

		var err error
		playerApps, err = mongo.GetPlayerApps(query.GetOffset64(), 100, filter2, query.GetOrderMongo(columns))
		if err != nil {
			log.Err(err)
		}
	}()

	// Get filtered
	var totalFiltered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		totalFiltered, err = mongo.CountDocuments(mongo.CollectionPlayerApps, filter2, 0)
		if err != nil {
			log.Err(err, r)
			return
		}
	}()

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionPlayerApps, filter, 0)
		if err != nil {
			log.Err(err, r)
			return
		}
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, total, totalFiltered, nil)
	for _, pa := range playerApps {
		response.AddRow([]interface{}{
			pa.AppID,                       // 0
			pa.AppName,                     // 1
			pa.GetIcon(),                   // 2
			pa.AppTime,                     // 3
			pa.GetTimeNice(),               // 4
			pa.GetPriceFormatted(code),     // 5
			pa.GetPriceHourFormatted(code), // 6
			pa.GetPath(),                   // 7
			pa.AppAchievementsHave,         // 8
			pa.AppAchievementsTotal,        // 9
			pa.GetAchievementPercent(),     // 10
		})
	}

	returnJSON(w, r, response)
}

func playerRecentAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return
	}

	query := datatable.NewDataTableQuery(r, false)

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
		apps, err = mongo.GetRecentApps(id, query.GetOffset64(), 100, query.GetOrderMongo(columns))
		log.Err(err)
	}()

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		player, err := mongo.GetPlayer(id)
		if err != nil {
			log.Err(err, r)
			return
		}

		total = int64(player.RecentAppsCount)
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, total, total, nil)
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

	returnJSON(w, r, response)
}

func playerAchievementsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	playerID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := datatable.NewDataTableQuery(r, true)

	//
	var wg sync.WaitGroup

	// Get apps
	var playerAchievements []mongo.PlayerAchievement
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		playerAchievements, err = mongo.GetPlayerAchievements(playerID, query.GetOffset64())
		log.Err(err)
	}()

	// Get total
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionPlayerAchievements, bson.D{{"player_id", playerID}}, 0)
		if err != nil {
			log.Err(err, r)
			return
		}
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, count, count, nil)
	for _, pa := range playerAchievements {
		response.AddRow([]interface{}{
			helpers.GetAppPath(pa.AppID, pa.AppName), // 0
			helpers.GetAppName(pa.AppID, pa.AppName), // 1
			helpers.GetAppIcon(pa.AppID, pa.AppIcon), // 2
			pa.AchievementName,                       // 3
			pa.GetAchievementIcon(),                  // 4
			pa.AchievementDescription,                // 5
			pa.AchievementDate,                       // 6
			pa.GetComplete(),                         // 7
		})
	}

	returnJSON(w, r, response)
}

func playerAchievementsSummaryAjaxHandler(w http.ResponseWriter, r *http.Request) {

	playerID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := datatable.NewDataTableQuery(r, true)

	filter := bson.D{
		{"player_id", playerID},
		{"app_achievements_have", bson.M{"$exists": true, "$gt": 0}},
	}

	//
	var wg sync.WaitGroup

	// Get apps
	var playerApps []mongo.PlayerApp
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		var columns = map[string]string{
			"0": "app_name",
			"1": "app_achievements_have, app_achievements_percent desc",
			"2": "app_achievements_percent, app_achievements_have desc",
		}

		playerApps, err = mongo.GetPlayerApps(query.GetOffset64(), 100, filter, query.GetOrderMongo(columns))
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get total
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionPlayerApps, filter, 0)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, count, count, nil)
	for _, pa := range playerApps {
		response.AddRow([]interface{}{
			helpers.GetAppPath(pa.AppID, pa.AppName), // 0
			helpers.GetAppName(pa.AppID, pa.AppName), // 1
			helpers.GetAppIcon(pa.AppID, pa.AppIcon), // 2
			pa.AppID,                                 // 3
			pa.AppAchievementsHave,                   // 4
			pa.AppAchievementsTotal,                  // 5
			pa.GetAchievementPercent(),               // 6
		})
	}

	returnJSON(w, r, response)
}

func playerFriendsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	playerID := chi.URLParam(r, "id")

	playerIDInt, err := strconv.ParseInt(playerID, 10, 64)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := datatable.NewDataTableQuery(r, true)

	//
	var wg sync.WaitGroup

	// Get apps
	var friends []mongo.PlayerFriend
	wg.Add(1)
	go func() {

		defer wg.Done()

		columns := map[string]string{
			"1": "level, avatar desc", // Avatar is to put unscanned players last
			"2": "games, avatar desc",
			"3": "since, avatar desc",
		}

		var err error
		friends, err = mongo.GetFriends(playerIDInt, query.GetOffset64(), 100, query.GetOrderMongo(columns))
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

	var response = datatable.NewDataTablesResponse(r, query, count, count, nil)
	for _, friend := range friends {

		var id = strconv.FormatInt(friend.FriendID, 10)

		response.AddRow([]interface{}{
			id,                      // 0
			friend.GetPath(),        // 1
			friend.GetAvatar(),      // 2
			friend.GetName(),        // 3
			friend.GetLevel(),       // 4
			friend.Scanned(),        // 5
			friend.Games,            // 6
			friend.GetFriendSince(), // 7
			friend.CommunityLink(),  // 8
		})
	}

	returnJSON(w, r, response)
}

func playerBadgesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return
	}

	var query = datatable.NewDataTableQuery(r, false)
	var lock sync.Mutex

	// Make filter
	var filter = bson.D{{Key: "player_id", Value: id}}
	filter2 := filter

	var search = query.GetSearchString("player-badge-search")
	if search != "" {
		quoted := regexp.QuoteMeta(search)
		filter2 = append(filter2, bson.E{Key: "app_name", Value: bson.M{"$regex": quoted, "$options": "i"}})
	}

	//
	var wg sync.WaitGroup

	// Get badges
	var badges []mongo.PlayerBadge //
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		var sortCols = map[string]string{
			"1": "badge_level",
			"2": "badge_scarcity",
			"3": "badge_completion_time",
		}

		badges, err = mongo.GetPlayerBadges(query.GetOffset64(), filter2, query.GetOrderMongo(sortCols))
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		lock.Lock()
		total, err = mongo.CountDocuments(mongo.CollectionPlayerBadges, filter, 0)
		lock.Unlock()
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get filtered
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		lock.Lock()
		filtered, err = mongo.CountDocuments(mongo.CollectionPlayerBadges, filter2, 0)
		lock.Unlock()
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, total, filtered, nil)
	for _, badge := range badges {

		var completionTime = badge.BadgeCompletionTime.Format(helpers.DateSQL)

		var icon = badge.GetIcon()
		if icon == "" {
			icon = helpers.DefaultAppIcon
		}

		response.AddRow([]interface{}{
			badge.AppID,         // 0
			badge.GetName(),     // 1
			badge.GetPath(),     // 2
			completionTime,      // 3
			badge.BadgeFoil,     // 4
			icon,                // 5
			badge.BadgeLevel,    // 6
			badge.BadgeScarcity, // 7
			badge.BadgeXP,       // 8
			badge.IsSpecial(),   // 9
			badge.IsEvent(),     // 10
		})
	}

	returnJSON(w, r, response)
}

func playerWishlistAppsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return
	}

	query := datatable.NewDataTableQuery(r, false)

	//
	var wg sync.WaitGroup

	// Get apps
	var wishlistApps []mongo.PlayerWishlistApp

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		code := sessionHelpers.GetProductCC(r)

		columns := map[string]string{
			"0": "order",
			"1": "app_name",
			"3": "app_release_date",
			"4": "app_prices." + string(code),
		}

		var err error
		wishlistApps, err = mongo.GetPlayerWishlistAppsByPlayer(id, query.GetOffset64(), 0, query.GetOrderMongo(columns))
		if err != nil {
			log.Err(err, r)
			return
		}
	}(r)

	// Get total
	var total int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		player, err := mongo.GetPlayer(id)
		if err != nil {
			log.Err(err, r)
		}

		total = int64(player.WishlistAppsCount)
	}(r)

	wg.Wait()

	var code = sessionHelpers.GetProductCC(r)
	var response = datatable.NewDataTablesResponse(r, query, total, total, nil)
	for _, app := range wishlistApps {

		var priceFormatted string

		if price, ok := app.AppPrices[code]; ok {
			priceFormatted = i18n.FormatPrice(i18n.GetProdCC(code).CurrencyCode, price)
		} else {
			priceFormatted = "-"
		}

		response.AddRow([]interface{}{
			app.AppID,                // 0
			app.GetName(),            // 1
			app.GetPath(),            // 2
			app.GetIcon(),            // 3
			app.Order,                // 4
			app.GetReleaseState(),    // 5
			app.GetReleaseDateNice(), // 6
			priceFormatted,           // 7
		})
	}

	returnJSON(w, r, response)
}

func playerGroupsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return
	}

	player, err := mongo.GetPlayer(id)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := datatable.NewDataTableQuery(r, false)

	//
	var wg sync.WaitGroup

	// Get groups
	var groups []mongo.PlayerGroup
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		columns := map[string]string{
			"0": "group_name",
			"1": "group_members",
		}

		var err error
		groups, err = mongo.GetPlayerGroups(id, query.GetOffset64(), 100, query.GetOrderMongo(columns))
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
		player, err := mongo.GetPlayer(id)
		if err != nil {
			log.Err(err, r)
		}

		total = int64(player.GroupsCount)
	}(r)

	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, total, total, nil)
	for _, group := range groups {

		response.AddRow([]interface{}{
			group.GroupID,                          // 0
			"",                                     // 1
			group.GetName(),                        // 2
			group.GetPath(),                        // 3
			group.GetIcon(),                        // 4
			group.GroupMembers,                     // 5
			group.GetType(),                        // 6
			group.GroupID == player.PrimaryGroupID, // 7
			group.GetURL(),                         // 8
		})
	}

	returnJSON(w, r, response)
}

func playersUpdateAjaxHandler(w http.ResponseWriter, r *http.Request) {

	message, success, err := func(r *http.Request) (string, bool, error) {

		if !nosurf.VerifyToken(nosurf.Token(r), r.URL.Query().Get("csrf")) || r.URL.Query().Get("csrf") == "" {
			return "Invalid CSRF token, please refresh", false, nil
		}

		idx, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			return "Invalid Player ID", false, err
		}

		idx, err = helpers.IsValidPlayerID(idx)
		if err != nil {
			return "Invalid Player ID", false, err
		}

		var message string

		player, err := mongo.GetPlayer(idx)
		if err == nil {
			message = "Updating player!"
		} else if err == mongo.ErrNoDocuments {
			message = "Looking for new player!"
			player = mongo.Player{ID: idx}
		} else {
			log.Err(err, r)
			return "Error looking for player", false, err
		}

		updateType := mongo.PlayerUpdateManual
		if sessionHelpers.IsAdmin(r) {
			message = "Admin update!"
			updateType = mongo.PlayerUpdateAdmin
		}

		if !player.NeedsUpdate(updateType) {
			return "Player can't be updated yet", false, nil
		}

		ua := r.UserAgent()
		err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID, UserAgent: &ua})
		if err == nil {
			log.Info(log.LogNameTriggerUpdate, r, ua)
		}
		err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
		if err != nil {
			log.Err(err, r)
			return "Something has gone wrong", false, err
		}

		return message, true, err
	}(r)

	var response = PlayersUpdateResponse{
		Success: success,
		Toast:   message,
		Log:     err,
	}

	returnJSON(w, r, response)
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
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementPlayers.String())
	builder.AddWhere("player_id", "=", id)
	builder.AddWhere("time", ">", "now()-365d")
	builder.AddGroupByTime("1d")
	builder.SetFillNone()

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc influx.HighChartsJSON

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0])
	}

	returnJSON(w, r, hc)
}
