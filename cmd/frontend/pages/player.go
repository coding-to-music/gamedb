package pages

import (
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/Jleagle/rabbit-go"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/middleware"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/rabbitweb"
	"github.com/go-chi/chi"
	"github.com/justinas/nosurf"
	"github.com/memcachier/mc/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
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
		log.ErrS(err)
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
			log.Info("player queued", zap.String("ua", ua))
		}
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue, queue.ErrIsBot)
		if err != nil {
			log.ErrS(err)
		}

		// Template
		tm := playerMissingTemplate{}
		tm.fill(w, r, "Looking for player!", "")
		tm.addAssetHighCharts()
		tm.addToast(Toast{Title: "Update", Message: "Player has been queued for an update", Success: true})
		tm.Player = player
		tm.DefaultAvatar = helpers.DefaultPlayerAvatar

		p := rabbit.Payload{}
		p.Preset(rabbit.RangeOneMinute)

		q, err := rabbitweb.GetRabbitWebClient().GetQueue(queue.QueuePlayers, p)
		if err != nil {
			log.ErrS(err)
		} else {
			tm.Queue = q.Messages
		}

		returnTemplate(w, r, "player_missing", tm)
		return

	} else if err != nil {

		log.ErrS(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the player."})
		return
	}

	// Private profile
	if player.Private {
		returnErrorTemplate(w, r, errorTemplate{Code: http.StatusForbidden, Message: "Private profile."})
		return
	}

	var code = session.GetProductCC(r)
	player.GameStats.All.ProductCC = code
	player.GameStats.Played.ProductCC = code

	//
	var wg sync.WaitGroup

	// Number of players
	var playersCount int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		playersCount, err = mongo.CountDocuments(mongo.CollectionPlayers, nil, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	var playersContinentCount int64
	if player.ContinentCode != "" {

		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error
			playersContinentCount, err = mongo.CountDocuments(mongo.CollectionPlayers, bson.D{{"continent_code", player.ContinentCode}}, 60*60*24*7)
			if err != nil {
				log.ErrS(err)
			}
		}()
	}

	var playersCountryCount int64
	if player.CountryCode != "" {

		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error
			playersCountryCount, err = mongo.CountDocuments(mongo.CollectionPlayers, bson.D{{"country_code", player.CountryCode}}, 60*60*24*7)
			if err != nil {
				log.ErrS(err)
			}
		}()
	}

	var playersStateCount int64
	if player.StateCode != "" {

		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error
			playersStateCount, err = mongo.CountDocuments(mongo.CollectionPlayers, bson.D{{"country_code", player.CountryCode}, {"status_code", player.StateCode}}, 60*60*24*7)
			if err != nil {
				log.ErrS(err)
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
				if err != nil {
					log.ErrS(err, player.BackgroundAppID)
				}
			} else if err != nil {
				log.ErrS(err, player.BackgroundAppID)
			}
		}()
	}

	// Type counts
	var typeCounts = map[string]int64{}
	wg.Add(1)
	go func() {

		defer wg.Done()

		var code = session.GetProductCC(r)

		counts, err := mongo.GetAppsGroupedByType(code)
		if err != nil {
			log.ErrS(err)
			return
		}

		for _, v := range counts {
			typeCounts[v.Format()] = v.Count
		}
	}()

	// Player aliases
	var aliases []mongo.PlayerAlias
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		aliases, err = mongo.GetPlayerAliases(player.ID, 0, 20)
		if err != nil {
			log.ErrS(err)
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

		err = helpers.IgnoreErrors(err, mc.ErrNotFound)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get user
	var user mysql.User
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		user, err = mysql.GetUserByKey("steam_id", player.ID, 0)
		err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
		if err != nil {
			log.ErrS(err)
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

	if player.Removed {
		primary = append(primary, "This profile has been removed from Steam")
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
			log.Info("player queued", zap.String("ua", ua))
			t.addToast(Toast{Title: "Update", Message: "Player has been queued for an update", Success: true})
		}
		err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
		if err != nil {
			log.ErrS(err)
		}
	}

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
	t.Aliases = aliases

	for _, v := range mongo.PlayerRankFields {

		rank := playerRankTemplate{}
		rank.Metric = v
		rank.Value = v.Value(player)

		if position, ok := player.Ranks[string(v)]; ok {
			rank.Geos = append(rank.Geos, playerRankGeoTemplate{
				Geo:   "Global",
				Rank:  position,
				Total: playersCount,
			})
		}
		if position, ok := player.Ranks[string(v)+"_continent-"+player.ContinentCode]; ok {
			rank.Geos = append(rank.Geos, playerRankGeoTemplate{
				Geo:   "Continent",
				Rank:  position,
				Total: playersContinentCount,
			})
		}
		if position, ok := player.Ranks[string(v)+"_country-"+player.CountryCode]; ok {
			rank.Geos = append(rank.Geos, playerRankGeoTemplate{
				Geo:   "Country",
				Rank:  position,
				Total: playersCountryCount,
			})
		}
		if position, ok := player.Ranks[string(v)+"_state-"+player.StateCode]; ok {
			rank.Geos = append(rank.Geos, playerRankGeoTemplate{
				Geo:   "State",
				Rank:  position,
				Total: playersStateCount,
			})
		}

		t.Ranks = append(t.Ranks, rank)
	}

	sort.Slice(t.Ranks, func(i, j int) bool {
		return t.Ranks[i].Metric.String() < t.Ranks[j].Metric.String()
	})

	returnTemplate(w, r, "player", t)
}

