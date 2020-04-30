package mongo

import (
	"math"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/Jleagle/steam-go/steamid"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/i18n"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	steamHelper "github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var CountriesWithStates = []string{"AT", "AU", "CA", "FR", "GB", "NZ", "PH", "SI", "US"}

type RankMetric string

func (rk RankMetric) String() string {
	switch rk {
	case RankKeyLevel:
		return "Level"
	case RankKeyBadges:
		return "Badges"
	case RankKeyFriends:
		return "Friends"
	case RankKeyComments:
		return "Comments"
	case RankKeyGames:
		return "Games"
	case RankKeyPlaytime:
		return "Playtime"
	}
	return ""
}

const (
	RankKeyLevel    RankMetric = "l"
	RankKeyBadges   RankMetric = "b"
	RankKeyFriends  RankMetric = "f"
	RankKeyComments RankMetric = "c"
	RankKeyGames    RankMetric = "g"
	RankKeyPlaytime RankMetric = "p"
)

var PlayerRankFields = map[string]RankMetric{
	"level":          RankKeyLevel,
	"games_count":    RankKeyGames,
	"badges_count":   RankKeyBadges,
	"play_time":      RankKeyPlaytime,
	"friends_count":  RankKeyFriends,
	"comments_count": RankKeyComments,
}

var PlayerRankFieldsInflux = map[RankMetric]string{
	RankKeyLevel:    "level_rank",
	RankKeyGames:    "games_rank",
	RankKeyBadges:   "badges_rank",
	RankKeyPlaytime: "playtime_rank",
	RankKeyFriends:  "friends_rank",
	RankKeyComments: "comments_rank",
}

type Player struct {
	Avatar            string                 `bson:"avatar"`
	BackgroundAppID   int                    `bson:"background_app_id"`
	BadgeIDs          []int                  `bson:"badge_ids"` // Only special badges
	BadgesCount       int                    `bson:"badges_count"`
	BadgeStats        ProfileBadgeStats      `bson:"badge_stats"`
	Bans              PlayerBans             `bson:"bans"`
	CommentsCount     int                    `bson:"comments_count"`
	ContinentCode     string                 `bson:"continent_code"`
	CountryCode       string                 `bson:"country_code"`
	Donated           int                    `bson:"donated"`
	FriendsCount      int                    `bson:"friends_count"`
	GamesByType       map[string]int         `bson:"games_by_type"`
	GamesCount        int                    `bson:"games_count"`
	GameStats         PlayerAppStatsTemplate `bson:"game_stats"`
	GroupsCount       int                    `bson:"groups_count"`
	ID                int64                  `bson:"_id"`
	LastBan           time.Time              `bson:"bans_last"`
	Level             int                    `bson:"level"`
	NumberOfGameBans  int                    `bson:"bans_game"`
	NumberOfVACBans   int                    `bson:"bans_cav"`
	PersonaName       string                 `bson:"persona_name"`
	PlayTime          int                    `bson:"play_time"`
	PrimaryGroupID    string                 `bson:"primary_clan_id_string"`
	Ranks             map[string]int         `bson:"ranks"`
	RecentAppsCount   int                    `bson:"recent_apps_count"`
	StateCode         string                 `bson:"status_code"`
	TimeCreated       time.Time              `bson:"time_created"` // Created on Steam
	UpdatedAt         time.Time              `bson:"updated_at"`
	VanintyURL        string                 `bson:"vanity_url"`
	WishlistAppsCount int                    `bson:"wishlist_apps_count"`
}

func (player Player) BSON() bson.D {

	// Stops ranks saving as null
	if player.Ranks == nil {
		player.Ranks = map[string]int{}
	}

	return bson.D{
		{"_id", player.ID},
		{"avatar", player.Avatar},
		{"background_app_id", player.BackgroundAppID},
		{"badge_ids", player.BadgeIDs},
		{"badge_stats", player.BadgeStats},
		{"bans", player.Bans},
		{"continent_code", player.ContinentCode},
		{"country_code", player.CountryCode},
		{"donated", player.Donated},
		{"game_stats", player.GameStats},
		{"games_by_type", player.GamesByType},
		{"bans_last", player.LastBan},
		{"bans_game", player.NumberOfGameBans},
		{"bans_cav", player.NumberOfVACBans},
		{"persona_name", player.PersonaName},
		{"primary_clan_id_string", player.PrimaryGroupID},
		{"status_code", player.StateCode},
		{"time_created", player.TimeCreated},
		{"updated_at", time.Now()},
		{"vanity_url", player.VanintyURL},
		{"wishlist_apps_count", player.WishlistAppsCount},
		{"recent_apps_count", player.RecentAppsCount},
		{"groups_count", player.GroupsCount},
		{"ranks", player.Ranks},

		// Rank Metrics
		{"badges_count", player.BadgesCount},
		{"friends_count", player.FriendsCount},
		{"games_count", player.GamesCount},
		{"level", player.Level},
		{"play_time", player.PlayTime},
		{"comments_count", player.CommentsCount},
	}
}

func (player Player) GetPath() string {
	return helpers.GetPlayerPath(player.ID, player.GetName())
}

func (player Player) GetName() string {
	return helpers.GetPlayerName(player.ID, player.PersonaName)
}

func (player Player) GetSteamTimeUnix() int64 {
	return player.TimeCreated.Unix()
}

func (player Player) GetSteamTimeNice() string {

	if player.TimeCreated.IsZero() || player.TimeCreated.Unix() == 0 {
		return "-"
	}
	return player.TimeCreated.Format(helpers.DateYear)
}

func (player Player) GetUpdatedUnix() int64 {
	return player.UpdatedAt.Unix()
}

func (player Player) GetUpdatedNice() string {
	return player.UpdatedAt.Format(helpers.DateTime)
}

func (player Player) CommunityLink() string {

	if player.VanintyURL != "" && player.VanintyURL != strconv.FormatInt(player.ID, 10) {
		return "https://steamcommunity.com/id/" + player.VanintyURL
	}

	return "https://steamcommunity.com/profiles/" + strconv.FormatInt(player.ID, 10)
}

func (player Player) GetStateName() string {

	if player.CountryCode == "" || player.StateCode == "" {
		return ""
	}

	if val, ok := i18n.States[player.CountryCode][player.StateCode]; ok {
		return val
	}

	return player.StateCode
}

func (player Player) GetMaxFriends() int {
	return helpers.GetPlayerMaxFriends(player.Level)
}

func (player Player) GetAvatar() string {
	return helpers.GetPlayerAvatar(player.Avatar)
}

func (player Player) GetFlag() string {
	return helpers.GetPlayerFlagPath(player.CountryCode)
}

func (player Player) GetCountry() string {
	return i18n.CountryCodeToName(player.CountryCode)
}

func (player Player) GetAvatar2() string {
	return helpers.GetPlayerAvatar2(player.Level)
}

func (player Player) GetTimeShort() (ret string) {
	return helpers.GetTimeShort(player.PlayTime, 2)
}

func (player Player) GetTimeLong() (ret string) {
	return helpers.GetTimeLong(player.PlayTime, 5)
}

//
func (player *Player) SetOwnedGames(saveRows bool) (steamapi.OwnedGames, error) {

	// Grab games from Steam
	resp, b, err := steamHelper.GetSteam().GetOwnedGames(player.ID)
	err = steamHelper.AllowSteamCodes(err, b, nil)
	if err != nil {
		return resp, err
	}

	// Save count
	player.GamesCount = len(resp.Games)

	// Start creating PlayerApp's
	var playerApps = map[int]*PlayerApp{}
	var appPrices = map[int]map[string]int{}
	var appPriceHour = map[int]map[string]float64{}
	var appIDs []int
	var playtime = 0
	for _, v := range resp.Games {
		playtime = playtime + v.PlaytimeForever
		appIDs = append(appIDs, v.AppID)
		playerApps[v.AppID] = &PlayerApp{
			PlayerID: player.ID,
			AppID:    v.AppID,
			AppName:  v.Name,
			AppIcon:  v.ImgIconURL,
			AppTime:  v.PlaytimeForever,
		}
		appPrices[v.AppID] = map[string]int{}
		appPriceHour[v.AppID] = map[string]float64{}
	}

	// Save playtime
	player.PlayTime = playtime

	//
	if !saveRows {
		return resp, nil
	}

	// Getting missing price info from MySQL
	gameRows, err := GetAppsByID(appIDs, bson.M{"_id": 1, "prices": 1, "type": 1})
	if err != nil {
		return resp, err
	}

	player.GamesByType = map[string]int{}

	for _, gameRow := range gameRows {

		// Set games by type
		if _, ok := player.GamesByType[gameRow.GetType()]; ok {
			player.GamesByType[gameRow.GetType()]++
		} else {
			player.GamesByType[gameRow.GetType()] = 1
		}

		//
		for code, vv := range gameRow.Prices {

			vv = gameRow.Prices.Get(code)

			appPrices[gameRow.ID][string(code)] = vv.Final
			if appPrices[gameRow.ID][string(code)] > 0 && playerApps[gameRow.ID].AppTime == 0 {
				appPriceHour[gameRow.ID][string(code)] = -1 // Infinite
			} else if appPrices[gameRow.ID][string(code)] > 0 && playerApps[gameRow.ID].AppTime > 0 {
				appPriceHour[gameRow.ID][string(code)] = (float64(appPrices[gameRow.ID][string(code)]) / 100) / (float64(playerApps[gameRow.ID].AppTime) / 60) * 100
			} else {
				appPriceHour[gameRow.ID][string(code)] = 0 // Free
			}
		}

		//
		playerApps[gameRow.ID].AppPrices = appPrices[gameRow.ID]
		log.Err(err)

		//
		playerApps[gameRow.ID].AppPriceHour = appPriceHour[gameRow.ID]
		log.Err(err)
	}

	// Save playerApps to Datastore
	err = UpdatePlayerApps(playerApps)
	if err != nil {
		return resp, err
	}

	// Get top game for background
	if len(appIDs) > 0 {

		sort.Slice(appIDs, func(i, j int) bool {

			var appID1 = appIDs[i]
			var appID2 = appIDs[j]

			return playerApps[appID1].AppTime > playerApps[appID2].AppTime
		})

		player.BackgroundAppID = appIDs[0]
	}

	// Save stats to player
	var gameStats = PlayerAppStatsTemplate{}
	for _, v := range playerApps {

		gameStats.All.AddApp(v.AppTime, appPrices[v.AppID], appPriceHour[v.AppID])
		if v.AppTime > 0 {
			gameStats.Played.AddApp(v.AppTime, appPrices[v.AppID], appPriceHour[v.AppID])
		}
	}

	player.GameStats = gameStats

	return resp, nil
}

func (player *Player) SetPlayerSummary() error {

	summary, b, err := steamHelper.GetSteam().GetPlayer(player.ID)
	err = steamHelper.AllowSteamCodes(err, b, nil)
	if err != nil {
		return err
	}

	// Avatar
	if summary.AvatarFull != "" && helpers.GetResponseCode(helpers.AvatarBase+summary.AvatarFull) == 200 {
		player.Avatar = summary.AvatarFull
	} else {
		player.Avatar = ""
	}

	//
	if strings.Contains(summary.ProfileURL, "profiles") {
		player.VanintyURL = path.Base(summary.ProfileURL)
	}

	player.CountryCode = summary.CountryCode
	player.ContinentCode = i18n.CountryCodeToContinent(summary.CountryCode)
	player.StateCode = summary.StateCode
	player.PersonaName = summary.PersonaName
	player.TimeCreated = time.Unix(summary.TimeCreated, 0)
	player.PrimaryGroupID = summary.PrimaryClanID

	return err
}

func (player *Player) SetLevel() error {

	level, b, err := steamHelper.GetSteam().GetSteamLevel(player.ID)
	err = steamHelper.AllowSteamCodes(err, b, nil)
	if err != nil {
		return err
	}

	player.Level = level

	return nil
}

func (player *Player) SetFriends(saveRows bool) error {

	// If it's a 401, it returns no results, we dont want to change remove the players friends.
	newFriendsSlice, _, err := steamHelper.GetSteam().GetFriendList(player.ID)
	if err2, ok := err.(steamapi.Error); ok && err2.Code == 401 {
		return nil
	}

	//
	player.FriendsCount = len(newFriendsSlice)

	if !saveRows {
		return nil
	}

	// Get data
	oldFriendsSlice, err := GetFriends(player.ID, 0, 0, nil)
	if err != nil {
		return err
	}

	newFriendsMap := map[int64]steamapi.Friend{}
	for _, friend := range newFriendsSlice {
		newFriendsMap[int64(friend.SteamID)] = friend
	}

	// Friends to add
	var friendIDsToAdd []int64
	var friendsToAdd = map[int64]*PlayerFriend{}
	for _, v := range newFriendsSlice {
		friendIDsToAdd = append(friendIDsToAdd, int64(v.SteamID))
		friendsToAdd[int64(v.SteamID)] = &PlayerFriend{
			PlayerID:     player.ID,
			FriendID:     int64(v.SteamID),
			Relationship: v.Relationship,
			FriendSince:  time.Unix(v.FriendSince, 0),
		}
	}

	// Friends to remove
	var friendsToRem []int64
	for _, v := range oldFriendsSlice {
		if _, ok := newFriendsMap[v.FriendID]; !ok {
			friendsToRem = append(friendsToRem, v.FriendID)
		}
	}

	// Fill in missing map the map
	friendRows, err := GetPlayersByID(friendIDsToAdd, bson.M{
		"_id":             1,
		"avatar":          1,
		"games_count":     1,
		"persona_name":    1,
		"level":           1,
		"time_logged_off": 1,
	})
	if err != nil {
		return err
	}

	for _, friend := range friendRows {
		if friend.ID != 0 {

			friendsToAdd[friend.ID].Avatar = friend.Avatar
			friendsToAdd[friend.ID].Games = friend.GamesCount
			friendsToAdd[friend.ID].Name = friend.GetName()
			friendsToAdd[friend.ID].Level = friend.Level
		}
	}

	// Update DB
	err = DeleteFriends(player.ID, friendsToRem)
	if err != nil {
		return err
	}

	var friendsToAddSlice []*PlayerFriend
	for _, v := range friendsToAdd {
		friendsToAddSlice = append(friendsToAddSlice, v)
	}

	err = UpdateFriends(friendsToAddSlice)
	if err != nil {
		return err
	}

	return nil
}

type UpdateType string

const (
	PlayerUpdateAuto   UpdateType = "auto"
	PlayerUpdateManual UpdateType = "manual"
	PlayerUpdateAdmin  UpdateType = "admin"
)

func (player Player) NeedsUpdate(updateType UpdateType) bool {

	var err error
	player.ID, err = helpers.IsValidPlayerID(player.ID)
	if err != nil {
		return false
	}

	switch updateType {
	case PlayerUpdateAdmin:
		return true
	case PlayerUpdateAuto:
		// On page requests
		if player.UpdatedAt.Add(time.Hour * 24).Before(time.Now()) {
			return true
		}
	case PlayerUpdateManual:
		// Non donators
		if player.Donated == 0 {
			if player.UpdatedAt.Add(time.Minute * 10).Before(time.Now()) {
				return true
			}
		} else {
			// Donators
			if player.UpdatedAt.Add(time.Minute * 1).Before(time.Now()) {
				return true
			}
		}
	}

	return false
}

func CreatePlayerIndexes() {

	var indexModels []mongo.IndexModel

	// These are for the ranking cron
	// And for players table  filtering
	for col := range PlayerRankFields {
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{{col, -1}},
		})
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{{"continent_code", 1}, {col, -1}},
		})
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{{"country_code", 1}, {col, -1}},
		})
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{{"country_code", 1}, {"status_code", 1}, {col, -1}},
		})
	}

	// For sorting main players table
	cols := []string{
		"level",
		"badges_count",
		"games_count",
		"play_time",
		"bans_game",
		"bans_cav",
		"bans_last",
		"friends_count",
		"comments_count",
	}

	for _, col := range cols {
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{{col, -1}},
		})
	}

	// Text index
	indexModels = append(indexModels, mongo.IndexModel{
		Keys:    bson.D{{"persona_name", "text"}, {"vanity_url", "text"}},
		Options: options.Index().SetName("text").SetWeights(bson.D{{"persona_name", 1}, {"vanity_url", 1}}),
	})

	// For player search in chatbot
	indexModels = append(indexModels, mongo.IndexModel{
		Keys: bson.D{{"persona_name", 1}},
		Options: options.Index().SetCollation(&options.Collation{
			Locale:   "en",
			Strength: 2, // Case insensitive
		}),
	})
	indexModels = append(indexModels, mongo.IndexModel{
		Keys: bson.D{{"vanity_url", 1}},
		Options: options.Index().SetCollation(&options.Collation{
			Locale:   "en",
			Strength: 2, // Case insensitive
		}),
	})

	//
	client, ctx, err := getMongo()
	if err != nil {
		log.Err(err)
		return
	}

	_, err = client.Database(MongoDatabase).Collection(CollectionPlayers.String()).Indexes().CreateMany(ctx, indexModels)
	log.Err(err)
}

