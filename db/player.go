package db

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/storage"
	"github.com/Jleagle/steam-go/steam"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gosimple/slug"
)

const defaultPlayerAvatar = "/assets/img/no-player-image.jpg"

var (
	ErrInvalidPlayerID   = errors.New("invalid id")
	ErrInvalidPlayerName = errors.New("invalid name")
)

type Player struct {
	Avatar           string    `datastore:"avatar,noindex"`           //
	Badges           string    `datastore:"badges,noindex"`           // []ProfileBadge
	BadgesCount      int       `datastore:"badges_count"`             //
	BadgeStats       string    `datastore:"badge_stats,noindex"`      // ProfileBadgeStats
	Bans             string    `datastore:"bans,noindex"`             // PlayerBans
	CountryCode      string    `datastore:"country_code"`             //
	CreatedAt        time.Time `datastore:"created_at"`               //
	Donated          int       `datastore:"donated"`                  //
	Friends          string    `datastore:"friends,noindex"`          // []ProfileFriend
	FriendsAddedAt   time.Time `datastore:"friends_added_at,noindex"` //
	FriendsCount     int       `datastore:"friends_count"`            //
	GamesCount       int       `datastore:"games_count"`              //
	GamesRecent      string    `datastore:"games_recent,noindex"`     // []ProfileRecentGame
	GameStats        string    `datastore:"game_stats,noindex"`       // PlayerAppStatsTemplate
	Groups           []int     `datastore:"groups,noindex"`           // []int
	LastLogOff       time.Time `datastore:"time_logged_off,noindex"`  //
	Level            int       `datastore:"level"`                    //
	NumberOfGameBans int       `datastore:"bans_game"`                //
	NumberOfVACBans  int       `datastore:"bans_cav"`                 //
	PersonaName      string    `datastore:"persona_name"`             //
	PlayerID         int64     `datastore:"player_id"`                //
	PlayTime         int       `datastore:"play_time"`                //
	PrimaryClanID    int       `datastore:"primary_clan_id,noindex"`  //
	RealName         string    `datastore:"real_name,noindex"`        //
	StateCode        string    `datastore:"status_code,noindex"`      //
	TimeCreated      time.Time `datastore:"time_created"`             //
	UpdatedAt        time.Time `datastore:"updated_at"`               //
	VanintyURL       string    `datastore:"vanity_url"`               //
}

func (p Player) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindPlayer, strconv.FormatInt(p.PlayerID, 10), nil)
}

func (p Player) GetPath() string {
	return GetPlayerPath(p.PlayerID, p.GetName())
}

func (p Player) GetName() string {
	return getPlayerName(p.PlayerID, p.PersonaName)
}

func (p Player) GetSteamTimeUnix() int64 {
	return p.TimeCreated.Unix()
}

func (p Player) GetSteamTimeNice() string {
	return p.TimeCreated.Format(helpers.DateYear)
}

func (p Player) GetLogoffUnix() int64 {
	return p.LastLogOff.Unix()
}

func (p Player) GetLogoffNice() string {
	return p.LastLogOff.Format(helpers.DateYearTime)
}

func (p Player) GetUpdatedUnix() int64 {
	return p.UpdatedAt.Unix()
}

func (p Player) GetUpdatedNice() string {
	return p.UpdatedAt.Format(helpers.DateTime)
}

func (p Player) GetSteamCommunityLink() string {
	return "https://steamcommunity.com/profiles/" + strconv.FormatInt(p.PlayerID, 10)
}

func (p Player) GetMaxFriends() int {
	return GetPlayerMaxFriends(p.Level)
}

func (p Player) GetAvatar() string {
	if strings.HasPrefix(p.Avatar, "http") {
		return p.Avatar
	} else if p.Avatar != "" {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/" + p.Avatar
	} else {
		return p.GetDefaultAvatar()
	}
}

func (p Player) GetDefaultAvatar() string {
	return defaultPlayerAvatar
}

func (p Player) GetFlag() string {
	return "/assets/img/flags/" + strings.ToLower(p.CountryCode) + ".png"
}

func (p Player) GetCountry() string {
	return helpers.CountryCodeToName(p.CountryCode)
}

func (p Player) GetAppIDs() (appIDs []int, err error) {

	if p.GamesCount == 0 {
		return
	}

	client, ctx, err := GetDSClient()
	if err != nil {
		return
	}

	var playerApps []PlayerApp
	q := datastore.NewQuery(KindPlayerApp).Filter("player_id =", p.PlayerID).KeysOnly()
	keys, err := client.GetAll(ctx, q, &playerApps)
	if err != nil {
		return
	}

	for _, v := range keys {
		_, appID, err := ParsePlayerAppKey(v)
		if err != nil {
			appIDs = append(appIDs, appID)
		}
	}

	return appIDs, nil
}

