package queue

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/mongo"
	"github.com/gamedb/website/websockets"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
)

type playerMessage struct {
	ID              int64                    `json:"id"`
	PICSProfileInfo RabbitMessageProfilePICS `json:"PICSProfileInfo"`
}

type playerQueue struct {
	baseQueue
}

func (q playerQueue) processMessages(msgs []amqp.Delivery) {

	msg := msgs[0]

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
		logError(errors.New("not valid player id: " + strconv.FormatInt(message.ID, 10)))
		payload.ack(msg)
		return
	}

	if !message.PICSProfileInfo.SteamID.IsIndividualAccount {
		logError(errors.New("not individual account id: " + strconv.FormatInt(message.ID, 10)))
		payload.ack(msg)
		return
	}

	// Convert steamID3 to steamID64
	id64, err := helpers.GetSteam().GetID(strconv.Itoa(message.PICSProfileInfo.SteamID.AccountID))
	if err != nil {
		logError(err, message.ID)
		payload.ack(msg)
		return
	}

	// Update player
	player := mongo.Player{}
	player.ID = id64
	player.RealName = message.PICSProfileInfo.RealName
	player.StateCode = message.PICSProfileInfo.StateName
	player.CountryCode = message.PICSProfileInfo.CountryName
	player.UpdatedAt = time.Now()

	// Get summary
	err = updatePlayerSummary(&player)
	if err != nil {

		logError(err, message.ID)

		if err == steam.ErrNoUserFound {
			payload.ack(msg)
		} else {
			payload.ackRetry(msg)
		}

		return
	}

	err = updatePlayerGames(&player)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	err = updatePlayerRecentGames(&player)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	err = updatePlayerBadges(&player)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	err = updatePlayerFriends(&player)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	err = updatePlayerLevel(&player)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	err = updatePlayerBans(&player)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	err = updatePlayerGroups(&player)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	// err = savePlayerToBuffer(player)
	// if err != nil {
	// 	logError(err, message.ID)
	// 	payload.ackRetry(msg)
	// 	return
	// }

	err = savePlayerToInflux(player)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	err = savePlayerMongo(player)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	err = mongo.CreateEvent(new(http.Request), player.ID, mongo.EventRefresh)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	// Send websocket
	page, err := websockets.GetPage(websockets.PageProfile)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	if page.HasConnections() {
		page.Send(strconv.FormatInt(player.ID, 10))
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
	TimeCreated string      `json:"TimeCreated"`
	RealName    string      `json:"RealName"`
	CityName    string      `json:"CityName"`
	StateName   string      `json:"StateName"`
	CountryName string      `json:"CountryName"`
	Headline    string      `json:"Headline"`
	Summary     string      `json:"Summary"`
	JobID       steamKitJob `json:"JobID"`
}