func GetPlayer(id int64) (player Player, err error) {

	var item = memcache.MemcachePlayer(id)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &player, func() (interface{}, error) {

		id, err := helpers.IsValidPlayerID(id)
		if err != nil {
			return player, helpers.ErrInvalidPlayerID
		}

		err = FindOne(CollectionPlayers, bson.D{{"_id", id}}, nil, nil, &player)
		return player, err
	})

	player.ID = id

	return player, err
}

func GetRandomPlayers(count int) (players []Player, err error) {

	cur, ctx, err := GetRandomRows(CollectionPlayers, count, nil, nil)
	if err != nil {
		return players, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var player Player
		err := cur.Decode(&player)
		if err != nil {
			log.Err(err, player.ID)
		}
		players = append(players, player)
	}

	return players, cur.Err()
}

func SearchPlayer(search string, projection bson.M) (player Player, queue bool, err error) {

	search = strings.TrimSpace(search)

	if search == "" {
		return player, false, helpers.ErrInvalidPlayerID
	}

	//
	var ops = options.FindOne()

	// Set to case insensitive
	ops.SetCollation(&options.Collation{
		Locale:   "en",
		Strength: 2,
	})

	if projection != nil {
		ops.SetProjection(projection)
	}

	client, ctx, err := getMongo()
	if err != nil {
		return player, false, err
	}

	c := client.Database(MongoDatabase).Collection(CollectionPlayers.String())

	// Get by ID
	id, err := steamid.ParsePlayerID(search)
	if err == nil {

		err = c.FindOne(ctx, bson.D{{"_id", id}}, ops).Decode(&player)
		err = helpers.IgnoreErrors(err, ErrNoDocuments)
		if err != nil {
			log.Err(err)
		}
	}

	if player.ID == 0 {

		err = c.FindOne(ctx, bson.D{{"persona_name", search}}, ops).Decode(&player)
		err = helpers.IgnoreErrors(err, ErrNoDocuments)
		if err != nil {
			log.Err(err)
		}
	}

	if player.ID == 0 {

		err = c.FindOne(ctx, bson.D{{"vanity_url", search}}, ops).Decode(&player)
		err = helpers.IgnoreErrors(err, ErrNoDocuments)
		if err != nil {
			log.Err(err)
		}
	}

	if player.ID == 0 {

		resp, _, err := steamHelper.GetSteam().ResolveVanityURL(search, steamapi.VanityURLProfile)
		if err == nil && resp.Success > 0 {

			player.ID = int64(resp.SteamID)

			var summaryLock sync.Mutex
			var gamesLock sync.Mutex
			var wg sync.WaitGroup
			for k := range projection {

				switch k {
				case "level":

					wg.Add(1)
					go func() {
						defer wg.Done()
						err = player.SetLevel()
						log.Err(err)
					}()

				case "persona_name", "avatar":

					wg.Add(1)
					go func() {
						defer wg.Done()
						summaryLock.Lock()
						if player.TimeCreated.IsZero() || player.TimeCreated.Unix() == 0 {
							err = player.SetPlayerSummary()
							log.Err(err)
						}
						summaryLock.Unlock()
					}()

				case "games_count", "play_time":

					wg.Add(1)
					go func() {
						defer wg.Done()
						gamesLock.Lock()
						if player.PlayTime > 0 {
							_, err = player.SetOwnedGames(false)
							log.Err(err)
						}
						gamesLock.Unlock()
					}()

				case "friends_count":

					wg.Add(1)
					go func() {
						defer wg.Done()
						err = player.SetFriends(false)
						log.Err(err)
					}()

				}
			}
			wg.Wait()
		}
	}

	if player.ID == 0 {
		return player, false, mongo.ErrNoDocuments
	}

	return player, true, nil
}

