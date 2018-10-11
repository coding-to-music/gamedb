package db

import (
	"encoding/json"
	"errors"
	"math"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/dustin/go-humanize"
	"github.com/gosimple/slug"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/memcache"
	"github.com/steam-authority/steam-authority/steami"
	"github.com/steam-authority/steam-authority/storage"
)

const maxBytesToStore = 1024 * 10
const defaultPlayerAvatar = "/assets/img/no-player-image.jpg"

var (
	ErrInvalidPlayerID   = errors.New("invalid id")
	ErrInvalidPlayerName = errors.New("invalid name")
)

type Player struct {
	CreatedAt        time.Time `datastore:"created_at"`               //
	UpdatedAt        time.Time `datastore:"updated_at"`               //
	FriendsAddedAt   time.Time `datastore:"friends_added_at,noindex"` //
	PlayerID         int64     `datastore:"player_id"`                //
	VanintyURL       string    `datastore:"vanity_url"`               //
	Avatar           string    `datastore:"avatar,noindex"`           //
	PersonaName      string    `datastore:"persona_name,noindex"`     //
	RealName         string    `datastore:"real_name,noindex"`        //
	CountryCode      string    `datastore:"country_code"`             //
	StateCode        string    `datastore:"status_code,noindex"`      //
	Level            int       `datastore:"level"`                    //
	GamesRecent      string    `datastore:"games_recent,noindex"`     // JSON
	GamesCount       int       `datastore:"games_count"`              //
	GameStats        string    `datastore:"game_stats,noindex"`       // JSON
	GameHeatMap      string    `datastore:"games_heat_map,noindex"`   // JSON
	Badges           string    `datastore:"badges,noindex"`           // JSON
	BadgesCount      int       `datastore:"badges_count"`             //
	BadgeStats       string    `datastore:"badge_stats,noindex"`      // JSON
	PlayTime         int       `datastore:"play_time"`                //
	TimeCreated      time.Time `datastore:"time_created"`             //
	LastLogOff       time.Time `datastore:"time_logged_off,noindex"`  //
	PrimaryClanID    int       `datastore:"primary_clan_id,noindex"`  //
	Friends          string    `datastore:"friends,noindex"`          // JSON
	FriendsCount     int       `datastore:"friends_count"`            //
	Donated          int       `datastore:"donated"`                  //
	Bans             string    `datastore:"bans,noindex"`             // JSON
	NumberOfVACBans  int       `datastore:"bans_cav"`                 //
	NumberOfGameBans int       `datastore:"bans_game"`                //
	Groups           []int     `datastore:"groups,noindex"`           //
}

func (p Player) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindPlayer, strconv.FormatInt(p.PlayerID, 10), nil)
}

func (p Player) GetPath() string {

	return getPlayerPath(p.PlayerID, p.GetName())
}

func (p Player) GetName() string {
	return getPlayerName(p.PlayerID, p.PersonaName)
}

func (p Player) GetSteamTimeUnix() (int64) {
	return p.TimeCreated.Unix()
}

func (p Player) GetSteamTimeNice() (string) {
	return p.TimeCreated.Format(helpers.DateYear)
}

func (p Player) GetLogoffUnix() (int64) {
	return p.LastLogOff.Unix()
}

func (p Player) GetLogoffNice() (string) {
	return p.LastLogOff.Format(helpers.DateYearTime)
}

func (p Player) GetUpdatedUnix() (int64) {
	return p.UpdatedAt.Unix()
}

func (p Player) GetUpdatedNice() (string) {
	return p.UpdatedAt.Format(helpers.DateTime)
}

func (p Player) GetSteamCommunityLink() string {
	return "https://steamcommunity.com/profiles/" + strconv.FormatInt(p.PlayerID, 10)
}

