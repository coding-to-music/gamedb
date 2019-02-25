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
	"github.com/gosimple/slug"
)

const defaultPlayerAvatar = "/assets/img/no-player-image.jpg"

var (
	ErrInvalidPlayerID   = errors.New("invalid id")
	ErrInvalidPlayerName = errors.New("invalid name")

	// ErrUpdatingPlayerTooSoon = errors.New("updating too soon")
	// ErrUpdatingPlayerBot     = errors.New("bots can't update")
	// ErrUpdatingPlayerInQueue = errors.New("player is already in the queue")
)

type Player struct {
	CreatedAt        time.Time `datastore:"created_at"`               //
	UpdatedAt        time.Time `datastore:"updated_at"`               //
	FriendsAddedAt   time.Time `datastore:"friends_added_at,noindex"` //
	PlayerID         int64     `datastore:"player_id"`                //
	VanintyURL       string    `datastore:"vanity_url"`               //
	Avatar           string    `datastore:"avatar,noindex"`           //
	PersonaName      string    `datastore:"persona_name"`             //
	RealName         string    `datastore:"real_name,noindex"`        //
	CountryCode      string    `datastore:"country_code"`             //
	StateCode        string    `datastore:"status_code,noindex"`      //
	Level            int       `datastore:"level"`                    //
	GamesRecent      string    `datastore:"games_recent,noindex"`     // []ProfileRecentGame
	GamesCount       int       `datastore:"games_count"`              //
	GameStats        string    `datastore:"game_stats,noindex"`       // PlayerAppStatsTemplate
	GameHeatMap      string    `datastore:"games_heat_map,noindex"`   // struct // Not used
	Badges           string    `datastore:"badges,noindex"`           // []ProfileBadge
	BadgesCount      int       `datastore:"badges_count"`             //
	BadgeStats       string    `datastore:"badge_stats,noindex"`      // ProfileBadgeStats
	PlayTime         int       `datastore:"play_time"`                //
	TimeCreated      time.Time `datastore:"time_created"`             //
	LastLogOff       time.Time `datastore:"time_logged_off,noindex"`  //
	PrimaryClanID    int       `datastore:"primary_clan_id,noindex"`  //
	Friends          string    `datastore:"friends,noindex"`          // []ProfileFriend
	FriendsCount     int       `datastore:"friends_count"`            //
	Donated          int       `datastore:"donated"`                  //
	Bans             string    `datastore:"bans,noindex"`             // PlayerBans
	NumberOfVACBans  int       `datastore:"bans_cav"`                 //
	NumberOfGameBans int       `datastore:"bans_game"`                //
	Groups           []int     `datastore:"groups,noindex"`           // []int
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

func (p Player) GetBans() (bans steam.GetPlayerBanResponse, err error) {

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

	err = checkForMissingPlayerFields(err)
	if err != nil {
		return player, err
	}

	return player, nil
}

func GetPlayerByName(name string) (player Player, err error) {

	if len(name) == 0 {
		return player, ErrInvalidPlayerName
	}

	client, ctx, err := GetDSClient()
	if err != nil {
		return player, err
	}

	var players []Player

	_, err = client.GetAll(ctx, datastore.NewQuery(KindPlayer).Filter("vanity_url =", name).Limit(1), &players)
	err = checkForMissingPlayerFields(err)
	if err != nil {
		return player, err
	}

	if len(players) == 0 {

		_, err = client.GetAll(ctx, datastore.NewQuery(KindPlayer).Filter("persona_name =", name).Limit(1), &players)
		err = checkForMissingPlayerFields(err)
		if err != nil {
			return player, err
		}
	}

	// Return the first one
	if len(players) > 0 {
		return players[0], nil
	}

	return player, datastore.ErrNoSuchEntity
}

func GetPlayersByEmail(email string) (ret []Player, err error) {

	client, ctx, err := GetDSClient()
	if err != nil {
		return ret, err
	}

	q := datastore.NewQuery(KindPlayer).Filter("settings_email =", email).Limit(1)

	var players []Player

	_, err = client.GetAll(ctx, q, &players)

	err = checkForMissingPlayerFields(err)
	if err != nil {
		return ret, err
	}

	if len(players) == 0 {
		return ret, datastore.ErrNoSuchEntity
	}

	return players, nil
}

func GetAllPlayers(order string, limit int) (players []Player, err error) {

	client, ctx, err := GetDSClient()
	if err != nil {
		return players, err
	}

	q := datastore.NewQuery(KindPlayer).Order(order)

	if limit > 0 {
		q = q.Limit(limit)
	}

	_, err = client.GetAll(ctx, q, &players)

	err = checkForMissingPlayerFields(err)
	return players, err
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

		players = append(players, playersChunk...)

		if checkGetMultiPlayerErrors(err) != nil {
			return players, err
		}
	}

	return players, nil
}

func checkGetMultiPlayerErrors(err error) error {

	if err != nil {

		if multiErr, ok := err.(datastore.MultiError); ok {

			for _, v := range multiErr {
				err2 := checkGetMultiPlayerErrors(v)
				if err2 != nil {
					return err2
				}
			}

		} else if err2, ok := err.(*datastore.ErrFieldMismatch); ok {

			err3 := checkForMissingPlayerFields(err2)
			if err3 != nil {
				return err3
			}

		} else if err.Error() == datastore.ErrNoSuchEntity.Error() {

			return nil

		} else {

			return err

		}
	}

	return nil
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

func checkForMissingPlayerFields(err error) error {

	if err == nil {
		return nil
	}

	if err2, ok := err.(*datastore.ErrFieldMismatch); ok {

		removedColumns := []string{
			"settings_email",
			"settings_password",
			"settings_alerts",
			"settings_hidden",
			"games",
		}

		if helpers.SliceHasString(removedColumns, err2.FieldName) {
			return nil
		}
	}

	return err
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