func GetPlayersByID(ids []int64, projection bson.M) (players []Player, err error) {

	if len(ids) < 1 {
		return players, nil
	}

	var idsBSON bson.A
	for _, v := range ids {
		idsBSON = append(idsBSON, v)
	}

	return GetPlayers(0, 0, nil, bson.D{{"_id", bson.M{"$in": idsBSON}}}, projection)
}

func GetPlayers(offset int64, limit int64, sort bson.D, filter bson.D, projection bson.M) (players []Player, err error) {

	cur, ctx, err := Find(CollectionPlayers, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return players, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var player Player
		err := cur.Decode(&player)
		if err != nil {
			log.Err(err, player.ID)
		} else {
			players = append(players, player)
		}
	}

	return players, cur.Err()
}

func GetUniquePlayerCountries() (codes []string, err error) {

	var item = memcache.MemcacheUniquePlayerCountryCodes

	err = memcache.GetSetInterface(item.Key, item.Expiration, &codes, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return codes, err
		}

		c := client.Database(MongoDatabase, options.Database()).Collection(CollectionPlayers.String())

		resp, err := c.Distinct(ctx, "country_code", bson.M{}, options.Distinct())
		if err != nil {
			return codes, err
		}

		for _, v := range resp {
			if code, ok := v.(string); ok {
				codes = append(codes, code)
			}
		}

		return codes, err
	})

	return codes, err
}

