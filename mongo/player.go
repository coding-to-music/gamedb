package mongo

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Jleagle/steam-go/steam"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const defaultPlayerAvatar = "/assets/img/no-player-image.jpg"

var (
	ErrInvalidPlayerID   = errors.New("invalid id")
	ErrInvalidPlayerName = errors.New("invalid name")
)

type Player struct {
	ID               int64     `bson:"_id"`             //
	Avatar           string    ``                       //
	Badges           string    ``                       // []ProfileBadge
	BadgesCount      int       `bson:"badges_count"`    //
	BadgeStats       string    `bson:"badge_stats"`     // ProfileBadgeStats
	Bans             string    ``                       // PlayerBans
	CountryCode      string    `bson:"country_code"`    //
	Donated          int       ``                       //
	Friends          string    ``                       // []ProfileFriend
	FriendsCount     int       `bson:"friends_count"`   //
	GamesCount       int       `bson:"games_count"`     //
	GamesRecent      string    `bson:"games_recent"`    // []ProfileRecentGame
	GameStats        string    `bson:"game_stats"`      // PlayerAppStatsTemplate
	Groups           []int     ``                       // []int
	LastLogOff       time.Time `bson:"time_logged_off"` //
	Level            int       ``                       //
	NumberOfGameBans int       `bson:"bans_game"`       //
	NumberOfVACBans  int       `bson:"bans_cav"`        //
	PersonaName      string    `bson:"persona_name"`    //
	PlayTime         int       `bson:"play_time"`       //
	PrimaryClanID    int       `bson:"primary_clan_id"` //
	RealName         string    `bson:"real_name"`       //
	StateCode        string    `bson:"status_code"`     //
	TimeCreated      time.Time `bson:"time_created"`    //
	UpdatedAt        time.Time `bson:"updated_at"`      //
	VanintyURL       string    `bson:"vanity_url"`      //
}

func (player Player) Key() interface{} {
	return player.ID
}

func (player Player) BSON() (ret interface{}) {

	return bson.M{
		"_id":             player.ID,
		"avatar":          player.Avatar,
		"badges":          player.Badges,
		"badges_count":    player.BadgesCount,
		"badge_stats":     player.BadgeStats,
		"bans":            player.Bans,
		"country_code":    player.CountryCode,
		"donated":         player.Donated,
		"friends":         player.Friends,
		"friends_count":   player.FriendsCount,
		"games_count":     player.GamesCount,
		"games_recent":    player.GamesRecent,
		"game_stats":      player.GameStats,
		"groups":          player.Groups,
		"time_logged_off": player.LastLogOff,
		"level":           player.Level,
		"bans_game":       player.NumberOfGameBans,
		"bans_cav":        player.NumberOfVACBans,
		"persona_name":    player.PersonaName,
		"play_time":       player.PlayTime,
		"primary_clan_id": player.PrimaryClanID,
		"real_name":       player.RealName,
		"status_code":     player.StateCode,
		"time_created":    player.TimeCreated,
		"updated_at":      time.Now(),
		"vanity_url":      player.VanintyURL,
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

func (player Player) GetSteamCommunityLink() string {
	return "https://steamcommunity.com/profiles/" + strconv.FormatInt(player.ID, 10)
}

func (player Player) GetMaxFriends() int {
	return helpers.GetPlayerMaxFriends(player.Level)
}

func (player Player) GetAvatar() string {
	if strings.HasPrefix(player.Avatar, "http") {
		return player.Avatar
	} else if player.Avatar != "" {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/" + player.Avatar
	} else {
		return player.GetDefaultAvatar()
	}
}

func (player Player) GetDefaultAvatar() string {
	return defaultPlayerAvatar
}

func (player Player) GetFlag() string {
	return "/assets/img/flags/" + strings.ToLower(player.CountryCode) + ".png"
}

func (player Player) GetCountry() string {
	return helpers.CountryCodeToName(player.CountryCode)
}

func (player Player) GetBadgeStats() (stats ProfileBadgeStats, err error) {

	err = helpers.Unmarshal([]byte(player.BadgeStats), &stats)
	return stats, err
}

func (player Player) GetBadges() (badges []ProfileBadge, err error) {

	if player.Badges == "" || player.Badges == "null" {
		return
	}

	var bytes []byte

	if helpers.IsStorageLocaion(player.Badges) {

		bytes, err = helpers.Download(helpers.PathBadges(player.ID))
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

	if helpers.IsStorageLocaion(player.Friends) {

		bytes, err = helpers.Download(helpers.PathFriends(player.ID))
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

	if helpers.IsStorageLocaion(player.GamesRecent) {

		bytes, err = helpers.Download(helpers.PathRecentGames(player.ID))
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

	if helpers.IsBot(userAgent) {
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

	if !IsValidPlayerID(id) {
		return player, ErrInvalidPlayerID
	}

	err = FindDocument(CollectionPlayers, "_id", id, &player)
	return player, err
}

func GetPlayers(offset int64, ops *options.FindOptions) (players []Player, err error) {

	if ops == nil {
		ops = options.Find()
	}

	client, ctx, err := GetMongo()
	if err != nil {
		return players, err
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionPlayers)

	cur, err := c.Find(ctx, bson.M{}, ops.SetLimit(100).SetSkip(offset))
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
		log.Err(err, player.ID)
		players = append(players, player)
	}

	return players, cur.Err()
}

func GetPlayersByIDs(ids []int64) (players []Player, err error) {

	if len(ids) < 1 {
		return players, nil
	}

	client, ctx, err := GetMongo()
	if err != nil {
		return players, err
	}

	var idsBSON bson.A
	for _, v := range ids {
		idsBSON = append(idsBSON, v)
	}

	c := client.Database(MongoDatabase).Collection(CollectionPlayers)
	cur, err := c.Find(ctx, bson.M{"_id": bson.M{"$in": idsBSON}}, options.Find())
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
		log.Err(err)
		players = append(players, player)
	}

	return players, err
}

func CountPlayers() (count int64, err error) {

	var item = helpers.MemcachePlayersCount

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		return CountDocuments(CollectionPlayers, bson.M{})
	})

	return count, err
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
	return helpers.GetPlayerPath(p.SteamID, p.Name)
}

func (p ProfileFriend) GetDefaultAvatar() string {
	return defaultPlayerAvatar
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
	return helpers.GetPlayerName(p.SteamID, p.Name)
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
	return helpers.GetAppName(p.AppID, p.AppName)
}

func (p ProfileBadge) GetIcon() string {
	if p.AppIcon == "" {
		return helpers.DefaultAppIcon
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
	return helpers.GetTimeShort(int(float64(p.Time)/float64(p.Count)), 2)
}

func (p playerAppStatsInnerTemplate) GetTotalTime() string {
	return helpers.GetTimeShort(p.Time, 2)
}