type playerRankTemplate struct {
	Metric mongo.RankMetric
	Geos   []playerRankGeoTemplate
	Value  int
}

func (rank playerRankTemplate) GetValue() string {

	if rank.Metric == mongo.RankKeyPlaytime {
		return helpers.GetTimeLong(rank.Value, 2)
	}
	return humanize.Comma(int64(rank.Value))
}

type playerRankGeoTemplate struct {
	Geo   string
	Rank  int
	Total int64
}

func (geo playerRankGeoTemplate) Percentile() string {

	p := float64(geo.Rank) / float64(geo.Total) * 100

	if p < 1 {
		return helpers.FloatToString(p, 2)
	} else if p < 10 {
		return helpers.FloatToString(p, 1)
	} else {
		return helpers.FloatToString(p, 0)
	}
}

func (geo playerRankGeoTemplate) GetPlayers() string {
	return humanize.FormatFloat("#,###.", float64(geo.Total))
}

type playerTemplate struct {
	globalTemplate
	Banners       map[string][]string
	CSRF          string
	DefaultAvatar string
	Player        mongo.Player
	Ranks         []playerRankTemplate
	Types         map[string]int64
	InQueue       bool
	User          mysql.User
	WishListTotal string
	Aliases       []mongo.PlayerAlias
}

func (t playerTemplate) includes() []string {
	return []string{"includes/social.gohtml"}
}

func (t playerTemplate) TypePercent(typex string) string {

	if _, ok := t.Player.GamesByType[typex]; !ok {
		return "0%"
	}

	if _, ok := t.Types[typex]; !ok {
		return "0%"
	}

	f := float64(t.Player.GamesByType[typex]) / float64(t.Types[typex]) * 100

	return helpers.FloatToString(f, 2) + "%"
}

type playerMissingTemplate struct {
	globalTemplate
	Player        mongo.Player
	DefaultAvatar string
	Queue         int
}

func playerAddFriendsHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	defer func() {
		session.Save(w, r)
		http.Redirect(w, r, "/players/"+id+"#friends", http.StatusFound)
	}()

	idx, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return
	}

	if !session.IsAdmin(r) {

		user, err := getUserFromSession(r)
		if err != nil {
			log.ErrS(err)
			session.SetFlash(r, session.SessionBad, "Something went wrong")
			return
		}

		playerID := session.GetPlayerIDFromSesion(r)
		if playerID != idx {
			session.SetFlash(r, session.SessionBad, "Invalid user")
			return
		}

		if user.Level <= mysql.UserLevel2 {
			session.SetFlash(r, session.SessionBad, "Invalid user level")
			return
		}
	}

	//
	var friendIDs []int64
	var friendIDsMap = map[int64]bool{}

	friends, err := mongo.GetFriends(idx, 0, 0, nil)
	if err != nil {
		log.ErrS(err)
	}
	for _, v := range friends {
		friendIDs = append(friendIDs, v.FriendID)
		friendIDsMap[v.FriendID] = true
	}

	// Remove players we already have
	players, err := mongo.GetPlayersByID(friendIDs, bson.M{"_id": 1})
	if err != nil {
		log.ErrS(err)
	}
	for _, v := range players {
		delete(friendIDsMap, v.ID)
	}

	// Queue the rest
	for friendID := range friendIDsMap {

		ua := r.UserAgent()
		err = queue.ProducePlayer(queue.PlayerMessage{ID: friendID, UserAgent: &ua})
		if err == nil {
			log.Info("player queued", zap.String("ua", ua))
		}
		err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
		if err != nil {
			log.ErrS(err)
		}
	}

	session.SetFlash(r, session.SessionGood, strconv.Itoa(len(friendIDsMap))+" friends queued")
}

func playerGamesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.ErrS(err)
		return
	}

	var query = datatable.NewDataTableQuery(r, false)
	var code = session.GetProductCC(r)

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
			"4": "app_achievements_percent, app_achievements_have x, app_achievements_total x",
		}

		var err error
		playerApps, err = mongo.GetPlayerApps(query.GetOffset64(), 100, filter2, query.GetOrderMongo(columns))
		if err != nil {
			log.ErrS(err)
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
			log.ErrS(err)
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
			log.ErrS(err)
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
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		player, err := mongo.GetPlayer(id)
		if err != nil {
			log.ErrS(err)
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
		log.ErrS(err)
		return
	}

	query := datatable.NewDataTableQuery(r, true)

	//
	var wg sync.WaitGroup

	// Get achievements
	var playerAchievements []mongo.PlayerAchievement
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		var columns = map[string]string{
			"1": "achievement_date",
			"2": "achievement_complete, achievement_date asc",
		}

		playerAchievements, err = mongo.GetPlayerAchievements(playerID, query.GetOffset64(), query.GetOrderMongo(columns))
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get total
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionPlayerAchievements, bson.D{{"player_id", playerID}}, 0)
		if err != nil {
			log.ErrS(err)
			return
		}
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, count, count, nil)
	for _, pa := range playerAchievements {

		achievedTimeFormatted := time.Unix(pa.AchievementDate, 0).Format(helpers.DateSQL)

		response.AddRow([]interface{}{
			helpers.GetAppPath(pa.AppID, pa.AppName), // 0
			helpers.GetAppName(pa.AppID, pa.AppName), // 1
			helpers.GetAppIcon(pa.AppID, pa.AppIcon), // 2
			pa.AchievementName,                       // 3
			pa.GetAchievementIcon(),                  // 4
			pa.AchievementDescription,                // 5
			pa.AchievementDate,                       // 6
			pa.GetComplete(),                         // 7
			achievedTimeFormatted,                    // 8
		})
	}

	returnJSON(w, r, response)
}

func playerFriendsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	playerID := chi.URLParam(r, "id")

	playerIDInt, err := strconv.ParseInt(playerID, 10, 64)
	if err != nil {
		log.ErrS(err)
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
			"4": "since, avatar desc",
		}

		var err error
		friends, err = mongo.GetFriends(playerIDInt, query.GetOffset64(), 100, query.GetOrderMongo(columns))
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get total
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountFriends(playerIDInt)
		if err != nil {
			log.ErrS(err)
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
			log.ErrS(err)
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
			log.ErrS(err)
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
			log.ErrS(err)
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

		code := session.GetProductCC(r)

		columns := map[string]string{
			"0": "order",
			"1": "app_name",
			"3": "app_release_date",
			"4": "app_prices." + string(code),
		}

		var err error
		wishlistApps, err = mongo.GetPlayerWishlistAppsByPlayer(id, query.GetOffset64(), 0, query.GetOrderMongo(columns), nil)
		if err != nil {
			log.ErrS(err)
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
			log.ErrS(err)
		}

		total = int64(player.WishlistAppsCount)
	}(r)

	wg.Wait()

	var code = session.GetProductCC(r)
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
		log.ErrS(err)
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
			log.ErrS(err)
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
			log.ErrS(err)
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
			group.GetGroupPath(),                   // 3
			group.GetGroupIcon(),                   // 4
			group.GroupMembers,                     // 5
			group.GroupType,                        // 6
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
			log.ErrS(err)
			return "Error looking for player", false, err
		}

		updateType := mongo.PlayerUpdateManual
		if session.IsAdmin(r) {
			message = "Admin update!"
			updateType = mongo.PlayerUpdateAdmin
		}

		if !player.NeedsUpdate(updateType) {
			return "Player can't be updated yet", false, nil
		}

		ua := r.UserAgent()
		err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID, UserAgent: &ua})
		if err == nil {
			log.Info("player queued", zap.String("ua", ua))
		}
		err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
		if err != nil {
			log.ErrS(err)
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
		log.ErrS("invalid id")
		return
	}

	fields := []influx.Field{
		influx.InfPlayersAchievements,
		influx.InfPlayersBadges,
		influx.InfPlayersBadgesFoil,
		influx.InfPlayersComments,
		influx.InfPlayersFriends,
		influx.InfPlayersGames,
		influx.InfPlayersLevel,
		influx.InfPlayersPlaytime,

		influx.InfPlayersAchievementsRank,
		influx.InfPlayersBadgesRank,
		influx.InfPlayersBadgesFoilRank,
		influx.InfPlayersCommentsRank,
		influx.InfPlayersFriendsRank,
		influx.InfPlayersGamesRank,
		influx.InfPlayersLevelRank,
		influx.InfPlayersPlaytimeRank,
	}

	builder := influxql.NewBuilder()
	for _, v := range fields {
		builder.AddSelect(v.Max().String(), v.Max().Alias())
	}

	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementPlayers.String())
	builder.AddWhere("player_id", "=", id)
	builder.AddWhere("time", ">", "now()-365d")
	builder.AddGroupByTime("1d")
	builder.SetFillNone()

	resp, err := influx.InfluxQuery(builder)
	if err != nil {
		log.Err(err.Error(), zap.String("query", builder.String()))
		return
	}

	var hc influx.HighChartsJSON

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0], true)
	}

	returnJSON(w, r, hc)
}