func GetUniquePlayerStates(country string) (codes []helpers.Tuple, err error) {

	var item = memcache.MemcacheUniquePlayerStateCodes(country)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &codes, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return codes, err
		}

		c := client.Database(MongoDatabase, options.Database()).Collection(CollectionPlayers.String())

		resp, err := c.Distinct(ctx, "status_code", bson.M{"country_code": country}, options.Distinct())
		if err != nil {
			return codes, err
		}

		for _, v := range resp {
			if stateCode, ok := v.(string); stateCode != "" && ok {

				name := stateCode
				if val, ok := i18n.States[country][stateCode]; ok {
					name = val
				}

				codes = append(codes, helpers.Tuple{Key: stateCode, Value: name})
			}
		}

		sort.Slice(codes, func(i, j int) bool {
			return codes[i].Value < codes[j].Value
		})

		return codes, err
	})

	return codes, err
}

func GetPlayerLevels() (counts []count, err error) {

	var item = memcache.MemcachePlayerLevels

	err = memcache.GetSetInterface(item.Key, item.Expiration, &counts, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return counts, err
		}

		pipeline := mongo.Pipeline{
			{{Key: "$group", Value: bson.M{"_id": "$level", "count": bson.M{"$sum": 1}}}},
		}

		cur, err := client.Database(MongoDatabase, options.Database()).Collection(CollectionPlayers.String()).Aggregate(ctx, pipeline, options.Aggregate())
		if err != nil {
			return counts, err
		}

		defer func() {
			err = cur.Close(ctx)
			log.Err(err)
		}()

		var counts []count
		for cur.Next(ctx) {

			var level count
			err := cur.Decode(&level)
			if err != nil {
				log.Err(err, level.ID)
			}
			counts = append(counts, level)
		}

		sort.Slice(counts, func(i, j int) bool {
			return counts[i].ID < counts[j].ID
		})

		return counts, cur.Err()
	})

	return counts, err
}