func (p Player) GetBadgeStats() (stats ProfileBadgeStats, err error) {

	err = helpers.Unmarshal([]byte(p.BadgeStats), &stats)
	return stats, err
}

func (p Player) GetBadges() (badges []ProfileBadge, err error) {

	if p.Badges == "" || p.Badges == "null" {
		return
	}

	var bytes []byte

	if helpers.IsStorageLocaion(p.Badges) {

		bytes, err = helpers.Download(helpers.PathBadges(p.PlayerID))
		err = helpers.IgnoreErrors(err, storage.ErrObjectNotExist)
		if err != nil {
			return badges, err
		}
	} else {
		bytes = []byte(p.Badges)
	}

	err = helpers.Unmarshal(bytes, &badges)
	return badges, err
}

func (p Player) GetFriends() (friends []ProfileFriend, err error) {

	if p.Friends == "" || p.Friends == "null" {
		return
	}

	var bytes []byte

	if helpers.IsStorageLocaion(p.Friends) {

		bytes, err = helpers.Download(helpers.PathFriends(p.PlayerID))
		err = helpers.IgnoreErrors(err, storage.ErrObjectNotExist)
		if err != nil {
			return friends, err
		}
	} else {
		bytes = []byte(p.Friends)
	}

	err = helpers.Unmarshal(bytes, &friends)
	return friends, err
}

func (p Player) GetRecentGames() (games []ProfileRecentGame, err error) {

	if p.GamesRecent == "" || p.GamesRecent == "null" {
		return
	}

	var bytes []byte

	if helpers.IsStorageLocaion(p.GamesRecent) {

		bytes, err = helpers.Download(helpers.PathRecentGames(p.PlayerID))
		err = helpers.IgnoreErrors(err, storage.ErrObjectNotExist)
		if err != nil {
			return games, err
		}
	} else {
		bytes = []byte(p.GamesRecent)
	}

	err = helpers.Unmarshal(bytes, &games)
	return games, err
}

func (p Player) GetBans() (bans PlayerBans, err error) {

	err = helpers.Unmarshal([]byte(p.Bans), &bans)
	return bans, err
}

func (p Player) GetGameStats(code steam.CountryCode) (stats PlayerAppStatsTemplate, err error) {

	err = helpers.Unmarshal([]byte(p.GameStats), &stats)

	stats.All.Code = code
	stats.Played.Code = code

	return stats, err
}

func (p Player) GetTimeShort() (ret string) {
	return helpers.GetTimeShort(p.PlayTime, 2)
}

func (p Player) GetTimeLong() (ret string) {
	return helpers.GetTimeLong(p.PlayTime, 5)
}

type UpdateType string

const (
	PlayerUpdateAuto    UpdateType = "auto"
	PlayerUpdateManual  UpdateType = "manual"
	PlayerUpdateFriends UpdateType = "friends"
	PlayerUpdateAdmin   UpdateType = "admin"
)

func (p Player) ShouldUpdate(userAgent string, updateType UpdateType) bool {

	if !IsValidPlayerID(p.PlayerID) {
		return false
	}

	if helpers.IsBot(userAgent) {
		return false
	}

	switch updateType {
	case PlayerUpdateAdmin:
		return true
	case PlayerUpdateFriends:
		if p.FriendsAddedAt.Add(time.Hour * 24 * 365).Unix() < time.Now().Unix() { // 1 year
			return true
		}
	case PlayerUpdateAuto:
		if p.UpdatedAt.Add(time.Hour * 24 * 7).Unix() < time.Now().Unix() { // 1 week
			return true
		}
	case PlayerUpdateManual:
		if p.Donated == 0 {
			if p.UpdatedAt.Add(time.Hour * 24).Unix() < time.Now().Unix() { // 1 day
				return true
			}
		} else {
			if p.UpdatedAt.Add(time.Hour * 1).Unix() < time.Now().Unix() { // 1 hour
				return true
			}
		}
	}

	return false
}

func (p *Player) Save() (err error) {

	if !IsValidPlayerID(p.PlayerID) {
		return ErrInvalidPlayerID
	}

	// Fix dates
	p.UpdatedAt = time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}

	return SaveKind(p.GetKey(), p)
}

func GetPlayerMaxFriends(level int) (ret int) {

	ret = 750

	if level > 100 {
		ret = 1250
	}

	if level > 200 {
		ret = 1750
	}

	if level > 300 {
		ret = 2000
	}

	return ret
}

func GetPlayerPath(id int64, name string) string {

	p := "/players/" + strconv.FormatInt(id, 10)
	if name != "" {
		s := slug.Make(name)
		if s != "" {
			p = p + "/" + s
		}
	}
	return p
}

func getPlayerName(id int64, name string) string {
	if name != "" {
		return name
	} else if id > 0 {
		return "Player " + strconv.FormatInt(id, 10)
	} else {
		return "Unknown Player"
	}
}

