package mongo

import (
	"errors"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
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

var (
	ErrInvalidPlayerID   = errors.New("invalid player id")
	ErrInvalidPlayerName = errors.New("invalid player name")
)

type Player struct {
	Avatar            string         `bson:"avatar"`                 //
	BackgroundAppID   int            `bson:"background_app_id"`      //
	BadgeIDs          []int          `bson:"badge_ids"`              // []int - Only special badges
	BadgesCount       int            `bson:"badges_count"`           //
	BadgeStats        string         `bson:"badge_stats"`            // ProfileBadgeStats
	Bans              string         `bson:"bans"`                   // PlayerBans
	CommentsCount     int            `bson:"comments_count"`         //
	ContinentCode     string         `bson:"continent_code"`         // Saved here for easier queries
	CountryCode       string         `bson:"country_code"`           //
	Donated           int            `bson:"donated"`                //
	FriendsCount      int            `bson:"friends_count"`          //
	GamesByType       map[string]int `bson:"games_by_type"`          //
	GamesCount        int            `bson:"games_count"`            //
	GameStats         string         `bson:"game_stats"`             // PlayerAppStatsTemplate
	GroupsCount       int            `bson:"groups_count"`           //
	ID                int64          `bson:"_id"`                    //
	LastBan           time.Time      `bson:"bans_last"`              //
	LastLogOff        time.Time      `bson:"time_logged_off"`        //
	Level             int            `bson:"level"`                  //
	NumberOfGameBans  int            `bson:"bans_game"`              //
	NumberOfVACBans   int            `bson:"bans_cav"`               //
	PersonaName       string         `bson:"persona_name"`           //
	PlayTime          int            `bson:"play_time"`              //
	PrimaryGroupID    string         `bson:"primary_clan_id_string"` //
	Ranks             map[string]int `bson:"ranks"`                  //
	RecentAppsCount   int            `bson:"recent_apps_count"`      //
	StateCode         string         `bson:"status_code"`            //
	TimeCreated       time.Time      `bson:"time_created"`           //
	UpdatedAt         time.Time      `bson:"updated_at"`             //
	VanintyURL        string         `bson:"vanity_url"`             //
	WishlistAppsCount int            `bson:"wishlist_apps_count"`    //
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
		{"time_logged_off", player.LastLogOff},
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
	return player.TimeCreated.Format(helpers.DateYear)
}

func (player Player) GetLogoffUnix() int64 {
	return player.LastLogOff.Unix()
}

func (player Player) GetLogoffNice() string {
	return player.LastLogOff.Format(helpers.DateYearTime)
}

func (player Player) GetUpdatedUnix() int64 {
	return player.UpdatedAt.Unix()
}

func (player Player) GetUpdatedNice() string {
	return player.UpdatedAt.Format(helpers.DateTime)
}

func (player Player) CommunityLink() string {

	if player.VanintyURL != "" {
		return "https://steamcommunity.com/id/" + player.VanintyURL
	}

	return "https://steamcommunity.com/profiles/" + strconv.FormatInt(player.ID, 10)
}

func (player Player) GetStateName() string {

	if player.CountryCode == "" || player.StateCode == "" {
		return ""
	}

	if val, ok := helpers.States[player.CountryCode][player.StateCode]; ok {
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
	return helpers.CountryCodeToName(player.CountryCode)
}

func (player Player) GetBadgeStats() (stats ProfileBadgeStats, err error) {

	err = helpers.Unmarshal([]byte(player.BadgeStats), &stats)
	return stats, err
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
func (player Player) GetSpecialBadges() (badges []PlayerBadge) {

	if player.BadgeIDs == nil || len(player.BadgeIDs) == 0 {
		return
	}

	for _, v := range player.BadgeIDs {

		if val, ok := GlobalBadges[v]; ok {
			badges = append(badges, val)
		}
	}

	sort.Slice(badges, func(i, j int) bool {
		return badges[i].GetUniqueID() > badges[j].GetUniqueID()
	})

	return badges
}

func (player Player) GetBans() (bans PlayerBans, err error) {

	err = helpers.Unmarshal([]byte(player.Bans), &bans)
	return bans, err
}

func (player Player) GetGameStats(code steam.ProductCC) (stats PlayerAppStatsTemplate, err error) {

	err = helpers.Unmarshal([]byte(player.GameStats), &stats)

	stats.All.ProductCC = code
	stats.Played.ProductCC = code

	return stats, err
}

type UpdateType string

const (
	PlayerUpdateAuto   UpdateType = "auto"
	PlayerUpdateManual UpdateType = "manual"
	PlayerUpdateAdmin  UpdateType = "admin"
)

func (player Player) NeedsUpdate(updateType UpdateType) bool {

	if !helpers.IsValidPlayerID(player.ID) {
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
			if player.UpdatedAt.Add(time.Hour * 6).Before(time.Now()) {
				return true
			}
		} else {
			// Donators
			if player.UpdatedAt.Add(time.Minute * 10).Before(time.Now()) {
				return true
			}
		}
	}

	return false
}

func CreatePlayerIndexes() {

	var indexModels []mongo.IndexModel

	// These are for the ranking cron
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

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &player, func() (interface{}, error) {

		if !helpers.IsValidPlayerID(id) {
			return player, ErrInvalidPlayerID
		}

		err = FindOne(CollectionPlayers, bson.D{{"_id", id}}, nil, nil, &player)
		if err != nil {
			return player, err
		}
		if player.ID == 0 {
			return player, ErrNoDocuments
		}

		return player, err
	})

	player.ID = id

	return player, err
}

func GetRandomPlayers(count int) (players []Player, err error) {

	cur, ctx, err := GetRandomRows(CollectionPlayers, count)
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

func SearchPlayer(s string, projection bson.M) (player Player, err error) {

	s = strings.TrimSpace(s)

	if s == "" {
		return player, ErrInvalidPlayerID
	}

	//
	var filter bson.M

	if helpers.RegexNumbers.MatchString(s) {

		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return player, ErrInvalidPlayerID
		}
		filter = bson.M{"_id": i}

	} else {

		// Regex for case insensitivity
		quoted := regexp.QuoteMeta(s)

		filter = bson.M{"$or": bson.A{
			bson.M{"persona_name": bson.M{"$regex": "^" + quoted + "$", "$options": "i"}},
			bson.M{"vanity_url": bson.M{"$regex": "^" + quoted + "$", "$options": "i"}},
		}}
	}

	client, ctx, err := getMongo()
	if err != nil {
		return player, err
	}

	ops := options.FindOne()
	if projection != nil {
		ops.SetProjection(projection)
	}

	c := client.Database(MongoDatabase).Collection(CollectionPlayers.String())
	result := c.FindOne(ctx, filter, ops)

	err = result.Decode(&player)
	return player, err
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

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &codes, func() (interface{}, error) {

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

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &codes, func() (interface{}, error) {

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
				if val, ok := helpers.States[country][stateCode]; ok {
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

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &counts, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return counts, err
		}

		pipeline := mongo.Pipeline{
			// {{Key: "$match", Value: bson.M{"level": bson.M{"$gt": 0}}}},
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

// PlayerAppStatsTemplate
type PlayerAppStatsTemplate struct {
	Played playerAppStatsInnerTemplate
	All    playerAppStatsInnerTemplate
}

type playerAppStatsInnerTemplate struct {
	Count     int
	Price     map[steam.ProductCC]int
	PriceHour map[steam.ProductCC]float64
	Time      int
	ProductCC steam.ProductCC
}

func (p *playerAppStatsInnerTemplate) AddApp(appTime int, prices map[string]int, priceHours map[string]float64) {

	p.Count++

	if p.Price == nil {
		p.Price = map[steam.ProductCC]int{}
	}

	if p.PriceHour == nil {
		p.PriceHour = map[steam.ProductCC]float64{}
	}

	for _, code := range helpers.GetProdCCs(true) {

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

	return helpers.FormatPrice(helpers.GetProdCC(p.ProductCC).CurrencyCode, int(math.Round(float64(p.Price[p.ProductCC])/float64(p.Count))), true)
}

func (p playerAppStatsInnerTemplate) GetTotalPrice() string {

	if p.Count == 0 {
		return "-"
	}

	return helpers.FormatPrice(helpers.GetProdCC(p.ProductCC).CurrencyCode, p.Price[p.ProductCC], true)
}

func (p playerAppStatsInnerTemplate) GetAveragePriceHour() string {

	if p.Count == 0 {
		return "-"
	}

	return helpers.FormatPrice(helpers.GetProdCC(p.ProductCC).CurrencyCode, int(p.PriceHour[p.ProductCC]/float64(p.Count)), true)
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