func GetPlayerLevelsRounded() (counts []count, err error) {

	var item = memcache.MemcachePlayerLevelsRounded

	err = memcache.GetSetInterface(item.Key, item.Expiration, &counts, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return counts, err
		}

		pipeline := mongo.Pipeline{
			{{Key: "$group", Value: bson.M{"_id": bson.M{"$trunc": bson.A{"$level", -1}}, "count": bson.M{"$sum": 1}}}},
		}

		cur, err := client.Database(MongoDatabase, options.Database()).Collection(CollectionPlayers.String()).Aggregate(ctx, pipeline, options.Aggregate())
		if err != nil {
			return counts, err
		}

		defer func() {
			err = cur.Close(ctx)
			log.Err(err)
		}()

		var counts []count
		for cur.Next(ctx) {

			var level count
			err := cur.Decode(&level)
			if err != nil {
				log.Err(err, level.ID)
			}
			counts = append(counts, level)
		}

		sort.Slice(counts, func(i, j int) bool {
			return counts[i].ID < counts[j].ID
		})

		return counts, cur.Err()
	})

	return counts, err
}

func BulkUpdatePlayers(writes []mongo.WriteModel) (err error) {

	if len(writes) == 0 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	c := client.Database(MongoDatabase).Collection(CollectionPlayers.String())

	_, err = c.BulkWrite(ctx, writes, options.BulkWrite().SetOrdered(false))
	return err
}

