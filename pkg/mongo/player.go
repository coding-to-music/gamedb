package mongo

import (
	"errors"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrInvalidPlayerID   = errors.New("invalid id")
	ErrInvalidPlayerName = errors.New("invalid name")
)

// Text Index
// {
//   "persona_name": "text",
//   "vanity_url": "text",
// }

type Player struct {
	ID                  int64              `bson:"_id"`                    //
	Avatar              string             `bson:"avatar"`                 //
	BackgroundAppID     int                `bson:"background_app_id"`      //
	BadgeIDs            []int              `bson:"badge_ids"`              // []int - Only special badges
	BadgeStats          string             `bson:"badge_stats"`            // ProfileBadgeStats
	Bans                string             `bson:"bans"`                   // PlayerBans
	CountryCode         string             `bson:"country_code"`           //
	Donated             int                `bson:"donated"`                //
	GameStats           string             `bson:"game_stats"`             // PlayerAppStatsTemplate
	GamesByType         map[string]float64 `bson:"games_by_type"`          //
	Groups              []string           `bson:"groups"`                 // []int - Can be greater than 64bit
	LastLogOff          time.Time          `bson:"time_logged_off"`        //
	NumberOfGameBans    int                `bson:"bans_game"`              //
	NumberOfVACBans     int                `bson:"bans_cav"`               //
	PersonaName         string             `bson:"persona_name"`           //
	PrimaryClanID       int                `bson:"primary_clan_id"`        //
	PrimaryClanIDString string             `bson:"primary_clan_id_string"` //
	StateCode           string             `bson:"status_code"`            //
	TimeCreated         time.Time          `bson:"time_created"`           //
	UpdatedAt           time.Time          `bson:"updated_at"`             //
	VanintyURL          string             `bson:"vanity_url"`             //
	Wishlist            []int              `bson:"wishlist_app_ids"`       //

	// Ranked
	BadgesCount  int `bson:"badges_count"`
	FriendsCount int `bson:"friends_count"`
	GamesCount   int `bson:"games_count"`
	Level        int `bson:"level"`
	PlayTime     int `bson:"play_time"`

	// Ranks
	BadgesRank   int `bson:"badges_rank"`
	FriendsRank  int `bson:"friends_rank"`
	GamesRank    int `bson:"games_rank"`
	LevelRank    int `bson:"level_rank"`
	PlayTimeRank int `bson:"play_time_rank"`
}

