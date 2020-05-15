package queue

import (
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
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

		if payload.ID == 0 {
			message.Ack(false)
			continue
		}

		if payload.UserAgent != nil && helpers.IsBot(*payload.UserAgent) {
			message.Ack(false)
			continue
		}

		payload.ID, err = helpers.IsValidPlayerID(payload.ID)
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
			if err == helpers.ErrInvalidPlayerID {
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

				if err == steamapi.ErrNoUserFound {
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

			err = memcache.Delete(
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

			wsPayload := StringPayload{String: strconv.FormatInt(player.ID, 10)}
			err = ProduceWebsocket(wsPayload, websockets.PagePlayer)
			if err != nil {
				log.Err(err, payload.ID)
			}
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		// Produce to sub queues
		var produces = map[rabbit.QueueName]interface{}{
			QueuePlayersSearch: PlayersSearchMessage{Player: player},
		}

		for k, v := range produces {
			err = produce(k, v)
			if err != nil {
				log.Err(err)
				sendToRetryQueue(message)
				break
			}
		}

		if message.ActionTaken {
			continue
		}

		//
		message.Ack(false)
	}
}

func updatePlayerSummary(player *mongo.Player) error {

	return player.SetPlayerSummary()
}

func updatePlayerGames(player *mongo.Player) error {

	resp, err := player.SetOwnedGames(true)
	if err != nil {
		return err
	}

	user, err := sql.GetUserByKey("steam_id", player.ID, 0)
	err = helpers.IgnoreErrors(err, sql.ErrRecordNotFound)
	if err != nil {
		return err
	}

	if user.Level >= sql.UserLevel1 {
		if player.UpdatedAt.Unix() < 1588244400 || player.UpdatedAt.Before(time.Now().Add(time.Hour*24*13*-1)) { // Just under 2 weeks
			for _, v := range resp.Games {
				if v.PlaytimeForever > 0 {
					err = ProducePlayerAchievements(player.ID, v.AppID)
					log.Err(err)
				}
			}
		}
	}

	return err
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

	newAppsMap := map[int]steamapi.RecentlyPlayedGame{}
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

	//
	user, err := sql.GetUserByKey("steam_id", player.ID, 0)
	err = helpers.IgnoreErrors(err, sql.ErrRecordNotFound)
	if err != nil {
		return err
	}

	if user.Level >= sql.UserLevel1 {
		if player.UpdatedAt.After(time.Now().Add(time.Hour * 24 * 13 * -1)) { // Just under 2 weeks
			for _, v := range newAppsSlice {
				err = ProducePlayerAchievements(player.ID, v.AppID)
				log.Err(err)
			}
		}
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
	player.BadgeStats = mongo.ProfileBadgeStats{
		PlayerXP:                   response.PlayerXP,
		PlayerLevel:                response.PlayerLevel,
		PlayerXPNeededToLevelUp:    response.PlayerXPNeededToLevelUp,
		PlayerXPNeededCurrentLevel: response.PlayerXPNeededCurrentLevel,
		PercentOfLevel:             response.GetPercentOfLevel(),
	}

	// Save badges
	var playerBadgeSlice []mongo.PlayerBadge
	var appIDSlice []int
	var specialBadgeIDSlice []int

	for _, badge := range response.Badges {

		appIDSlice = append(appIDSlice, badge.AppID)
		playerBadgeSlice = append(playerBadgeSlice, mongo.PlayerBadge{
			AppID:               badge.AppID,
			BadgeCompletionTime: time.Unix(badge.CompletionTime, 0),
			BadgeFoil:           bool(badge.BorderColor),
			BadgeID:             badge.BadgeID,
			BadgeItemID:         int64(badge.CommunityItemID),
			BadgeLevel:          badge.Level,
			BadgeScarcity:       badge.Scarcity,
			BadgeXP:             badge.XP,
			PlayerID:            player.ID,
			PlayerIcon:          player.Avatar,
			PlayerName:          player.PersonaName,
		})

		// Add significant badges to profile
		if badge.AppID == 0 {
			_, ok := mongo.GlobalBadges[badge.BadgeID]
			if ok {
				specialBadgeIDSlice = append(specialBadgeIDSlice, badge.BadgeID)
			}
		} else {
			_, ok := mongo.GlobalBadges[badge.AppID]
			if ok {
				specialBadgeIDSlice = append(specialBadgeIDSlice, badge.AppID)
			}
		}
	}
	appIDSlice = helpers.UniqueInt(appIDSlice)

	player.BadgeIDs = helpers.UniqueInt(specialBadgeIDSlice)

	// Make map of app rows
	var appRowsMap = map[int]mongo.App{}
	appRows, err := mongo.GetAppsByID(appIDSlice, bson.M{"_id": 1, "name": 1, "icon": 1})
	if err != nil {
		return err
	}

	for _, v := range appRows {
		appRowsMap[v.ID] = v
	}

	// Finish badges slice
	for k, v := range playerBadgeSlice {
		if app, ok := appRowsMap[v.AppID]; ok {
			playerBadgeSlice[k].AppName = app.Name
			playerBadgeSlice[k].BadgeIcon = app.Icon
		}
	}

	// Save to Mongo
	return mongo.UpdatePlayerBadges(playerBadgeSlice)
}

func updatePlayerFriends(player *mongo.Player) error {

	return player.SetFriends(true)
}

func updatePlayerLevel(player *mongo.Player) error {

	return player.SetLevel()
}

func updatePlayerBans(player *mongo.Player) error {

	response, b, err := steamHelper.GetSteam().GetPlayerBans(player.ID)
	err = steamHelper.AllowSteamCodes(err, b, nil)
	if err == steamapi.ErrNoUserFound {
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
	player.Bans = mongo.PlayerBans{
		CommunityBanned:  response.CommunityBanned,
		VACBanned:        response.VACBanned,
		NumberOfVACBans:  response.NumberOfVACBans,
		DaysSinceLastBan: response.DaysSinceLastBan,
		NumberOfGameBans: response.NumberOfGameBans,
		EconomyBan:       response.EconomyBan,
	}

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
	for _, group := range newGroupsSlice {
		newPlayerGroupSlice = append(newPlayerGroupSlice, mongo.PlayerGroup{
			PlayerID:     player.ID,
			PlayerName:   player.PersonaName,
			PlayerAvatar: player.Avatar,
			GroupID:      group.ID,
			GroupName:    helpers.TruncateString(group.Name, 1000, ""), // Truncated as caused mongo driver issue
			GroupIcon:    group.Icon,
			GroupMembers: group.Members,
			GroupType:    group.Type,
			GroupURL:     group.URL,
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
			err = helpers.IgnoreErrors(err, memcache.ErrInQueue, ErrIsBot)
			if err != nil {
				log.Err(err)
			}
		}
	}

	return nil
}

func updatePlayerWishlistApps(player *mongo.Player) error {

	// New
	resp, b, err := steamHelper.GetSteam().GetWishlist(player.ID)
	err = steamHelper.AllowSteamCodes(err, b, []int{500})
	if err == steamapi.ErrWishlistNotFound {
		return nil
	} else if err != nil {
		return err
	}

	var newAppSlice = resp.Items

	player.WishlistAppsCount = len(resp.Items)

	var newAppMap = map[int]steamapi.WishlistItem{}
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
	apps, err := mongo.GetAppsByID(appIDs, bson.M{"_id": 1, "name": 1, "icon": 1, "release_state": 1, "release_date": 1, "release_date_unix": 1, "prices": 1})
	if err != nil {
		return err
	}

	var appsMap = map[int]mongo.App{}
	for _, app := range apps {
		appsMap[app.ID] = app
	}

	for k, v := range toAdd {
		toAdd[k].AppPrices = appsMap[v.AppID].Prices.Map()
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