// ProfileBadgeStats
type ProfileBadgeStats struct {
	PlayerXP                   int
	PlayerLevel                int
	PlayerXPNeededToLevelUp    int
	PlayerXPNeededCurrentLevel int
	PercentOfLevel             int
}

// PlayerBans
type PlayerBans struct {
	CommunityBanned  bool   `json:"community_banned"`
	VACBanned        bool   `json:"vac_banned"`
	NumberOfVACBans  int    `json:"number_of_vac_bans"`
	DaysSinceLastBan int    `json:"days_since_last_ban"`
	NumberOfGameBans int    `json:"number_of_game_bans"`
	EconomyBan       string `json:"economy_ban"`
}

func (pb PlayerBans) History() bool {
	return pb.CommunityBanned || pb.VACBanned || pb.NumberOfVACBans > 0 || pb.DaysSinceLastBan > 0 || pb.NumberOfGameBans > 0 || pb.EconomyBan != "none"
}

// PlayerAppStatsTemplate
type PlayerAppStatsTemplate struct {
	Played playerAppStatsInnerTemplate
	All    playerAppStatsInnerTemplate
}

type playerAppStatsInnerTemplate struct {
	Count     int
	Price     map[steamapi.ProductCC]int
	PriceHour map[steamapi.ProductCC]float64
	Time      int
	ProductCC steamapi.ProductCC
}