func (p Player) GetGameHeatMap() string {

	if p.GameHeatMap == "" {
		return "[]"
	}
	return p.GameHeatMap
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

func (p Player) GetAllPlayerApps(sort string, limit int) (apps []PlayerApp, err error) { // todo, only return keys!! (Use ParsePlayerAppKey func)

	if p.GamesCount == 0 {
		return apps, err
	}

	client, ctx, err := GetDSClient()
	if err != nil {
		return apps, err
	}

	q := datastore.NewQuery(KindPlayerApp).Filter("player_id =", p.PlayerID).Order(sort)

	if limit > 0 {
		q.Limit(limit)
	}

	_, err = client.GetAll(ctx, q, &apps)
	return apps, err
}

func (p Player) GetBadges() (badges []ProfileBadge, err error) {

	if p.Badges == "" {
		return
	}

	var bytes []byte

	if storage.IsStorageLocaion(p.Badges) {
		bytes, err = storage.Download(storage.PathBadges(p.PlayerID))
		if err != nil {
			return badges, err
		}
	} else {
		bytes = []byte(p.Badges)
	}

	err = helpers.Unmarshal(bytes, &badges)
	return badges, err
}

func (p Player) GetBadgeStats() (stats ProfileBadgeStats, err error) {

	err = helpers.Unmarshal([]byte(p.BadgeStats), &stats)
	return stats, err
}

func (p Player) GetFriends() (friends []ProfileFriend, err error) {

	if p.Friends == "" {
		return
	}

	var bytes []byte

	if storage.IsStorageLocaion(p.Friends) {
		bytes, err = storage.Download(storage.PathFriends(p.PlayerID))
		if err != nil {
			return friends, err
		}
	} else {
		bytes = []byte(p.Friends)
	}

	err = helpers.Unmarshal(bytes, &friends)
	return friends, err
}

func (p Player) GetRecentGames() (games []steam.RecentlyPlayedGame, err error) {

	if p.GamesRecent == "" {
		return
	}

	var bytes []byte

	if storage.IsStorageLocaion(p.GamesRecent) {
		bytes, err = storage.Download(storage.PathRecentGames(p.PlayerID))
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

func (p Player) GetGameStats() (stats PlayerAppStatsTemplate, err error) {

	err = helpers.Unmarshal([]byte(p.GameStats), &stats)
	return stats, err
}

func (p Player) ShouldUpdateFriends() bool {
	return p.FriendsAddedAt.Unix() < (time.Now().Unix() - int64(60*60*24*30))
}

func (p Player) GetTimeShort() (ret string) {
	return helpers.GetTimeShort(p.PlayTime, 2)
}

func (p Player) GetTimeLong() (ret string) {
	return helpers.GetTimeLong(p.PlayTime, 5)
}

func (p *Player) Update(userAgent string) (errs []error) {

	if !IsValidPlayerID(p.PlayerID) {
		return []error{ErrInvalidPlayerID}
	}

	if helpers.IsBot(userAgent) {
		return []error{}
	}

	if p.UpdatedAt.Unix() > (time.Now().Unix() - int64(60*60*24)) { // 1 Day
		return []error{}
	}

	// Get summary
	err := p.updateSummary()
	if err != nil {
		return []error{err}
	}

	// Async the rest
	var wg sync.WaitGroup

	// Get games
	wg.Add(1)
	go func(p *Player) {
		err = p.updateGames()
		if err != nil {
			errs = append(errs, err)
		}
		wg.Done()
	}(p)

	// Get recent games
	wg.Add(1)
	go func(p *Player) {
		err = p.updateRecentGames()
		if err != nil {
			errs = append(errs, err)
		}
		wg.Done()
	}(p)

	// Get badges
	wg.Add(1)
	go func(p *Player) {
		err = p.updateBadges()
		if err != nil {
			errs = append(errs, err)
		}
		wg.Done()
	}(p)

	// Get friends
	wg.Add(1)
	go func(p *Player) {
		err = p.updateFriends()
		if err != nil {
			errs = append(errs, err)
		}
		wg.Done()
	}(p)

	// Get level
	wg.Add(1)
	go func(p *Player) {
		err = p.updateLevel()
		if err != nil {
			errs = append(errs, err)
		}
		wg.Done()
	}(p)

	// Get bans
	wg.Add(1)
	go func(p *Player) {
		err = p.updateBans()
		if err != nil {
			errs = append(errs, err)
		}
		wg.Done()
	}(p)

	// Get groups
	wg.Add(1)
	go func(p *Player) {
		err = p.updateGroups()
		if err != nil {
			errs = append(errs, err)
		}
		wg.Done()
	}(p)

	// Wait
	wg.Wait()

	err = p.Save()
	if err != nil {
		errs = append(errs, err)
	}

	return errs
}

func (p *Player) updateSummary() (error) {

	summary, _, err := steami.Steam().GetPlayer(p.PlayerID)
	if err != nil {
		return err
	}

	p.Avatar = strings.Replace(summary.AvatarFull, "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/", "", 1)
	p.VanintyURL = path.Base(summary.ProfileURL)
	p.RealName = summary.RealName
	p.CountryCode = summary.LOCCountryCode
	p.StateCode = summary.LOCStateCode
	p.PersonaName = summary.PersonaName
	p.TimeCreated = time.Unix(summary.TimeCreated, 0)
	p.LastLogOff = time.Unix(summary.LastLogOff, 0)
	p.PrimaryClanID = summary.PrimaryClanID

	return err
}

func (p *Player) updateGames() (error) {

	resp, _, err := steami.Steam().GetOwnedGames(p.PlayerID)
	if err != nil {
		return err
	}

	// Loop apps
	var appsMap = map[int]*PlayerApp{}
	var appIDs []int
	var playtime = 0
	for _, v := range resp.Games {
		playtime = playtime + v.PlaytimeForever
		appIDs = append(appIDs, v.AppID)
		appsMap[v.AppID] = &PlayerApp{
			PlayerID:     p.PlayerID,
			AppID:        v.AppID,
			AppName:      v.Name,
			AppIcon:      v.ImgIconURL,
			AppTime:      v.PlaytimeForever,
			AppPrice:     0,
			AppPriceHour: 0,
		}
	}

	// Save data to player
	p.GamesCount = len(resp.Games)
	p.PlayTime = playtime

	// Go get price info from MySQL
	gamesSQL, err := GetAppsByID(appIDs, []string{"id", "price_final"})
	if err != nil {
		return err
	}

	for _, v := range gamesSQL {
		if v.PriceFinal > 0 {
			appsMap[v.ID].AppPrice = v.PriceFinal
			appsMap[v.ID].SetPriceHour()
		}
	}

	// Convert to slice
	var appsSlice []Kind
	for _, v := range appsMap {
		appsSlice = append(appsSlice, *v)
	}

	err = BulkSaveKinds(appsSlice, KindPlayerApp)
	if err != nil {
		return err
	}

	// Make stats
	var gameStats = PlayerAppStatsTemplate{}
	for _, v := range appsMap {

		gameStats.All.AddApp(*v)
		if v.AppTime > 0 {
			gameStats.Played.AddApp(*v)
		}
	}

	bytes, err := json.Marshal(gameStats)
	if err != nil {
		return err
	}

	p.GameStats = string(bytes)

	// Make heatmap
	var roundedPrices []int
	var maxPrice int
	for _, v := range appsMap {

		var roundedPrice = int(math.Floor(float64(v.AppPrice)/500) * 5) // Round down to nearest 5

		roundedPrices = append(roundedPrices, roundedPrice)

		maxPrice = int(math.Max(float64(roundedPrice), float64(maxPrice)))
	}

	ret := make([][]int, (maxPrice/5)+1)
	for i := 0; i <= maxPrice/5; i++ {
		ret[i] = []int{0, 0}
	}
	for _, v := range roundedPrices {
		ret[(v / 5)] = []int{0, ret[(v / 5)][1] + 1}
	}

	bytes, err = json.Marshal(ret)
	if err != nil {
		return err
	}

	p.GameHeatMap = string(bytes)

	return nil
}

func (p *Player) updateRecentGames() (error) {

	recentResponse, _, err := steami.Steam().GetRecentlyPlayedGames(p.PlayerID)
	if err != nil {
		return err
	}

	// Encode to JSON bytes
	bytes, err := json.Marshal(recentResponse.Games)
	if err != nil {
		return err
	}

	// Upload
	if len(bytes) > maxBytesToStore {
		storagePath := storage.PathRecentGames(p.PlayerID)
		err = storage.Upload(storagePath, bytes, false)
		if err != nil {
			return err
		}
		p.GamesRecent = storagePath
	} else {
		p.GamesRecent = string(bytes)
	}

	return nil
}

func (p *Player) updateBadges() (error) {

	response, _, err := steami.Steam().GetBadges(p.PlayerID)
	if err != nil {
		return err
	}

	// Save count
	p.BadgesCount = len(response.Badges)

	// Save stats
	stats := ProfileBadgeStats{
		PlayerXP:                   response.PlayerXP,
		PlayerLevel:                response.PlayerLevel,
		PlayerXPNeededToLevelUp:    response.PlayerXPNeededToLevelUp,
		PlayerXPNeededCurrentLevel: response.PlayerXPNeededCurrentLevel,
		PercentOfLevel:             response.GetPercentOfLevel(),
	}

	bytes, err := json.Marshal(stats)
	if err != nil {
		return err
	}

	p.BadgeStats = string(bytes)

	// Start badges slice
	var badgeSlice []ProfileBadge
	var appIDSlice []int
	for _, v := range response.Badges {
		appIDSlice = append(appIDSlice, v.AppID)
		badgeSlice = append(badgeSlice, ProfileBadge{
			BadgeID:        v.BadgeID,
			AppID:          v.AppID,
			Level:          v.Level,
			CompletionTime: v.CompletionTime,
			XP:             v.XP,
			Scarcity:       v.Scarcity,
		})
	}
	appIDSlice = helpers.Unique(appIDSlice)

	// Make map of app rows
	var appRowsMap = map[int]App{}
	appRows, err := GetAppsByID(appIDSlice, []string{"id", "name", "icon"})
	for _, v := range appRows {
		appRowsMap[v.ID] = v
	}

	// Finish badges slice
	for k, v := range badgeSlice {
		if app, ok := appRowsMap[v.AppID]; ok {
			badgeSlice[k].AppName = app.GetName()
			badgeSlice[k].AppIcon = app.GetIcon()
		}
	}

	// Encode to JSON bytes
	bytes, err = json.Marshal(badgeSlice)
	if err != nil {
		return err
	}

	// Upload
	if len(bytes) > maxBytesToStore {
		storagePath := storage.PathBadges(p.PlayerID)
		err = storage.Upload(storagePath, bytes, false)
		if err != nil {
			return err
		}
		p.Badges = storagePath
	} else {
		p.Badges = string(bytes)
	}

	return nil
}

func (p *Player) updateFriends() (error) {

	resp, _, err := steami.Steam().GetFriendList(p.PlayerID)
	if err != nil {
		return err
	}

	p.FriendsCount = len(resp.Friends)

	// Make friend ID slice & map
	var friendsMap = map[int64]*ProfileFriend{}
	var friendsSlice []int64
	for _, v := range resp.Friends {

		friendsSlice = append(friendsSlice, v.SteamID)

		friendsMap[v.SteamID] = &ProfileFriend{
			SteamID:      v.SteamID,
			Relationship: v.Relationship,
			FriendSince:  v.FriendSince,
		}
	}

	// Get friends from DS
	friendRows, err := GetPlayersByIDs(friendsSlice)
	if err != nil {
		return err
	}

	// Fill in the map
	for _, v := range friendRows {
		if v.PlayerID != 0 {

			friendsMap[v.PlayerID].Avatar = v.GetAvatar()
			friendsMap[v.PlayerID].Games = v.GamesCount
			friendsMap[v.PlayerID].Name = v.GetName()
			friendsMap[v.PlayerID].Level = v.Level
			friendsMap[v.PlayerID].LoggedOff = v.GetLogoffUnix()

		}
	}

	// Make into map again, so it can be unmarshalled

	var friends []ProfileFriend
	for _, v := range friendsMap {
		friends = append(friends, *v)
	}

	// Encode to JSON bytes
	bytes, err := json.Marshal(friends)
	if err != nil {
		return err
	}

	// Upload
	if len(bytes) > maxBytesToStore {
		storagePath := storage.PathFriends(p.PlayerID)
		err = storage.Upload(storagePath, bytes, false)
		if err != nil {
			return err
		}
		p.Friends = storagePath
	} else {
		p.Friends = string(bytes)
	}

	return nil
}

func (p *Player) updateLevel() (error) {

	level, _, err := steami.Steam().GetSteamLevel(p.PlayerID)
	if err != nil {
		return err
	}

	p.Level = level

	return nil
}

func (p *Player) updateBans() (error) {

	bans, _, err := steami.Steam().GetPlayerBans(p.PlayerID)
	if err == steam.ErrNoUserFound {
		return nil
	} else if err != nil {
		return err
	}

	p.NumberOfGameBans = bans.NumberOfGameBans
	p.NumberOfVACBans = bans.NumberOfVACBans

	// Encode to JSON bytes
	bytes, err := json.Marshal(bans)
	if err != nil {
		return err
	}

	p.Bans = string(bytes)

	return nil
}

func (p *Player) updateGroups() (error) {

	resp, _, err := steami.Steam().GetUserGroupList(p.PlayerID)
	if err != nil {
		return err
	}

	p.Groups = resp.GetIDs()

	return nil
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

	_, err = SaveKind(p.GetKey(), p)
	if err != nil {
		return err
	}

	return nil
}

func getPlayerPath(id int64, name string) string {

	p := "/players/" + strconv.FormatInt(id, 10)
	if name != "" {
		p = p + "/" + slug.Make(name)
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

// todo, check this is acurate
func IsValidPlayerID(id int64) bool {

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

	q := datastore.NewQuery(KindPlayer).Filter("vanity_url =", name).Limit(1)

	var players []Player

	_, err = client.GetAll(ctx, q, &players)

	err = checkForMissingPlayerFields(err)
	if err != nil {
		return player, err
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

	chunks := chunkKeys(keys, 0)
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

		} else if err.Error() == ErrNoSuchEntity.Error() {

			return nil

		} else {

			return err

		}
	}

	return nil
}

func CountPlayers() (count int, err error) {

	return memcache.GetSetInt(memcache.CountPlayers, func() (count int, err error) {

		client, ctx, err := GetDSClient()
		if err != nil {
			return count, err
		}

		q := datastore.NewQuery(KindPlayer)
		count, err = client.Count(ctx, q)
		return count, err
	})
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
	Price     int
	PriceHour float64
	Time      int
}

func (p *playerAppStatsInnerTemplate) AddApp(app PlayerApp) {
	p.Count++
	p.Price = p.Price + app.AppPrice
	p.PriceHour = p.PriceHour + app.AppPriceHour
	p.Time = p.Time + app.AppTime
}

func (p playerAppStatsInnerTemplate) GetAveragePrice() float64 {
	return helpers.DollarsFloat(float64(p.Price) / float64(p.Count))
}

func (p playerAppStatsInnerTemplate) GetTotalPrice() float64 {
	return helpers.DollarsFloat(float64(p.Price))
}

func (p playerAppStatsInnerTemplate) GetAveragePriceHour() float64 {
	return helpers.DollarsFloat(p.PriceHour / float64(p.Count))
}
func (p playerAppStatsInnerTemplate) GetAverageTime() string {
	return helpers.GetTimeShort(int(float64(p.Time)/float64(p.Count)), 2)
}

func (p playerAppStatsInnerTemplate) GetTotalTime() string {
	return helpers.GetTimeShort(p.Time, 2)
}

// ProfileFriend
type ProfileFriend struct {
	SteamID      int64  `json:"id"`
	Relationship string `json:"rs"`
	FriendSince  int64  `json:"fs"`
	Avatar       string `json:"ic"`
	Name         string `json:"nm"`
	Games        int    `json:"gm"`
	Level        int    `json:"lv"`
	LoggedOff    int64  `json:"lo"`
}

func (p ProfileFriend) Scanned() bool {
	return p.LoggedOff > 0
}

func (p ProfileFriend) GetPath() string {
	return getPlayerPath(p.SteamID, p.Name)
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
	if p.Scanned() {
		return time.Unix(p.FriendSince, 0).Format(helpers.DateYearTime)
	}
	return "-"
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

func (p ProfileBadge) GetTimeFormatted() (string) {
	return time.Unix(p.CompletionTime, 0).Format(helpers.DateYearTime)
}

func (p ProfileBadge) GetAppPath() (string) {
	return getAppPath(p.AppID, p.AppName)
}

func (p ProfileBadge) GetAppName() (string) {
	if p.AppID == 0 {
		return "No App"
	}
	return getAppName(p.AppID, p.AppName)
}

func (p ProfileBadge) GetIcon() (string) {
	if p.AppIcon == "" {
		return "/assets/img/no-app-image-square.jpg"
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