func IsValidPlayerID(id int64) bool {

	if id == 0 {
		return false
	}

	if id < 10000000000000000 {
		return false
	}

	if len(strconv.FormatInt(id, 10)) != 17 {
		return false
	}

	return true
}

func GetPlayer(id int64) (ret Player, err error) {

	if !IsValidPlayerID(id) {
		return ret, ErrInvalidPlayerID
	}

	client, ctx, err := GetDSClient()
	if err != nil {
		return ret, err
	}

	key := datastore.NameKey(KindPlayer, strconv.FormatInt(id, 10), nil)

	player := Player{}
	player.PlayerID = id

	err = client.Get(ctx, key, &player)
	err = handleDSSingleError(err, oldPlayerFields)
	return player, err
}

func GetPlayerByName(name string) (player Player, err error) {

	if len(name) == 0 {
		return player, ErrInvalidPlayerName
	}

	client, ctx, err := GetDSClient()
	if err != nil {
		return player, err
	}

	var players = make([]Player, 0, 1)

	_, err = client.GetAll(ctx, datastore.NewQuery(KindPlayer).Filter("vanity_url =", name).Limit(1), &players)
	if err == nil && len(players) > 0 {
		return players[0], err
	}

	_, err = client.GetAll(ctx, datastore.NewQuery(KindPlayer).Filter("persona_name =", name).Limit(1), &players)
	if err == nil && len(players) > 0 {
		return players[0], err
	}

	users, err := GetUsersByEmail(name)
	if err == nil && len(users) > 0 {

		players, err := GetPlayersByIDs([]int64{users[0].PlayerID})
		if err == nil && len(players) > 0 {
			return players[0], err
		}
	}

	return player, err
}

func GetAllBrokenPlayers() (brokenPlayers []int64) {

	client, ctx, err := GetDSClient()
	if err != nil {
		log.Err(err)
		return
	}

	var players []Player
	allKeys, err := client.GetAll(ctx, datastore.NewQuery(KindPlayer).KeysOnly(), &players)
	if err != nil {
		log.Err(err)
		return
	}

	keyChunks := chunkKeys(allKeys)
	for _, chunk := range keyChunks {

		players := make([]Player, len(chunk))
		err = client.GetMulti(ctx, chunk, players)

		var count = 0

		if multiErr, ok := err.(datastore.MultiError); ok {

			for k, v := range multiErr {

				if v != nil {

					count++

					i, err := strconv.ParseInt(chunk[k].Name, 10, 64)
					log.Err(err)

					brokenPlayers = append(brokenPlayers, i)
				}
			}
		}

		log.Info("Found " + strconv.Itoa(count) + " broken players")
	}

	return brokenPlayers
}

func GetAllPlayers(order string, limit int, keysOnly bool) (players []Player, keys []*datastore.Key, err error) {

	client, ctx, err := GetDSClient()
	if err != nil {
		return players, keys, err
	}

	q := datastore.NewQuery(KindPlayer)

	if keysOnly {
		q = q.KeysOnly()
	}

	if order != "" {
		q = q.Order(order)
	}

	if limit > 0 {
		q = q.Limit(limit)
	}

	keys, err = client.GetAll(ctx, q, &players)
	err = handleDSMultiError(err, oldPlayerFields)
	return players, keys, err
}

func GetPlayersByIDs(ids []int64) (players []Player, err error) {

	if len(ids) > 2000 { // Max friends limit
		return players, ErrorTooMany
	}

	client, ctx, err := GetDSClient()
	if err != nil {
		return players, err
	}

	var keys []*datastore.Key
	for _, v := range ids {
		keys = append(keys, datastore.NameKey(KindPlayer, strconv.FormatInt(v, 10), nil))
	}

	chunks := chunkKeys(keys)
	for _, chunk := range chunks {

		playersChunk := make([]Player, len(chunk))

		err = client.GetMulti(ctx, chunk, playersChunk)
		err = handleDSMultiError(err, oldPlayerFields)
		if err != nil {
			return players, err
		}

		players = append(players, playersChunk...)
	}

	return players, nil
}

func CountPlayers() (count int, err error) {

	var item = helpers.MemcacheCountPlayers

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		var count int

		client, ctx, err := GetDSClient()
		if err != nil {
			return count, err
		}

		q := datastore.NewQuery(KindPlayer)
		count, err = client.Count(ctx, q)
		return count, err
	})

	return count, err
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
	return GetPlayerPath(p.SteamID, p.Name)
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
	return getPlayerName(p.SteamID, p.Name)
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
	return GetAppPath(p.AppID, p.AppName)
}

func (p ProfileBadge) GetAppName() string {
	if p.AppID == 0 {
		return "No App"
	}
	return getAppName(p.AppID, p.AppName)
}

func (p ProfileBadge) GetIcon() string {
	if p.AppIcon == "" {
		return DefaultAppIcon
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