func (p *playerAppStatsInnerTemplate) AddApp(appTime int, prices map[string]int, priceHours map[string]float64) {

	p.Count++

	if p.Price == nil {
		p.Price = map[steamapi.ProductCC]int{}
	}

	if p.PriceHour == nil {
		p.PriceHour = map[steamapi.ProductCC]float64{}
	}

	for _, code := range i18n.GetProdCCs(true) {

		// Sometimes priceHour can be -1, meaning infinite
		var priceHour = priceHours[string(code.ProductCode)]
		if priceHour < 0 {
			priceHour = 0
		}

		p.Price[code.ProductCode] = p.Price[code.ProductCode] + prices[string(code.ProductCode)]
		p.PriceHour[code.ProductCode] = p.PriceHour[code.ProductCode] + priceHour
		p.Time = p.Time + appTime
	}
}

func (p playerAppStatsInnerTemplate) GetAveragePrice() string {

	if p.Count == 0 {
		return "-"
	}

	return i18n.FormatPrice(i18n.GetProdCC(p.ProductCC).CurrencyCode, int(math.Round(float64(p.Price[p.ProductCC])/float64(p.Count))), true)
}

func (p playerAppStatsInnerTemplate) GetTotalPrice() string {

	if p.Count == 0 {
		return "-"
	}

	return i18n.FormatPrice(i18n.GetProdCC(p.ProductCC).CurrencyCode, p.Price[p.ProductCC], true)
}

func (p playerAppStatsInnerTemplate) GetAveragePriceHour() string {

	if p.Count == 0 {
		return "-"
	}

	return i18n.FormatPrice(i18n.GetProdCC(p.ProductCC).CurrencyCode, int(p.PriceHour[p.ProductCC]/float64(p.Count)), true)
}

func (p playerAppStatsInnerTemplate) GetAverageTime() string {

	if p.Count == 0 {
		return "-"
	}

	return helpers.GetTimeShort(int(float64(p.Time)/float64(p.Count)), 2)
}

func (p playerAppStatsInnerTemplate) GetTotalTime() string {

	if p.Count == 0 {
		return "-"
	}

	return helpers.GetTimeShort(p.Time, 2)
}
