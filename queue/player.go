package queue

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/websockets"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
)

type playerMessage struct {
	ID              int64                    `json:"id"`
	PICSProfileInfo RabbitMessageProfilePICS `` // Leave JSON key name as default
}

type playerQueue struct {
	baseQueue
}

func (q playerQueue) processMessage(msg amqp.Delivery) {

	var err error
	var payload = baseMessage{
		Message: playerMessage{},
	}

	err = helpers.Unmarshal(msg.Body, &payload)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	var message playerMessage
	err = mapstructure.Decode(payload.Message, &message)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	if payload.Attempt > 1 {
		logInfo("Consuming player " + strconv.FormatInt(message.ID, 10) + ", attempt " + strconv.Itoa(payload.Attempt))
	}

	if !message.PICSProfileInfo.SteamID.IsValid {
		logError(errors.New("not valid account id"))
		payload.ack(msg)
		return
	}

	if !message.PICSProfileInfo.SteamID.IsIndividualAccount {
		logError(errors.New("not individual account id"))
		payload.ack(msg)
		return
	}

	// Convert steamID3 to steamID64
	id64, err := helpers.GetSteam().GetID(strconv.Itoa(message.PICSProfileInfo.SteamID.AccountID))
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	// Update player
	player, err := db.GetPlayer(id64)
	err = helpers.IgnoreErrors(err, datastore.ErrNoSuchEntity)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	player.PlayerID = id64
	player.RealName = message.PICSProfileInfo.RealName
	player.StateCode = message.PICSProfileInfo.StateName
	player.CountryCode = message.PICSProfileInfo.CountryName

	// Get summary
	err = updatePlayerSummary(&player)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	err = updatePlayerGames(&player)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	err = updatePlayerRecentGames(&player)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	err = updatePlayerBadges(&player)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	err = updatePlayerFriends(&player)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	err = updatePlayerLevel(&player)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	err = updatePlayerBans(&player)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	err = updatePlayerGroups(&player)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	err = db.CreateEvent(new(http.Request), player.PlayerID, db.EventRefresh)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	err = player.Save()
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	// Send websocket
	page, err := websockets.GetPage(websockets.PageProfile)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	if page.HasConnections() {
		page.Send(strconv.FormatInt(player.PlayerID, 10))
	}

	payload.ack(msg)
}

type RabbitMessageProfilePICS struct {
	Result  int `json:"Result"`
	SteamID struct {
		IsBlankAnonAccount            bool `json:"IsBlankAnonAccount"`
		IsGameServerAccount           bool `json:"IsGameServerAccount"`
		IsPersistentGameServerAccount bool `json:"IsPersistentGameServerAccount"`
		IsAnonGameServerAccount       bool `json:"IsAnonGameServerAccount"`
		IsContentServerAccount        bool `json:"IsContentServerAccount"`
		IsClanAccount                 bool `json:"IsClanAccount"`
		IsChatAccount                 bool `json:"IsChatAccount"`
		IsLobby                       bool `json:"IsLobby"`
		IsIndividualAccount           bool `json:"IsIndividualAccount"`
		IsAnonAccount                 bool `json:"IsAnonAccount"`
		IsAnonUserAccount             bool `json:"IsAnonUserAccount"`
		IsConsoleUserAccount          bool `json:"IsConsoleUserAccount"`
		IsValid                       bool `json:"IsValid"`
		AccountID                     int  `json:"AccountID"` // steamID3
		AccountInstance               int  `json:"AccountInstance"`
		AccountType                   int  `json:"AccountType"`
		AccountUniverse               int  `json:"AccountUniverse"`
	} `json:"SteamID"`
	TimeCreated time.Time   `json:"TimeCreated"`
	RealName    string      `json:"RealName"`
	CityName    string      `json:"CityName"`
	StateName   string      `json:"StateName"`
	CountryName string      `json:"CountryName"`
	Headline    string      `json:"Headline"`
	Summary     string      `json:"Summary"`
	JobID       steamKitJob `json:"JobID"`
}

func updatePlayerSummary(player *db.Player) error {

	summary, _, err := helpers.GetSteam().GetPlayer(player.PlayerID)
	if err != nil {
		return err
	}

	player.Avatar = strings.Replace(summary.AvatarFull, "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/", "", 1)
	player.VanintyURL = path.Base(summary.ProfileURL)
	player.RealName = summary.RealName
	player.CountryCode = summary.LOCCountryCode
	player.StateCode = summary.LOCStateCode
	player.PersonaName = summary.PersonaName
	player.TimeCreated = time.Unix(summary.TimeCreated, 0)
	player.LastLogOff = time.Unix(summary.LastLogOff, 0)
	player.PrimaryClanID = int(summary.PrimaryClanID)

	return err
}

