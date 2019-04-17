package mongo

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Jleagle/steam-go/steam"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/website/pkg"
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
	ID               int64     `bson:"_id"`             //
	Avatar           string    `bson:"avatar"`          //
	Badges           string    `bson:"badges"`          // []ProfileBadge
	BadgeStats       string    `bson:"badge_stats"`     // ProfileBadgeStats
	Bans             string    `bson:"bans"`            // PlayerBans
	CountryCode      string    `bson:"country_code"`    //
	Donated          int       `bson:"donated"`         //
	Friends          string    `bson:"friends"`         // []ProfileFriend
	GamesRecent      string    `bson:"games_recent"`    // []ProfileRecentGame
	GameStats        string    `bson:"game_stats"`      // PlayerAppStatsTemplate
	Groups           []int     `bson:"groups"`          // []int
	LastLogOff       time.Time `bson:"time_logged_off"` //
	NumberOfGameBans int       `bson:"bans_game"`       //
	NumberOfVACBans  int       `bson:"bans_cav"`        //
	PersonaName      string    `bson:"persona_name"`    //
	PrimaryClanID    int       `bson:"primary_clan_id"` //
	RealName         string    `bson:"real_name"`       //
	StateCode        string    `bson:"status_code"`     //
	TimeCreated      time.Time `bson:"time_created"`    //
	UpdatedAt        time.Time `bson:"updated_at"`      //
	VanintyURL       string    `bson:"vanity_url"`      //

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

	return pkg.M{
		"_id":             player.ID,
		"avatar":          player.Avatar,
		"badges":          player.Badges,
		"badge_stats":     player.BadgeStats,
		"bans":            player.Bans,
		"country_code":    player.CountryCode,
		"donated":         player.Donated,
		"friends":         player.Friends,
		"games_recent":    player.GamesRecent,
		"game_stats":      player.GameStats,
		"groups":          player.Groups,
		"time_logged_off": player.LastLogOff,
		"bans_game":       player.NumberOfGameBans,
		"bans_cav":        player.NumberOfVACBans,
		"persona_name":    player.PersonaName,
		"primary_clan_id": player.PrimaryClanID,
		"real_name":       player.RealName,
		"status_code":     player.StateCode,
		"time_created":    player.TimeCreated,
		"updated_at":      time.Now(),
		"vanity_url":      player.VanintyURL,

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
	return pkg.GetPlayerPath(player.ID, player.GetName())
}

func (player Player) GetName() string {
	return pkg.GetPlayerName(player.ID, player.PersonaName)
}

func (player Player) GetSteamTimeUnix() int64 {
	return player.TimeCreated.Unix()
}

func (player Player) GetSteamTimeNice() string {
	return player.TimeCreated.Format(pkg.DateYear)
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
	return player.UpdatedAt.Format(pkg.DateTime)
}

func (player Player) GetSteamCommunityLink() string {
	return "https://steamcommunity.com/profiles/" + strconv.FormatInt(player.ID, 10)
}

func (player Player) GetMaxFriends() int {
	return pkg.GetPlayerMaxFriends(player.Level)
}

func (player Player) GetAvatar() string {
	return pkg.GetPlayerAvatar(player.Avatar)
}

func (player Player) GetFlag() string {
	return pkg.GetPlayerFlagPath(player.CountryCode)
}

func (player Player) GetCountry() string {
	return pkg.CountryCodeToName(player.CountryCode)
}

func (player Player) GetBadgeStats() (stats ProfileBadgeStats, err error) {

	err = helpers.Unmarshal([]byte(player.BadgeStats), &stats)
	return stats, err
}

func (player Player) GetAvatar2() string {
	return pkg.GetAvatar2(player.Level)
}

func (player Player) GetTimeShort() (ret string) {
	return pkg.GetTimeShort(player.PlayTime, 2)
}

func (player Player) GetTimeLong() (ret string) {
	return pkg.GetTimeLong(player.PlayTime, 5)
}