func (player Player) BSON() (ret interface{}) {

	return M{
		"_id":                    player.ID,
		"avatar":                 player.Avatar,
		"background_app_id":      player.BackgroundAppID,
		"badge_ids":              player.BadgeIDs,
		"badge_stats":            player.BadgeStats,
		"bans":                   player.Bans,
		"country_code":           player.CountryCode,
		"donated":                player.Donated,
		"game_stats":             player.GameStats,
		"games_by_type":          player.GamesByType,
		"groups":                 player.Groups,
		"time_logged_off":        player.LastLogOff,
		"bans_game":              player.NumberOfGameBans,
		"bans_cav":               player.NumberOfVACBans,
		"persona_name":           player.PersonaName,
		"primary_clan_id":        player.PrimaryClanID,
		"primary_clan_id_string": player.PrimaryClanIDString,
		"status_code":            player.StateCode,
		"time_created":           player.TimeCreated,
		"updated_at":             time.Now(),
		"vanity_url":             player.VanintyURL,
		"wishlist_app_ids":       player.Wishlist,

		// Ranked
		"badges_count":  player.BadgesCount,
		"friends_count": player.FriendsCount,
		"games_count":   player.GamesCount,
		"level":         player.Level,
		"play_time":     player.PlayTime,

		// Ranks
		"badges_rank":    player.BadgesRank,
		"friends_rank":   player.FriendsRank,
		"games_rank":     player.GamesRank,
		"level_rank":     player.LevelRank,
		"play_time_rank": player.PlayTimeRank,
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
	return "https://steamcommunity.com/profiles/" + strconv.FormatInt(player.ID, 10)
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
func (player Player) GetBadgesRank() string {

	if player.BadgesRank == 0 {
		return "-"
	}
	return helpers.OrdinalComma(player.BadgesRank)
}

func (player Player) GetFriendsRank() string {

	if player.FriendsRank == 0 {
		return "-"
	}
	return helpers.OrdinalComma(player.FriendsRank)
}

func (player Player) GetGamesRank() string {

	if player.GamesRank == 0 {
		return "-"
	}
	return helpers.OrdinalComma(player.GamesRank)
}

func (player Player) GetLevelRank() string {

	if player.LevelRank == 0 {
		return "-"
	}
	return helpers.OrdinalComma(player.LevelRank)
}

func (player Player) GetPlaytimeRank() string {

	if player.PlayTimeRank == 0 {
		return "-"
	}
	return helpers.OrdinalComma(player.PlayTimeRank)
}

//
func (player Player) GetSpecialBadges() (badges []PlayerBadge) {

	if player.BadgeIDs == nil || len(player.BadgeIDs) == 0 {
		return
	}

	for _, v := range player.BadgeIDs {

		if val, ok := Badges[v]; ok {
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
		if player.UpdatedAt.Add(time.Hour * 24 * 7).Unix() < time.Now().Unix() { // 1 week
			return true
		}
	case PlayerUpdateManual:
		if player.Donated == 0 {
			if player.UpdatedAt.Add(time.Hour * 24).Unix() < time.Now().Unix() { // 1 day
				return true
			}
		} else {
			if player.UpdatedAt.Add(time.Hour * 1).Unix() < time.Now().Unix() { // 1 hour
				return true
			}
		}
	}

	return false
}

func GetPlayer(id int64) (player Player, err error) {

	var item = helpers.MemcachePlayer(id)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &player, func() (interface{}, error) {

		if !helpers.IsValidPlayerID(id) {
			return player, ErrInvalidPlayerID
		}

		err = FindDocumentByKey(CollectionPlayers, "_id", id, nil, &player)
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

func SearchPlayer(s string, projection M) (player Player, err error) {

	if s == "" {
		return player, ErrInvalidPlayerID
	}

	client, ctx, err := getMongo()
	if err != nil {
		return player, err
	}

	var filter M

	i, _ := strconv.ParseInt(s, 10, 64)
	if helpers.IsValidPlayerID(i) {
		filter = M{"_id": s}
	} else {
		filter = M{"$text": M{"$search": s}}
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

func GetPlayers(offset int64, limit int64, sort D, filter interface{}, projection M, ops *options.FindOptions) (players []Player, err error) {

	return getPlayers(offset, limit, sort, filter, projection, ops)
}

func GetPlayersByID(ids []int64, projection M) (players []Player, err error) {

	if len(ids) < 1 {
		return players, nil
	}

	var idsBSON A
	for _, v := range ids {
		idsBSON = append(idsBSON, v)
	}

	return getPlayers(0, 0, nil, M{"_id": M{"$in": idsBSON}}, projection, nil)
}

func getPlayers(offset int64, limit int64, sort D, filter interface{}, projection M, ops *options.FindOptions) (players []Player, err error) {

	if filter == nil {
		filter = M{}
	}

	client, ctx, err := getMongo()
	if err != nil {
		return players, err
	}

	if ops == nil {
		ops = options.Find()
	}
	if offset > 0 {
		ops.SetSkip(offset)
	}
	if limit > 0 {
		ops.SetLimit(limit)
	}
	if sort != nil {
		ops.SetSort(sort)
	}

	if projection != nil {
		ops.SetProjection(projection)
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionPlayers.String())
	cur, err := c.Find(ctx, filter, ops)
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

func DeletePlayer(id int64) (err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionPlayers.String())

	_, err = c.DeleteOne(ctx, M{"_id": id}, options.Delete())
	return err
}

func CountPlayers() (count int64, err error) {

	var item = helpers.MemcachePlayersCount

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		return CountDocuments(CollectionPlayers, M{}, 0)
	})

	return count, err
}

func CountPlayersWithBan() (count int64, err error) {

	var item = helpers.MemcachePlayersCount

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		filter := M{
			"$or": A{
				M{"bans_game": M{"$gt": 0}},
				M{"bans_cav": M{"$gt": 0}},
			},
		}

		return CountDocuments(CollectionPlayers, filter, 0)
	})

	return count, err
}

func RankPlayers(col string, colToUpdate string) (err error) {

	players, err := getPlayers(0, 0, D{{col, -1}}, M{col: M{"$gt": 0}}, M{"_id": 1}, nil)
	if err != nil {
		return err
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for k, v := range players {

		write := mongo.NewUpdateOneModel()
		write.SetFilter(M{"_id": v.ID})
		write.SetUpdate(M{"$set": M{colToUpdate: k + 1}})
		write.SetUpsert(false)

		writes = append(writes, write)
	}

	c := client.Database(MongoDatabase).Collection(CollectionPlayers.String())

	// Clear all current values
	_, err = c.UpdateMany(ctx, M{colToUpdate: M{"$ne": 0}}, M{"$set": M{colToUpdate: 0}}, options.Update())
	log.Err(err)

	// Write in new values
	_, err = c.BulkWrite(ctx, writes, options.BulkWrite())
	log.Err(err)

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
		p.Price[code.ProductCode] = p.Price[code.ProductCode] + prices[string(code.ProductCode)]
		p.PriceHour[code.ProductCode] = p.PriceHour[code.ProductCode] + priceHours[string(code.ProductCode)]
		p.Time = p.Time + appTime
	}
}

func (p playerAppStatsInnerTemplate) GetAveragePrice() string {
	return helpers.FormatPrice(helpers.GetProdCC(p.ProductCC).CurrencyCode, int(math.Round(float64(p.Price[p.ProductCC])/float64(p.Count))))
}

func (p playerAppStatsInnerTemplate) GetTotalPrice() string {
	return helpers.FormatPrice(helpers.GetProdCC(p.ProductCC).CurrencyCode, p.Price[p.ProductCC])
}

func (p playerAppStatsInnerTemplate) GetAveragePriceHour() string {
	return helpers.FormatPrice(helpers.GetProdCC(p.ProductCC).CurrencyCode, int(p.PriceHour[p.ProductCC]/float64(p.Count)))
}

func (p playerAppStatsInnerTemplate) GetAverageTime() string {
	return helpers.GetTimeShort(int(float64(p.Time)/float64(p.Count)), 2)
}

func (p playerAppStatsInnerTemplate) GetTotalTime() string {
	return helpers.GetTimeShort(p.Time, 2)
}