func updatePlayerGames(player *db.Player) error {

	// Grab games from Steam
	resp, _, err := helpers.GetSteam().GetOwnedGames(player.PlayerID)
	if err != nil {
		return err
	}

	// Save count
	player.GamesCount = len(resp.Games)

	// Start creating PlayerApp's
	var playerApps = map[int]*db.PlayerApp{}
	var appPrices = map[int]map[string]int{}
	var appPriceHour = map[int]map[string]float64{}
	var appIDs []int
	var playtime = 0
	for _, v := range resp.Games {
		playtime = playtime + v.PlaytimeForever
		appIDs = append(appIDs, v.AppID)
		playerApps[v.AppID] = &db.PlayerApp{
			PlayerID: player.PlayerID,
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

	// Getting missing price info from MySQL
	gameRows, err := db.GetAppsByID(appIDs, []string{"id", "prices"})
	if err != nil {
		return err
	}

	for _, v := range gameRows {

		prices, err := v.GetPrices()
		if err != nil {
			logError(err)
			continue
		}

		for code, vv := range prices {

			appPrices[v.ID][string(code)] = vv.Final
			if appPrices[v.ID][string(code)] > 0 && playerApps[v.ID].AppTime == 0 {
				appPriceHour[v.ID][string(code)] = -1
			} else if appPrices[v.ID][string(code)] > 0 && playerApps[v.ID].AppTime > 0 {
				appPriceHour[v.ID][string(code)] = (float64(appPrices[v.ID][string(code)]) / 100) / (float64(playerApps[v.ID].AppTime) / 60)
			} else {
				appPriceHour[v.ID][string(code)] = 0
			}
		}

		//
		err = mapstructure.Decode(appPrices[v.ID], &playerApps[v.ID].AppPrices)
		logError(err)

		//
		err = mapstructure.Decode(appPriceHour[v.ID], &playerApps[v.ID].AppPriceHour)
		logError(err)
	}

	// Save playerApps to Datastore
	var appsSlice []db.Kind
	for _, v := range playerApps {
		appsSlice = append(appsSlice, *v)
	}

	err = db.BulkSaveKinds(appsSlice, db.KindPlayerApp, true)
	if err != nil {
		return err
	}

	// Save stats to player
	var gameStats = db.PlayerAppStatsTemplate{}
	for _, v := range playerApps {

		gameStats.All.AddApp(v.AppTime, appPrices[v.AppID], appPriceHour[v.AppID])
		if v.AppTime > 0 {
			gameStats.Played.AddApp(v.AppTime, appPrices[v.AppID], appPriceHour[v.AppID])
		}
	}

	b, err := json.Marshal(gameStats)
	if err != nil {
		return err
	}

	player.GameStats = string(b)

	// Make heatmap
	// var roundedPrices []int
	// var maxPrice int
	// for _, v := range playerApps {
	//
	// 	var roundedPrice = int(math.Floor(float64(v.AppPrice)/500) * 5) // Round down to nearest 5
	//
	// 	roundedPrices = append(roundedPrices, roundedPrice)
	//
	// 	maxPrice = int(math.Max(float64(roundedPrice), float64(maxPrice)))
	// }
	//
	// ret := make([][]int, (maxPrice/5)+1)
	// for i := 0; i <= maxPrice/5; i++ {
	// 	ret[i] = []int{0, 0}
	// }
	// for _, v := range roundedPrices {
	// 	ret[(v / 5)] = []int{0, ret[(v / 5)][1] + 1}
	// }
	//
	// b, err = json.Marshal(ret)
	// if err != nil {
	// 	return err
	// }
	//
	// player.GameHeatMap = string(b)

	return nil
}

func updatePlayerRecentGames(player *db.Player) error {

	recentResponse, _, err := helpers.GetSteam().GetRecentlyPlayedGames(player.PlayerID)
	if err != nil {
		return err
	}

	var games []db.ProfileRecentGame
	for _, v := range recentResponse.Games {
		games = append(games, db.ProfileRecentGame{
			AppID:           v.AppID,
			Name:            v.Name,
			PlayTime2Weeks:  v.PlayTime2Weeks,
			PlayTimeForever: v.PlayTimeForever,
			ImgIconURL:      v.ImgIconURL,
			ImgLogoURL:      v.ImgLogoURL,
		})
	}

	b, err := json.Marshal(games)
	if err != nil {
		return err
	}

	// Upload
	if len(b) > maxBytesToStore {
		storagePath := helpers.PathRecentGames(player.PlayerID)
		err = helpers.Upload(storagePath, b)
		if err != nil {
			return err
		}
		player.GamesRecent = storagePath
	} else {
		player.GamesRecent = string(b)
	}

	return nil
}

func updatePlayerBadges(player *db.Player) error {

	response, _, err := helpers.GetSteam().GetBadges(player.PlayerID)
	if err != nil {
		return err
	}

	// Save count
	player.BadgesCount = len(response.Badges)

	// Save stats
	stats := db.ProfileBadgeStats{
		PlayerXP:                   response.PlayerXP,
		PlayerLevel:                response.PlayerLevel,
		PlayerXPNeededToLevelUp:    response.PlayerXPNeededToLevelUp,
		PlayerXPNeededCurrentLevel: response.PlayerXPNeededCurrentLevel,
		PercentOfLevel:             response.GetPercentOfLevel(),
	}

	b, err := json.Marshal(stats)
	if err != nil {
		return err
	}

	player.BadgeStats = string(b)

	// Start badges slice
	var badgeSlice []db.ProfileBadge
	var appIDSlice []int
	for _, v := range response.Badges {
		appIDSlice = append(appIDSlice, v.AppID)
		badgeSlice = append(badgeSlice, db.ProfileBadge{
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
	var appRowsMap = map[int]db.App{}
	appRows, err := db.GetAppsByID(appIDSlice, []string{"id", "name", "icon"})
	logError(err)

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
	b, err = json.Marshal(badgeSlice)
	if err != nil {
		return err
	}

	// Upload
	if len(b) > maxBytesToStore {
		storagePath := helpers.PathBadges(player.PlayerID)
		err = helpers.Upload(storagePath, b)
		if err != nil {
			return err
		}
		player.Badges = storagePath
	} else {
		player.Badges = string(b)
	}

	return nil
}

func updatePlayerFriends(player *db.Player) error {

	resp, _, err := helpers.GetSteam().GetFriendList(player.PlayerID)

	// This endpoint seems to error if the player is private, so it's probably fine.
	err2, ok := err.(steam.Error)
	if ok && (err2.Code == 401) {
		return nil
	}
	if err != nil {
		return err
	}

	player.FriendsCount = len(resp.Friends)

	// Make friend ID slice & map
	var friendsMap = map[int64]*db.ProfileFriend{}
	var friendsSlice []int64
	for _, v := range resp.Friends {

		friendsSlice = append(friendsSlice, int64(v.SteamID))

		friendsMap[int64(v.SteamID)] = &db.ProfileFriend{
			SteamID:     int64(v.SteamID),
			FriendSince: v.FriendSince,
		}
	}

	// Get friends from DS
	friendRows, err := db.GetPlayersByIDs(friendsSlice)
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

	var friends []db.ProfileFriend
	for _, v := range friendsMap {
		friends = append(friends, *v)
	}

	// Encode to JSON bytes
	b, err := json.Marshal(friends)
	if err != nil {
		return err
	}

	// Upload
	if len(b) > maxBytesToStore {
		storagePath := helpers.PathFriends(player.PlayerID)
		err = helpers.Upload(storagePath, b)
		if err != nil {
			return err
		}
		player.Friends = storagePath
	} else {
		player.Friends = string(b)
	}

	return nil
}

func updatePlayerLevel(player *db.Player) error {

	level, _, err := helpers.GetSteam().GetSteamLevel(player.PlayerID)
	if err != nil {
		return err
	}

	player.Level = level

	return nil
}

func updatePlayerBans(player *db.Player) error {

	response, _, err := helpers.GetSteam().GetPlayerBans(player.PlayerID)
	if err == steam.ErrNoUserFound {
		return nil
	} else if err != nil {
		return err
	}

	player.NumberOfGameBans = response.NumberOfGameBans
	player.NumberOfVACBans = response.NumberOfVACBans

	var bans db.PlayerBans
	bans.CommunityBanned = response.CommunityBanned
	bans.VACBanned = response.VACBanned
	bans.NumberOfVACBans = response.NumberOfVACBans
	bans.DaysSinceLastBan = response.DaysSinceLastBan
	bans.NumberOfGameBans = response.NumberOfGameBans
	bans.EconomyBan = response.EconomyBan

	// Encode to JSON bytes
	b, err := json.Marshal(bans)
	if err != nil {
		return err
	}

	player.Bans = string(b)

	return nil
}

func updatePlayerGroups(player *db.Player) error {

	resp, _, err := helpers.GetSteam().GetUserGroupList(player.PlayerID)

	// This endpoint seems to error if the player is private, so it's probably fine.
	err2, ok := err.(steam.Error)
	if ok && (err2.Code == 403) {
		return nil
	}
	if err != nil {
		return err
	}

	player.Groups = resp.GetIDs()

	return nil
}
