package queue

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
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
				helpers.LogSteamError(err, message.ID)
				payload.ackRetry(msg)
			}
			return
		}

		err = updatePlayerGames(&player)
		if err != nil {
			helpers.LogSteamError(err, message.ID)
			payload.ackRetry(msg)
			return
		}

		err = updatePlayerRecentGames(&player)
		if err != nil {
			helpers.LogSteamError(err, message.ID)
			payload.ackRetry(msg)
			return
		}

		err = updatePlayerBadges(&player)
		if err != nil {
			helpers.LogSteamError(err, message.ID)
			payload.ackRetry(msg)
			return
		}

		err = updatePlayerFriends(&player)
		if err != nil {
			helpers.LogSteamError(err, message.ID)
			payload.ackRetry(msg)
			return
		}

		err = updatePlayerLevel(&player)
		if err != nil {
			helpers.LogSteamError(err, message.ID)
			payload.ackRetry(msg)
			return
		}

		err = updatePlayerBans(&player)
		if err != nil {
			helpers.LogSteamError(err, message.ID)
			payload.ackRetry(msg)
			return
		}

		err = updatePlayerGroups(&player)
		if err != nil {
			helpers.LogSteamError(err, message.ID)
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
			helpers.LogSteamError(err, message.ID)
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

	// Clear caches
	wg.Add(1)
	go func() {

		defer wg.Done()

		err = helpers.RemoveKeyFromMemCacheViaPubSub(helpers.MemcachePlayer(player.ID).Key)
		if err != nil {
			logError(err, message.ID)
		}
	}()

	// Websocket
	wg.Add(1)
	go func() {

		defer wg.Done()

		wsPayload := websockets.PubSubIDStringPayload{} // String, as int64 too large for js
		wsPayload.ID = strconv.FormatInt(player.ID, 10)
		wsPayload.Pages = []websockets.WebsocketPage{websockets.PagePlayer}

		_, err = helpers.Publish(helpers.PubSubTopicWebsockets, wsPayload)
		if err != nil {
			logError(err, message.ID)
		}
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
	err = helpers.AllowSteamCodes(err, b, nil)
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
	err = helpers.AllowSteamCodes(err, b, nil)
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

	// Get top game for background
	if len(appIDs) > 0 {

		sort.Slice(appIDs, func(i, j int) bool {

			var appID1 = appIDs[i]
			var appID2 = appIDs[j]

			return playerApps[appID1].AppTime > playerApps[appID2].AppTime
		})

		player.BackgroundAppID = appIDs[0]
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

	// Get data
	oldAppsSlice, err := mongo.GetRecentApps(player.ID, 0, 0, nil)
	if err != nil {
		return err
	}

	newAppsSlice, b, err := helpers.GetSteam().GetRecentlyPlayedGames(player.ID)
	err = helpers.AllowSteamCodes(err, b, nil)
	if err != nil {
		return err
	}

	newAppsMap := map[int]steam.RecentlyPlayedGame{}
	for _, app := range newAppsSlice {
		newAppsMap[app.AppID] = app
	}

	// Apps to update
	var appsToAdd []mongo.PlayerRecentApp
	for _, v := range newAppsSlice {
		appsToAdd = append(appsToAdd, mongo.PlayerRecentApp{
			PlayerID:        player.ID,
			AppID:           v.AppID,
			AppName:         v.Name,
			PlayTime2Weeks:  v.PlayTime2Weeks,
			PlayTimeForever: v.PlayTimeForever,
			Icon:            v.ImgIconURL,
			// Logo:            v.ImgLogoURL,
		})
	}

	// Apps to remove
	var appsToRem []int
	for _, v := range oldAppsSlice {
		if _, ok := newAppsMap[v.AppID]; !ok {
			appsToRem = append(appsToRem, v.AppID)
		}
	}

	// Update DB
	err = mongo.DeleteRecentApps(player.ID, appsToRem)
	if err != nil {
		return err
	}

	err = mongo.UpdateRecentApps(appsToAdd)
	if err != nil {
		return err
	}

	return nil
}

func updatePlayerBadges(player *mongo.Player) error {

	response, b, err := helpers.GetSteam().GetBadges(player.ID)
	err = helpers.AllowSteamCodes(err, b, nil)
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

	// Save badges
	var playerBadgeSlice []mongo.PlayerBadge
	var appIDSlice []int
	var specialBadgeIDSlice []int

	for _, v := range response.Badges {
		appIDSlice = append(appIDSlice, v.AppID)
		playerBadgeSlice = append(playerBadgeSlice, mongo.PlayerBadge{
			AppID:               v.AppID,
			BadgeCompletionTime: time.Unix(v.CompletionTime, 0),
			BadgeFoil:           bool(v.BorderColor),
			BadgeID:             v.BadgeID,
			BadgeItemID:         int64(v.CommunityItemID),
			BadgeLevel:          v.Level,
			BadgeScarcity:       v.Scarcity,
			BadgeXP:             v.XP,
			PlayerID:            player.ID,
			PlayerIcon:          player.Avatar,
			PlayerName:          player.GetName(),
		})

		// Add significant badges to profile
		if v.AppID == 0 {
			_, ok := mongo.Badges[v.BadgeID]
			if ok {
				specialBadgeIDSlice = append(specialBadgeIDSlice, v.BadgeID)
			}
		} else {
			_, ok := mongo.Badges[v.AppID]
			if ok {
				specialBadgeIDSlice = append(specialBadgeIDSlice, v.AppID)
			}
		}
	}
	appIDSlice = helpers.Unique(appIDSlice)

	player.BadgeIDs = helpers.Unique(specialBadgeIDSlice)

	// Make map of app rows
	var appRowsMap = map[int]sql.App{}
	appRows, err := sql.GetAppsByID(appIDSlice, []string{"id", "name", "icon"})
	logError(err)

	for _, v := range appRows {
		appRowsMap[v.ID] = v
	}

	// Finish badges slice
	for k, v := range playerBadgeSlice {
		if app, ok := appRowsMap[v.AppID]; ok {
			playerBadgeSlice[k].AppName = app.GetName()
			playerBadgeSlice[k].BadgeIcon = app.GetIcon()
		}
	}

	// Save to Mongo
	return mongo.UpdatePlayerBadges(playerBadgeSlice)
}

func updatePlayerFriends(player *mongo.Player) error {

	// Get data
	oldFriendsSlice, err := mongo.GetFriends(player.ID, 0, 0, nil)
	if err != nil {
		return err
	}

	newFriendsSlice, b, err := helpers.GetSteam().GetFriendList(player.ID)
	err = helpers.AllowSteamCodes(err, b, []int{401})
	if err != nil {
		return err
	}

	newFriendsMap := map[int64]steam.Friend{}
	for _, friend := range newFriendsSlice {
		newFriendsMap[int64(friend.SteamID)] = friend
	}

	// Friends to add
	var friendIDsToAdd []int64
	var friendsToAdd = map[int64]*mongo.PlayerFriend{}
	for _, v := range newFriendsSlice {
		friendIDsToAdd = append(friendIDsToAdd, int64(v.SteamID))
		friendsToAdd[int64(v.SteamID)] = &mongo.PlayerFriend{
			PlayerID:     player.ID,
			FriendID:     int64(v.SteamID),
			Relationship: v.Relationship,
			FriendSince:  time.Unix(v.FriendSince, 0),
		}
	}

	// Friends to remove
	var friendsToRem []int64
	for _, v := range oldFriendsSlice {
		if _, ok := newFriendsMap[v.FriendID]; !ok {
			friendsToRem = append(friendsToRem, v.FriendID)
		}
	}

	// Fill in missing map the map
	friendRows, err := mongo.GetPlayersByID(friendIDsToAdd, mongo.M{
		"_id":             1,
		"avatar":          1,
		"games_count":     1,
		"persona_name":    1,
		"level":           1,
		"time_logged_off": 1,
	})
	if err != nil {
		return err
	}

	for _, friend := range friendRows {
		if friend.ID != 0 {

			friendsToAdd[friend.ID].Avatar = friend.Avatar
			friendsToAdd[friend.ID].Games = friend.GamesCount
			friendsToAdd[friend.ID].Name = friend.GetName()
			friendsToAdd[friend.ID].Level = friend.Level
			friendsToAdd[friend.ID].LoggedOff = friend.LastLogOff
		}
	}

	// Update DB
	err = mongo.DeleteFriends(player.ID, friendsToRem)
	if err != nil {
		return err
	}

	var friendsToAddSlice []*mongo.PlayerFriend
	for _, v := range friendsToAdd {
		friendsToAddSlice = append(friendsToAddSlice, v)
	}

	err = mongo.UpdateFriends(friendsToAddSlice)
	if err != nil {
		return err
	}

	//
	player.FriendsCount = len(newFriendsSlice)

	return nil
}

func updatePlayerLevel(player *mongo.Player) error {

	level, b, err := helpers.GetSteam().GetSteamLevel(player.ID)
	err = helpers.AllowSteamCodes(err, b, nil)
	if err != nil {
		return err
	}

	player.Level = level

	return nil
}

func updatePlayerBans(player *mongo.Player) error {

	response, b, err := helpers.GetSteam().GetPlayerBans(player.ID)
	err = helpers.AllowSteamCodes(err, b, nil)
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
	err = helpers.AllowSteamCodes(err, b, []int{403})
	if err != nil {
		return err
	}

	player.Groups = resp.GetIDs()

	// Queue groups for update
	err = ProduceGroup(resp.GetIDs())
	logError(err)

	return nil
}

func updatePlayerWishlist(player *mongo.Player) error {

	resp, b, err := helpers.GetSteam().GetWishlist(player.ID)
	err = helpers.AllowSteamCodes(err, b, []int{500})
	if err == steam.ErrWishlistNotFound {
		return nil
	} else if err != nil {
		return err
	}

	// Make into a slice so we can sort
	var appsSlice []wishlistItemPlusID
	for k, v := range resp.Items {
		i, err := strconv.Atoi(k)
		logError(err)
		if err == nil {
			appsSlice = append(appsSlice, wishlistItemPlusID{
				item:  v,
				appID: i,
			})
		}
	}

	sort.Slice(appsSlice, func(i, j int) bool {
		return appsSlice[i].item.Priority > appsSlice[j].item.Priority
	})

	var appIDs []int

	for _, appID := range appsSlice {
		appIDs = append(appIDs, appID.appID)
	}

	player.Wishlist = appIDs

	return nil
}

type wishlistItemPlusID struct {
	item  steam.WishlistItem
	appID int
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
