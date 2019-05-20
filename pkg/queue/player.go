package queue

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
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
		Message:       playerMessage{},
		OriginalQueue: queueGoPlayers,
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
		logError(errors.New("not a valid player id: " + strconv.FormatInt(message.ID, 10)))
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
	player, err := mongo.GetPlayer(id64)
	err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
	if err != nil {
		logError(err, message.ID)
		payload.ack(msg)
		return
	}

	player.ID = id64
	player.RealName = message.PICSProfileInfo.RealName
	player.StateCode = message.PICSProfileInfo.StateName
	player.CountryCode = message.PICSProfileInfo.CountryName
	player.UpdatedAt = time.Now()

	//
	var wg sync.WaitGroup

	// Calls to api.steampowered.com
	wg.Add(1)
	go func() {

		defer wg.Done()

		err = updatePlayerSummary(&player)
		if err != nil {

			if err == steam.ErrNoUserFound {
				payload.ack(msg)
			} else {
				logError(err, message.ID)
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
	}()

	// Calls to store.steampowered.com
	wg.Add(1)
	go func() {

		defer wg.Done()

		err = updatePlayerWishlist(&player)
		if err != nil {
			logError(err, message.ID)
			payload.ackRetry(msg)
			return
		}

	}()

	wg.Wait()

	if payload.actionTaken {
		return
	}

	// Write to databases
	wg.Add(1)
	go func() {

		defer wg.Done()

		err = savePlayerMongo(player)
		if err != nil {
			logError(err, message.ID)
			payload.ackRetry(msg)
			return
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		err = savePlayerToInflux(player)
		if err != nil {
			logError(err, message.ID)
			payload.ackRetry(msg)
			return
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		err = mongo.CreatePlayerEvent(&http.Request{}, player.ID, mongo.EventRefresh)
		if err != nil {
			logError(err, message.ID)
			payload.ackRetry(msg)
			return
		}
	}()

	wg.Wait()

	if payload.actionTaken {
		return
	}

	// Other bits
	wg.Add(1)
	go func() {

		defer wg.Done()

		// Send websocket
		wsPayload := websockets.PubSubID64Payload{}
		wsPayload.ID = strconv.FormatInt(player.ID, 10)
		wsPayload.Pages = []websockets.WebsocketPage{websockets.PagePlayer}

		_, err = helpers.Publish(helpers.PubSubWebsockets, wsPayload)
		log.Err(err)
	}()

	wg.Wait()

	if payload.actionTaken {
		return
	}

	//
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

	summary, b, err := helpers.GetSteam().GetPlayer(player.ID)
	err = helpers.HandleSteamStoreErr(err, b, nil)
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
	resp, b, err := helpers.GetSteam().GetOwnedGames(player.ID)
	err = helpers.HandleSteamStoreErr(err, b, nil)
	if err != nil {
		return err
	}

	// Save count
	player.GamesCount = len(resp.Games)

	// Start creating PlayerApp's
	var playerApps = map[int]*mongo.PlayerApp{}
	var appPrices = map[int]map[string]int{}
	var appPriceHour = map[int]map[string]float64{}
	var appIDs []int
	var playtime = 0
	for _, v := range resp.Games {
		playtime = playtime + v.PlaytimeForever
		appIDs = append(appIDs, v.AppID)
		playerApps[v.AppID] = &mongo.PlayerApp{
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

	// Getting missing price info from MySQL
	gameRows, err := sql.GetAppsByID(appIDs, []string{"id", "prices"})
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
	err = mongo.UpdatePlayerApps(playerApps)
	if err != nil {
		return err
	}

	// Save stats to player
	var gameStats = mongo.PlayerAppStatsTemplate{}
	for _, v := range playerApps {

		gameStats.All.AddApp(v.AppTime, appPrices[v.AppID], appPriceHour[v.AppID])
		if v.AppTime > 0 {
			gameStats.Played.AddApp(v.AppTime, appPrices[v.AppID], appPriceHour[v.AppID])
		}
	}

	b, err = json.Marshal(gameStats)
	if err != nil {
		return err
	}

	player.GameStats = string(b)

	return nil
}

func updatePlayerRecentGames(player *mongo.Player) error {

	recentResponse, b, err := helpers.GetSteam().GetRecentlyPlayedGames(player.ID)
	err = helpers.HandleSteamStoreErr(err, b, nil)
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

	b, err = json.Marshal(games)
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

	response, b, err := helpers.GetSteam().GetBadges(player.ID)
	err = helpers.HandleSteamStoreErr(err, b, nil)
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

	b, err = json.Marshal(stats)
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
	var appRowsMap = map[int]sql.App{}
	appRows, err := sql.GetAppsByID(appIDSlice, []string{"id", "name", "icon"})
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

	resp, b, err := helpers.GetSteam().GetFriendList(player.ID)
	err = helpers.HandleSteamStoreErr(err, b, []int{401})
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
	friendRows, err := mongo.GetPlayersByID(friendsSlice, nil)
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

	// Make into map again, so it can be marshalled
	var friends []mongo.ProfileFriend
	for _, v := range friendsMap {
		friends = append(friends, *v)
	}

	// Encode to JSON bytes
	b, err = json.Marshal(friends)
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

	level, b, err := helpers.GetSteam().GetSteamLevel(player.ID)
	err = helpers.HandleSteamStoreErr(err, b, nil)
	if err != nil {
		return err
	}

	player.Level = level

	return nil
}

func updatePlayerBans(player *mongo.Player) error {

	response, b, err := helpers.GetSteam().GetPlayerBans(player.ID)
	err = helpers.HandleSteamStoreErr(err, b, nil)
	if err == steam.ErrNoUserFound {
		return nil
	}
	if err != nil {
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
	b, err = json.Marshal(bans)
	if err != nil {
		return err
	}

	player.Bans = string(b)

	return nil
}

func updatePlayerGroups(player *mongo.Player) error {

	resp, b, err := helpers.GetSteam().GetUserGroupList(player.ID)
	err = helpers.HandleSteamStoreErr(err, b, []int{403})
	if err != nil {
		return err
	}

	player.Groups = resp.GetIDs()

	// Queue groups for update
	for _, v := range player.Groups {
		err = ProduceGroup(strconv.FormatInt(v, 10))
		log.Err(err)
	}

	return nil
}

func updatePlayerWishlist(player *mongo.Player) error {

	resp, b, err := helpers.GetSteam().GetWishlist(player.ID)
	err = helpers.HandleSteamStoreErr(err, b, []int{500})
	if err == steam.ErrWishlistNotFound {
		return nil
	} else if err != nil {
		return err
	}

	var appIDs []int

	for k := range resp.Items {

		i, err := strconv.Atoi(k)
		if err != nil {
			appIDs = append(appIDs, i)
		}
	}

	player.Wishlist = appIDs

	return nil
}

func savePlayerMongo(player mongo.Player) error {

	_, err := mongo.ReplaceDocument(mongo.CollectionPlayers, mongo.M{"_id": player.ID}, player)

	return err
}

func savePlayerToInflux(player mongo.Player) (err error) {

	fields := map[string]interface{}{
		"level":    player.Level,
		"games":    player.GamesCount,
		"badges":   player.BadgesCount,
		"playtime": player.PlayTime,
		"friends":  player.FriendsCount,
	}

	if player.LevelRank > 0 {
		fields["level_rank"] = player.LevelRank
	}
	if player.GamesRank > 0 {
		fields["games_rank"] = player.GamesRank
	}
	if player.BadgesRank > 0 {
		fields["badges_rank"] = player.BadgesRank
	}
	if player.PlayTimeRank > 0 {
		fields["playtime_rank"] = player.PlayTimeRank
	}
	if player.FriendsRank > 0 {
		fields["friends_rank"] = player.FriendsRank
	}

	_, err = helpers.InfluxWrite(helpers.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(helpers.InfluxMeasurementPlayers),
		Tags: map[string]string{
			"player_id": strconv.FormatInt(player.ID, 10),
		},
		Fields:    fields,
		Time:      time.Now(),
		Precision: "m",
	})

	return err
}