func updatePlayerSummary(player *mongo.Player) error {

	summary, _, err := helpers.GetSteam().GetPlayer(player.ID)
	if err != nil {
		return err
	}

	// Avatar
	var avatarBase = "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/"
	if summary.AvatarFull != "" && helpers.GetResponseCode(avatarBase+summary.AvatarFull) == 200 {
		player.Avatar = summary.AvatarFull
	} else {
		player.Avatar = ""
	}

	//
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

func updatePlayerGames(player *mongo.Player) error {

	// Grab games from Steam
	// resp, _, err := helpers.GetSteam().GetOwnedGames(player.ID)
	// if err != nil {
	// 	return err
	// }
	//
	// // Save count
	// player.GamesCount = len(resp.Games)
	//
	// // Start creating PlayerApp's
	// var playerApps = map[int]*db.PlayerApp{}
	// var appPrices = map[int]map[string]int{}
	// var appPriceHour = map[int]map[string]float64{}
	// var appIDs []int
	// var playtime = 0
	// for _, v := range resp.Games {
	// 	playtime = playtime + v.PlaytimeForever
	// 	appIDs = append(appIDs, v.AppID)
	// 	playerApps[v.AppID] = &db.PlayerApp{
	// 		PlayerID: player.ID,
	// 		AppID:    v.AppID,
	// 		AppName:  v.Name,
	// 		AppIcon:  v.ImgIconURL,
	// 		AppTime:  v.PlaytimeForever,
	// 	}
	// 	appPrices[v.AppID] = map[string]int{}
	// 	appPriceHour[v.AppID] = map[string]float64{}
	// }
	//
	// // Save playtime
	// player.PlayTime = playtime
	//
	// // Getting missing price info from MySQL
	// gameRows, err := db.GetAppsByID(appIDs, []string{"id", "prices"})
	// if err != nil {
	// 	return err
	// }
	//
	// for _, v := range gameRows {
	//
	// 	prices, err := v.GetPrices()
	// 	if err != nil {
	// 		logError(err)
	// 		continue
	// 	}
	//
	// 	for code, vv := range prices {
	//
	// 		appPrices[v.ID][string(code)] = vv.Final
	// 		if appPrices[v.ID][string(code)] > 0 && playerApps[v.ID].AppTime == 0 {
	// 			appPriceHour[v.ID][string(code)] = -1
	// 		} else if appPrices[v.ID][string(code)] > 0 && playerApps[v.ID].AppTime > 0 {
	// 			appPriceHour[v.ID][string(code)] = (float64(appPrices[v.ID][string(code)]) / 100) / (float64(playerApps[v.ID].AppTime) / 60)
	// 		} else {
	// 			appPriceHour[v.ID][string(code)] = 0
	// 		}
	// 	}
	//
	// 	//
	// 	err = mapstructure.Decode(appPrices[v.ID], &playerApps[v.ID].AppPrices)
	// 	logError(err)
	//
	// 	//
	// 	err = mapstructure.Decode(appPriceHour[v.ID], &playerApps[v.ID].AppPriceHour)
	// 	logError(err)
	// }
	//
	// // Save playerApps to Datastore
	// var appsSlice []db.Kind
	// for _, v := range playerApps {
	// 	appsSlice = append(appsSlice, *v)
	// }
	//
	// err = db.BulkSaveKinds(appsSlice, db.KindPlayerApp, true)
	// if err != nil {
	// 	return err
	// }
	//
	// // Save stats to player
	// var gameStats = mongo.PlayerAppStatsTemplate{}
	// for _, v := range playerApps {
	//
	// 	gameStats.All.AddApp(v.AppTime, appPrices[v.AppID], appPriceHour[v.AppID])
	// 	if v.AppTime > 0 {
	// 		gameStats.Played.AddApp(v.AppTime, appPrices[v.AppID], appPriceHour[v.AppID])
	// 	}
	// }
	//
	// b, err := json.Marshal(gameStats)
	// if err != nil {
	// 	return err
	// }
	//
	// player.GameStats = string(b)

	return nil
}

func updatePlayerRecentGames(player *mongo.Player) error {

	recentResponse, _, err := helpers.GetSteam().GetRecentlyPlayedGames(player.ID)
	if err != nil {
		return err
	}

	var games []mongo.ProfileRecentGame
	for _, v := range recentResponse.Games {
		games = append(games, mongo.ProfileRecentGame{
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
		storagePath := helpers.PathRecentGames(player.ID)
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

func updatePlayerBadges(player *mongo.Player) error {

	response, _, err := helpers.GetSteam().GetBadges(player.ID)
	if err != nil {
		return err
	}

	// Save count
	player.BadgesCount = len(response.Badges)

	// Save stats
	stats := mongo.ProfileBadgeStats{
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
	var badgeSlice []mongo.ProfileBadge
	var appIDSlice []int
	for _, v := range response.Badges {
		appIDSlice = append(appIDSlice, v.AppID)
		badgeSlice = append(badgeSlice, mongo.ProfileBadge{
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
		storagePath := helpers.PathBadges(player.ID)
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

func updatePlayerFriends(player *mongo.Player) error {

	resp, _, err := helpers.GetSteam().GetFriendList(player.ID)

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
	var friendsMap = map[int64]*mongo.ProfileFriend{}
	var friendsSlice []int64
	for _, v := range resp.Friends {

		friendsSlice = append(friendsSlice, int64(v.SteamID))

		friendsMap[int64(v.SteamID)] = &mongo.ProfileFriend{
			SteamID:     int64(v.SteamID),
			FriendSince: v.FriendSince,
		}
	}

	// Get friends from DS
	friendRows, err := mongo.GetPlayersByIDs(friendsSlice)
	if err != nil {
		return err
	}

	// Fill in the map
	for _, friend := range friendRows {
		if friend.ID != 0 {

			friendsMap[friend.ID].Avatar = friend.GetAvatar()
			friendsMap[friend.ID].Games = friend.GamesCount
			friendsMap[friend.ID].Name = friend.GetName()
			friendsMap[friend.ID].Level = friend.Level
			friendsMap[friend.ID].LoggedOff = friend.GetLogoffUnix()

		}
	}

	// Make into map again, so it can be unmarshalled

	var friends []mongo.ProfileFriend
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
		storagePath := helpers.PathFriends(player.ID)
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

func updatePlayerLevel(player *mongo.Player) error {

	level, _, err := helpers.GetSteam().GetSteamLevel(player.ID)
	if err != nil {
		return err
	}

	player.Level = level

	return nil
}

func updatePlayerBans(player *mongo.Player) error {

	response, _, err := helpers.GetSteam().GetPlayerBans(player.ID)
	if err == steam.ErrNoUserFound {
		return nil
	} else if err != nil {
		return err
	}

	player.NumberOfGameBans = response.NumberOfGameBans
	player.NumberOfVACBans = response.NumberOfVACBans

	var bans mongo.PlayerBans
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

func updatePlayerGroups(player *mongo.Player) error {

	resp, _, err := helpers.GetSteam().GetUserGroupList(player.ID)

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

func savePlayerMongo(player mongo.Player) error {

	_, err := mongo.ReplaceDocument(mongo.CollectionPlayers, bson.M{"_id": player.ID}, player)

	return err
}

func savePlayerToInflux(player mongo.Player) (err error) {

	// ranks, err := db.GetRank(player.ID)
	// if err != nil && err != db.ErrNoSuchEntity {
	// 	return err
	// }
	//
	// fields := map[string]interface{}{
	// 	"level":    player.Level,
	// 	"games":    player.GamesCount,
	// 	"badges":   player.BadgesCount,
	// 	"playtime": player.PlayTime,
	// 	"friends":  player.FriendsCount,
	//
	// 	"level_rank":    ranks.LevelRank,
	// 	"games_rank":    ranks.GamesRank,
	// 	"badges_rank":   ranks.BadgesRank,
	// 	"playtime_rank": ranks.PlayTimeRank,
	// 	"friends_rank":  ranks.FriendsRank,
	// }
	//
	// _, err = db.InfluxWrite(db.InfluxRetentionPolicyAllTime, influx.Point{
	// 	Measurement: string(db.InfluxMeasurementPlayers),
	// 	Tags: map[string]string{
	// 		"player_id": strconv.FormatInt(player.ID, 10),
	// 	},
	// 	Fields:    fields,
	// 	Time:      time.Now(),
	// 	Precision: "m",
	// })

	return err
}
