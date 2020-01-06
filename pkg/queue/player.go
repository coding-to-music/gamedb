package queue

import (
	"encoding/json"
	"path"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	pubsubHelpers "github.com/gamedb/gamedb/pkg/helpers/pubsub"
	steamHelper "github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
	influx "github.com/influxdata/influxdb1-client"
	"go.mongodb.org/mongo-driver/bson"
)

type PlayerMessage struct {
	ID         int64   `json:"id"`
	SkipGroups bool    `json:"dont_queue_groups"`
	UserAgent  *string `json:"user_agent"`
}

func playerHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := PlayerMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		// Update player
		player, err := mongo.GetPlayer(payload.ID)
		err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
		if err != nil {

			log.Err(err, payload.ID)
			if err == mongo.ErrInvalidPlayerID {
				sendToFailQueue(message)
			} else {
				sendToRetryQueue(message)
			}
			continue
		}

		player.ID = payload.ID

		//
		var wg sync.WaitGroup

		// Calls to api.steampowered.com
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = updatePlayerSummary(&player)
			if err != nil {

				if err == steam.ErrNoUserFound {
					message.Ack(false)
				} else {
					steamHelper.LogSteamError(err, payload.ID)
					sendToRetryQueue(message)
				}
				return
			}

			err = updatePlayerGames(&player)
			if err != nil {
				steamHelper.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = updatePlayerRecentGames(&player)
			if err != nil {
				steamHelper.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = updatePlayerBadges(&player)
			if err != nil {
				steamHelper.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = updatePlayerFriends(&player)
			if err != nil {
				steamHelper.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = updatePlayerLevel(&player)
			if err != nil {
				steamHelper.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = updatePlayerBans(&player)
			if err != nil {
				steamHelper.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = updatePlayerGroups(&player, payload)
			if err != nil {
				steamHelper.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		// Calls to store.steampowered.com
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = updatePlayerWishlistApps(&player)
			if err != nil {
				steamHelper.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = updatePlayerComments(&player)
			if err != nil {
				steamHelper.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		// Write to databases
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = savePlayerMongo(player)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Add(1)
		go func() {

			defer wg.Done()

			err = savePlayerToInflux(player)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Add(1)
		go func() {

			defer wg.Done()

			user, err := sql.GetUserByKey("steam_id", player.ID, 0)
			if err == sql.ErrRecordNotFound {
				return
			}
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = mongo.CreateUserEvent(nil, user.ID, mongo.EventRefresh)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		// Clear caches
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = memcache.RemoveKeyFromMemCacheViaPubSub(
				memcache.MemcachePlayer(player.ID).Key,
				memcache.MemcachePlayerInQueue(player.ID).Key,
			)
			if err != nil {
				log.Err(err, payload.ID)
			}
		}()

		// Websocket
		wg.Add(1)
		go func() {

			defer wg.Done()

			wsPayload := websockets.PubSubIDStringPayload{} // String, as int64 too large for js
			wsPayload.ID = strconv.FormatInt(player.ID, 10)
			wsPayload.Pages = []websockets.WebsocketPage{websockets.PagePlayer}

			_, err = pubsubHelpers.Publish(pubsubHelpers.PubSubTopicWebsockets, wsPayload)
			if err != nil {
				log.Err(err, payload.ID)
			}
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		//
		message.Ack(false)
	}
}

func updatePlayerSummary(player *mongo.Player) error {

	summary, b, err := steamHelper.GetSteam().GetPlayer(player.ID)
	err = steamHelper.AllowSteamCodes(err, b, nil)
	if err != nil {
		return err
	}

	// Avatar
	if summary.AvatarFull != "" && helpers.GetResponseCode(helpers.AvatarBase+summary.AvatarFull) == 200 {
		player.Avatar = summary.AvatarFull
	} else {
		player.Avatar = ""
	}

	//
	player.VanintyURL = path.Base(summary.ProfileURL)
	player.CountryCode = summary.CountryCode
	player.ContinentCode = helpers.CountryCodeToContinent(summary.CountryCode)
	player.StateCode = summary.StateCode
	player.PersonaName = summary.PersonaName
	player.TimeCreated = time.Unix(summary.TimeCreated, 0)
	player.LastLogOff = time.Unix(summary.LastLogOff, 0)
	player.PrimaryGroupID = summary.PrimaryClanID

	return err
}

func updatePlayerGames(player *mongo.Player) error {

	// Grab games from Steam
	resp, b, err := steamHelper.GetSteam().GetOwnedGames(player.ID)
	err = steamHelper.AllowSteamCodes(err, b, nil)
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
	gameRows, err := sql.GetAppsByID(appIDs, []string{"id", "prices", "type"})
	if err != nil {
		return err
	}

	player.GamesByType = map[string]int{}

	for _, gameRow := range gameRows {

		// Set games by type
		if _, ok := player.GamesByType[gameRow.GetType()]; ok {
			player.GamesByType[gameRow.GetType()]++
		} else {
			player.GamesByType[gameRow.GetType()] = 1
		}

		//
		prices := gameRow.GetPrices()
		for code, vv := range prices {

			vv = prices.Get(code)

			appPrices[gameRow.ID][string(code)] = vv.Final
			if appPrices[gameRow.ID][string(code)] > 0 && playerApps[gameRow.ID].AppTime == 0 {
				appPriceHour[gameRow.ID][string(code)] = -1 // Infinite
			} else if appPrices[gameRow.ID][string(code)] > 0 && playerApps[gameRow.ID].AppTime > 0 {
				appPriceHour[gameRow.ID][string(code)] = (float64(appPrices[gameRow.ID][string(code)]) / 100) / (float64(playerApps[gameRow.ID].AppTime) / 60) * 100
			} else {
				appPriceHour[gameRow.ID][string(code)] = 0 // Free
			}
		}

		//
		playerApps[gameRow.ID].AppPrices = appPrices[gameRow.ID]
		log.Err(err)

		//
		playerApps[gameRow.ID].AppPriceHour = appPriceHour[gameRow.ID]
		log.Err(err)
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

	newAppsSlice, b, err := steamHelper.GetSteam().GetRecentlyPlayedGames(player.ID)
	err = steamHelper.AllowSteamCodes(err, b, nil)
	if err != nil {
		return err
	}

	player.RecentAppsCount = len(newAppsSlice)

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
			AppName:         helpers.GetAppName(v.AppID, v.Name),
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

	response, b, err := steamHelper.GetSteam().GetBadges(player.ID)
	err = steamHelper.AllowSteamCodes(err, b, nil)
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
			_, ok := mongo.GlobalBadges[v.BadgeID]
			if ok {
				specialBadgeIDSlice = append(specialBadgeIDSlice, v.BadgeID)
			}
		} else {
			_, ok := mongo.GlobalBadges[v.AppID]
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
	if err != nil {
		return err
	}

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

	// If it's a 401, it returns no results, we dont want to change remove the players friends.
	newFriendsSlice, _, err := steamHelper.GetSteam().GetFriendList(player.ID)
	if err2, ok := err.(steam.Error); ok && err2.Code == 401 {
		return nil
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
	friendRows, err := mongo.GetPlayersByID(friendIDsToAdd, bson.M{
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
	player.FriendsCount = len(newFriendsMap)

	return nil
}

func updatePlayerLevel(player *mongo.Player) error {

	level, b, err := steamHelper.GetSteam().GetSteamLevel(player.ID)
	err = steamHelper.AllowSteamCodes(err, b, nil)
	if err != nil {
		return err
	}

	player.Level = level

	return nil
}

func updatePlayerBans(player *mongo.Player) error {

	response, b, err := steamHelper.GetSteam().GetPlayerBans(player.ID)
	err = steamHelper.AllowSteamCodes(err, b, nil)
	if err == steam.ErrNoUserFound {
		return nil
	}
	if err != nil {
		return err
	}

	player.NumberOfGameBans = response.NumberOfGameBans
	player.NumberOfVACBans = response.NumberOfVACBans

	if response.NumberOfVACBans > 0 {
		player.LastBan = time.Now().Add(time.Hour * 24 * time.Duration(response.DaysSinceLastBan) * -1)
	} else {
		player.LastBan = time.Unix(0, 0)
	}

	//
	bans := mongo.PlayerBans{}
	bans.CommunityBanned = response.CommunityBanned
	bans.VACBanned = response.VACBanned
	bans.NumberOfVACBans = response.NumberOfVACBans
	bans.DaysSinceLastBan = response.DaysSinceLastBan
	bans.NumberOfGameBans = response.NumberOfGameBans
	bans.EconomyBan = response.EconomyBan

	b, err = json.Marshal(bans)
	if err != nil {
		return err
	}

	player.Bans = string(b)

	return nil
}

func updatePlayerGroups(player *mongo.Player, payload PlayerMessage) error {

	// Old groups
	oldGroupsSlice, err := mongo.GetPlayerGroups(player.ID, 0, 0, nil)
	if err != nil {
		return err
	}

	oldGroupsMap := map[string]mongo.PlayerGroup{}
	for _, v := range oldGroupsSlice {
		oldGroupsMap[v.GroupID] = v
	}

	// Current groups response
	currentSlice, b, err := steamHelper.GetSteam().GetUserGroupList(player.ID)
	err = steamHelper.AllowSteamCodes(err, b, []int{403})
	if err != nil {
		return err
	}

	player.GroupsCount = len(currentSlice.GetIDs())

	currentMap := map[string]string{}
	for _, v := range currentSlice.GetIDs() {
		currentMap[v] = v
	}

	// Make list of groups to add
	var newGroupIDs []string
	for _, v := range currentSlice.GetIDs() {
		if _, ok := oldGroupsMap[v]; !ok {
			newGroupIDs = append(newGroupIDs, v)
		}
	}

	// Find new groups
	newGroupsSlice, err := mongo.GetGroupsByID(newGroupIDs, nil)
	if err != nil {
		return err
	}

	// Add
	var newPlayerGroupSlice []mongo.PlayerGroup
	for _, v := range newGroupsSlice {
		newPlayerGroupSlice = append(newPlayerGroupSlice, mongo.PlayerGroup{
			PlayerID:     player.ID,
			PlayerName:   player.PersonaName,
			PlayerAvatar: player.Avatar,
			GroupID:      v.ID,
			GroupName:    v.Name,
			GroupIcon:    v.Icon,
			GroupMembers: v.Members,
			GroupType:    v.Type,
			GroupURL:     v.URL,
		})
	}

	err = mongo.InsertPlayerGroups(newPlayerGroupSlice)
	if err != nil {
		return err
	}

	// Delete
	var toDelete []string
	for _, v := range oldGroupsSlice {
		if _, ok := currentMap[v.GroupID]; !ok {
			toDelete = append(toDelete, v.GroupID)
		}
	}

	err = mongo.DeletePlayerGroups(player.ID, toDelete)
	if err != nil {
		return err
	}

	// Queue groups for update
	if !payload.SkipGroups {
		for _, id := range currentSlice.GetIDs() {
			err = ProduceGroup(GroupMessage{ID: id, UserAgent: payload.UserAgent})
			log.Err(err)
		}
	}

	return nil
}

func updatePlayerWishlistApps(player *mongo.Player) error {

	// New
	resp, b, err := steamHelper.GetSteam().GetWishlist(player.ID)
	err = steamHelper.AllowSteamCodes(err, b, []int{500})
	if err == steam.ErrWishlistNotFound {
		return nil
	} else if err != nil {
		return err
	}

	var newAppSlice = resp.Items

	player.WishlistAppsCount = len(resp.Items)

	var newAppMap = map[int]steam.WishlistItem{}
	for k, v := range newAppSlice {
		newAppMap[int(k)] = v
	}

	// Old
	oldAppsSlice, err := mongo.GetPlayerWishlistAppsByPlayer(player.ID, 0, 0, nil)
	if err != nil {
		return err
	}

	oldAppsMap := map[int]mongo.PlayerWishlistApp{}
	for _, v := range oldAppsSlice {
		oldAppsMap[v.AppID] = v
	}

	// Delete
	var toDelete []int
	for _, v := range oldAppsSlice {
		if _, ok := newAppMap[v.AppID]; !ok {
			toDelete = append(toDelete, v.AppID)
		}
	}

	err = mongo.DeletePlayerWishlistApps(player.ID, toDelete)
	if err != nil {
		return err
	}

	// Add
	var appIDs []int
	var toAdd []mongo.PlayerWishlistApp
	for appID, v := range newAppMap {
		if _, ok := oldAppsMap[appID]; !ok {
			appIDs = append(appIDs, appID)
			toAdd = append(toAdd, mongo.PlayerWishlistApp{
				PlayerID: player.ID,
				AppID:    appID,
				Order:    v.Priority,
			})
		}
	}

	// Fill in data from SQL
	apps, err := sql.GetAppsByID(appIDs, []string{"id", "name", "icon", "release_state", "release_date", "release_date_unix", "prices"})
	if err != nil {
		return err
	}

	var appsMap = map[int]sql.App{}
	for _, app := range apps {
		appsMap[app.ID] = app
	}

	for k, v := range toAdd {
		toAdd[k].AppPrices = appsMap[v.AppID].GetPrices().Map()
		toAdd[k].AppName = appsMap[v.AppID].Name
		toAdd[k].AppIcon = appsMap[v.AppID].Icon
		toAdd[k].AppReleaseState = appsMap[v.AppID].ReleaseState
		toAdd[k].AppReleaseDate = time.Unix(appsMap[v.AppID].ReleaseDateUnix, 0)
		toAdd[k].AppReleaseDateNice = appsMap[v.AppID].ReleaseDate
	}

	err = mongo.InsertPlayerWishlistApps(toAdd)
	if err != nil {
		return err
	}

	return nil
}

func updatePlayerComments(player *mongo.Player) error {

	resp, _, err := steamHelper.GetSteam().GetComments(player.ID, 1, 0)
	if err != nil {
		return err
	}

	player.CommentsCount = resp.TotalCount

	return nil
}

func savePlayerMongo(player mongo.Player) error {

	_, err := mongo.ReplaceOne(mongo.CollectionPlayers, bson.D{{"_id", player.ID}}, player)
	return err
}

func savePlayerToInflux(player mongo.Player) (err error) {

	fields := map[string]interface{}{
		"level":    player.Level,
		"games":    player.GamesCount,
		"badges":   player.BadgesCount,
		"playtime": player.PlayTime,
		"friends":  player.FriendsCount,
		"comments": player.CommentsCount,
	}

	// Add ranks to map
	for k, v := range mongo.PlayerRankFieldsInflux {

		if val, ok := player.Ranks[string(k)]; ok && val > 0 {
			fields[v] = val
		}
	}

	// Save
	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementPlayers),
		Tags: map[string]string{
			"player_id": strconv.FormatInt(player.ID, 10),
		},
		Fields:    fields,
		Time:      time.Now(),
		Precision: "m",
	})

	return err
}