//
func (player Player) GetBadgesRank() string {

	if player.BadgesRank == 0 {
		return "-"
	}
	return pkg.OrdinalComma(player.BadgesRank)
}

func (player Player) GetFriendsRank() string {

	if player.FriendsRank == 0 {
		return "-"
	}
	return pkg.OrdinalComma(player.FriendsRank)
}

func (player Player) GetGamesRank() string {

	if player.GamesRank == 0 {
		return "-"
	}
	return pkg.OrdinalComma(player.GamesRank)
}

func (player Player) GetLevelRank() string {

	if player.LevelRank == 0 {
		return "-"
	}
	return pkg.OrdinalComma(player.LevelRank)
}

func (player Player) GetPlaytimeRank() string {

	if player.PlayTimeRank == 0 {
		return "-"
	}
	return pkg.OrdinalComma(player.PlayTimeRank)
}

//
func (player Player) GetBadges() (badges []ProfileBadge, err error) {

	if player.Badges == "" || player.Badges == "null" {
		return
	}

	var bytes []byte

	if pkg.IsStorageLocaion(player.Badges) {

		bytes, err = pkg.Download(pkg.PathBadges(player.ID))
		err = helpers.IgnoreErrors(err, storage.ErrObjectNotExist)
		if err != nil {
			return badges, err
		}
	} else {
		bytes = []byte(player.Badges)
	}

	err = helpers.Unmarshal(bytes, &badges)
	return badges, err
}

func (player Player) GetFriends() (friends []ProfileFriend, err error) {

	if player.Friends == "" || player.Friends == "null" {
		return
	}

	var bytes []byte

	if pkg.IsStorageLocaion(player.Friends) {

		bytes, err = pkg.Download(pkg.PathFriends(player.ID))
		err = helpers.IgnoreErrors(err, storage.ErrObjectNotExist)
		if err != nil {
			return friends, err
		}
	} else {
		bytes = []byte(player.Friends)
	}

	err = helpers.Unmarshal(bytes, &friends)
	return friends, err
}

func (player Player) GetRecentGames() (games []ProfileRecentGame, err error) {

	if player.GamesRecent == "" || player.GamesRecent == "null" {
		return
	}

	var bytes []byte

	if pkg.IsStorageLocaion(player.GamesRecent) {

		bytes, err = pkg.Download(pkg.PathRecentGames(player.ID))
		err = helpers.IgnoreErrors(err, storage.ErrObjectNotExist)
		if err != nil {
			return games, err
		}
	} else {
		bytes = []byte(player.GamesRecent)
	}

	err = helpers.Unmarshal(bytes, &games)
	return games, err
}

func (player Player) GetBans() (bans PlayerBans, err error) {

	err = helpers.Unmarshal([]byte(player.Bans), &bans)
	return bans, err
}

func (player Player) GetGameStats(code steam.CountryCode) (stats PlayerAppStatsTemplate, err error) {

	err = helpers.Unmarshal([]byte(player.GameStats), &stats)

	stats.All.Code = code
	stats.Played.Code = code

	return stats, err
}

type UpdateType string

const (
	PlayerUpdateAuto   UpdateType = "auto"
	PlayerUpdateManual UpdateType = "manual"
	PlayerUpdateAdmin  UpdateType = "admin"
)

func (player Player) ShouldUpdate(userAgent string, updateType UpdateType) bool {

	if !IsValidPlayerID(player.ID) {
		return false
	}

	if pkg.IsBot(userAgent) {
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

func (player Player) OutputForJSON(rank string) (output []interface{}) {

	return []interface{}{
		rank,                             //
		strconv.FormatInt(player.ID, 10), //
		player.PersonaName,               //
		player.GetAvatar(),               //
		player.GetAvatar2(),              //
		player.Level,                     //
		player.GamesCount,                //
		player.BadgesCount,               //
		player.GetTimeShort(),            //
		player.GetTimeLong(),             //
		player.FriendsCount,              //
		player.GetFlag(),                 //
		player.GetCountry(),              //
	}
}

func GetPlayer(id int64) (player Player, err error) {

	if !IsValidPlayerID(id) {
		return player, ErrInvalidPlayerID
	}

	err = pkg.FindDocument(pkg.CollectionPlayers, "_id", id, nil, &player)
	return player, err
}

func SearchPlayer(s string) (player Player, err error) {

	if s == "" {
		return player, ErrInvalidPlayerID
	}

	client, ctx, err := pkg.getMongo()
	if err != nil {
		return player, err
	}

	var filter pkg.M

	i, _ := strconv.ParseInt(s, 10, 64)
	if pkg.IsValidPlayerID(i) {
		filter = pkg.M{"_id": s}
	} else {
		filter = pkg.M{"$text": pkg.M{"$search": s}}
	}

	c := client.Database(pkg.MongoDatabase).Collection(pkg.CollectionPlayers.String())
	result := c.FindOne(ctx, filter, options.FindOne())

	err = result.Decode(&player)
	return player, err
}

func GetPlayers(offset int64, limit int64, sort pkg.D, filter pkg.M, projection pkg.M) (players []Player, err error) {

	return getPlayers(offset, limit, sort, filter, projection)
}

func GetPlayersByID(ids []int64, projection pkg.M) (players []Player, err error) {

	if len(ids) < 1 {
		return players, nil
	}

	var idsBSON pkg.A
	for _, v := range ids {
		idsBSON = append(idsBSON, v)
	}

	return getPlayers(0, 0, nil, pkg.M{"_id": pkg.M{"$in": idsBSON}}, projection)
}

func getPlayers(offset int64, limit int64, sort pkg.D, filter interface{}, projection pkg.M) (players []Player, err error) {

	if filter == nil {
		filter = pkg.M{}
	}

	client, ctx, err := pkg.getMongo()
	if err != nil {
		return players, err
	}

	ops := options.Find()
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

	c := client.Database(pkg.MongoDatabase, options.Database()).Collection(pkg.CollectionPlayers.String())
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

func CountPlayers() (count int64, err error) {

	var item = pkg.MemcachePlayersCount

	err = pkg.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		return pkg.CountDocuments(pkg.CollectionPlayers, pkg.M{})
	})

	return count, err
}

func RankPlayers(col string, colToUpdate string) (err error) {

	players, err := getPlayers(0, 0, pkg.D{{col, -1}}, pkg.M{col: pkg.M{"$gt": 0}}, pkg.M{"_id": 1})
	if err != nil {
		return err
	}

	client, ctx, err := pkg.getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for k, v := range players {

		write := mongo.NewUpdateOneModel()
		write.SetFilter(pkg.M{"_id": v.ID})
		write.SetUpdate(pkg.M{"$set": pkg.M{colToUpdate: k + 1}})
		write.SetUpsert(false)

		writes = append(writes, write)
	}

	c := client.Database(pkg.MongoDatabase).Collection(pkg.CollectionPlayers.String())

	// Clear all current values
	_, err = c.UpdateMany(ctx, pkg.M{colToUpdate: pkg.M{"$ne": 0}}, pkg.M{"$set": pkg.M{colToUpdate: 0}}, options.Update())
	log.Err(err)

	// Write in new values
	_, err = c.BulkWrite(ctx, writes, options.BulkWrite())
	log.Err(err)

	return err
}

func IsValidPlayerID(id int64) bool {

	if id == 0 {
		return false
	}

	idString := strconv.FormatInt(id, 10)

	if !strings.HasPrefix(idString, "76") {
		return false
	}

	if len(idString) != 17 {
		return false
	}

	return true
}

// ProfileFriend
type ProfileFriend struct {
	SteamID     int64  `json:"id"`
	FriendSince int64  `json:"fs"`
	Avatar      string `json:"ic"`
	Name        string `json:"nm"`
	Games       int    `json:"gm"`
	Level       int    `json:"lv"`
	LoggedOff   int64  `json:"lo"`
}

func (p ProfileFriend) Scanned() bool {
	return p.LoggedOff > 0
}

func (p ProfileFriend) GetPath() string {
	return pkg.GetPlayerPath(p.SteamID, p.Name)
}

func (p ProfileFriend) GetLoggedOff() string {
	if p.Scanned() {
		return time.Unix(p.LoggedOff, 0).Format(helpers.DateYearTime)
	}
	return "-"
}

func (p ProfileFriend) GetFriendSince() string {
	return time.Unix(p.FriendSince, 0).Format(helpers.DateYearTime)
}

func (p ProfileFriend) GetName() string {
	return pkg.GetPlayerName(p.SteamID, p.Name)
}

func (p ProfileFriend) GetLevel() string {
	if p.Scanned() {
		return humanize.Comma(int64(p.Level))
	}
	return "-"
}

// ProfileBadge
type ProfileBadge struct {
	BadgeID        int    `json:"bi"`
	AppID          int    `json:"ai"`
	AppName        string `json:"an"`
	AppIcon        string `json:"ac"`
	Level          int    `json:"lv"`
	CompletionTime int64  `json:"ct"`
	XP             int    `json:"xp"`
	Scarcity       int    `json:"sc"`
}

func (p ProfileBadge) GetTimeFormatted() string {
	return time.Unix(p.CompletionTime, 0).Format(helpers.DateYearTime)
}

func (p ProfileBadge) GetAppPath() string {
	return helpers.GetAppPath(p.AppID, p.AppName)
}

func (p ProfileBadge) GetAppName() string {
	if p.AppID == 0 {
		return "No App"
	}
	return pkg.GetAppName(p.AppID, p.AppName)
}

func (p ProfileBadge) GetIcon() string {
	if p.AppIcon == "" {
		return pkg.DefaultAppIcon
	}
	return p.AppIcon
}

// ProfileBadge
type ProfileBadgeStats struct {
	PlayerXP                   int
	PlayerLevel                int
	PlayerXPNeededToLevelUp    int
	PlayerXPNeededCurrentLevel int
	PercentOfLevel             int
}

// ProfileRecentGame
type ProfileRecentGame struct {
	AppID           int    `json:"i"`
	Name            string `json:"n"`
	PlayTime2Weeks  int    `json:"p"`
	PlayTimeForever int    `json:"f"`
	ImgIconURL      string `json:"c"`
	ImgLogoURL      string `json:"l"`
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
	Price     map[steam.CountryCode]int
	PriceHour map[steam.CountryCode]float64
	Time      int
	Code      steam.CountryCode
}

func (p *playerAppStatsInnerTemplate) AddApp(appTime int, prices map[string]int, priceHours map[string]float64) {

	p.Count++

	for code := range steam.Countries {

		if p.Price == nil {
			p.Price = map[steam.CountryCode]int{}
		}

		if p.PriceHour == nil {
			p.PriceHour = map[steam.CountryCode]float64{}
		}

		p.Price[code] = p.Price[code] + prices[string(code)]
		p.PriceHour[code] = p.PriceHour[code] + priceHours[string(code)]
		p.Time = p.Time + appTime
	}
}

func (p playerAppStatsInnerTemplate) GetAveragePrice() float64 {
	return helpers.RoundFloatTo2DP(float64(p.Price[p.Code]) / float64(p.Count))
}

func (p playerAppStatsInnerTemplate) GetTotalPrice() float64 {
	return helpers.RoundFloatTo2DP(float64(p.Price[p.Code]))
}

func (p playerAppStatsInnerTemplate) GetAveragePriceHour() float64 {
	return helpers.RoundFloatTo2DP(p.PriceHour[p.Code] / float64(p.Count))
}
func (p playerAppStatsInnerTemplate) GetAverageTime() string {
	return pkg.GetTimeShort(int(float64(p.Time)/float64(p.Count)), 2)
}

func (p playerAppStatsInnerTemplate) GetTotalTime() string {
	return pkg.GetTimeShort(p.Time, 2)
}
