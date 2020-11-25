package queue

import (
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/Jleagle/steam-go/steamid"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/oauth"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/gamedb/gamedb/pkg/websockets"
	influx "github.com/influxdata/influxdb1-client"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type PlayerMessage struct {
	ID                       int64   `json:"id"`
	SkipGroupUpdate          bool    `json:"dont_queue_group"`
	SkipAchievements         bool    `json:"skip_achievements"`
	SkipExistingPlayer       bool    `json:"skip_existing_player"`
	ForceAchievementsRefresh bool    `json:"force_achievements_refresh"`
	UserAgent                *string `json:"user_agent"`
}

func playerHandler(message *rabbit.Message) {

	payload := PlayerMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	if payload.ID == 0 {
		message.Ack()
		return
	}

	if payload.UserAgent != nil && helpers.IsBot(*payload.UserAgent) {
		message.Ack()
		return
	}

	payload.ID, err = helpers.IsValidPlayerID(payload.ID)
	if err != nil {
		message.Ack()
		return
	}

	// Update player
	player, err := mongo.GetPlayer(payload.ID)
	if err == nil && payload.SkipExistingPlayer {
		message.Ack()
		return
	}
	newPlayer := err == mongo.ErrNoDocuments
	err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
	if err != nil {

		log.ErrS(err, payload.ID)
		if err == steamid.ErrInvalidPlayerID {
			sendToFailQueue(message)
		} else {
			sendToRetryQueue(message)
		}
		return
	}

	player.ID = payload.ID

	// Websocket
	defer func() {

		wsPayload := PlayerPayload{
			ID:            strconv.FormatInt(player.ID, 10),
			Name:          player.GetName(),
			Link:          player.GetPath(),
			Avatar:        player.GetAvatar(),
			CommunityLink: player.CommunityLink(),
			UpdatedAt:     time.Now().Unix(),
			Queue:         "player",
		}

		err = ProduceWebsocket(wsPayload, websockets.PagePlayer)
		if err != nil {
			log.ErrS(err, payload.ID)
		}
	}()

	// Skip removed players
	if player.Removed {
		message.Ack()
		return
	}

	//
	var wg sync.WaitGroup

	// Calls to api.steampowered.com
	wg.Add(1)
	go func() {

		defer wg.Done()

		err = updatePlayerSummary(&player)
		if err != nil {

			if err == steamapi.ErrProfileMissing {
				player.Removed = true
			} else {
				steam.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
			}
			return
		}

		err = updatePlayerRecentGames(&player, payload)
		if err != nil {
			steam.LogSteamError(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		err = updatePlayerFriends(&player)
		if err != nil {
			steam.LogSteamError(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		err = updatePlayerLevel(&player)
		if err != nil {
			steam.LogSteamError(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		err = updatePlayerBans(&player)
		if err != nil {
			steam.LogSteamError(err, payload.ID)
			sendToRetryQueue(message)
			return
		}
	}()

	// Calls to store.steampowered.com
	wg.Add(1)
	go func() {

		defer wg.Done()

		err = updatePlayerComments(&player)
		if err != nil {
			steam.LogSteamError(err, payload.ID)
			sendToRetryQueue(message)
			return
		}
	}()

	wg.Wait()

	if message.ActionTaken {
		return
	}

	// Read from Mongo databases
	wg.Add(1)
	go func() {

		defer wg.Done()

		apps, err := mongo.GetPlayerWishlistAppsByPlayer(player.ID, 0, 0, nil, bson.M{"app_prices": 1})
		if err != nil {
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		var total = map[steamapi.ProductCC]int{}
		for _, app := range apps {

			for code, price := range app.AppPrices {
				total[code] += price
			}
		}

		player.WishlistTotalCost = total
	}()

	wg.Wait()

	if message.ActionTaken {
		return
	}

	// Write to databases
	wg.Add(1)
	go func() {

		defer wg.Done()

		err = savePlayerRow(player)
		if err != nil {
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}
	}()

	if newPlayer {
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = updatePlayerFriendRows(player)
			if err != nil {
				log.ErrS(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()
	}

	wg.Add(1)
	go func() {

		defer wg.Done()

		err = savePlayerToInflux(player)
		if err != nil {
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		user, err := mysql.GetUserByProviderID(oauth.ProviderSteam, strconv.FormatInt(player.ID, 10))
		if err == mysql.ErrRecordNotFound {
			return
		}
		if err != nil {
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		err = mongo.NewEvent(nil, user.ID, mongo.EventRefresh)
		if err != nil {
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}
	}()

	wg.Wait()

	if message.ActionTaken {
		return
	}

	// Clear caches
	wg.Add(1)
	go func() {

		defer wg.Done()

		var items = []string{
			memcache.MemcachePlayer(player.ID).Key,
			memcache.MemcachePlayerInQueue(player.ID).Key,
		}

		err = memcache.Delete(items...)
		if err != nil {
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}
	}()

	wg.Wait()

	if message.ActionTaken {
		return
	}

	// Produce to sub queues
	var produces = []QueueMessageInterface{
		// PlayersSearchMessage{Player: &player}, // Done in sub queues
		PlayerBadgesMessage{
			PlayerID:     player.ID,
			PlayerName:   player.PersonaName,
			PlayerAvatar: player.Avatar,
			OldCount:     player.BadgesCount,
			OldFoilCount: player.BadgesFoilCount,
		},
		PlayerGamesMessage{
			PlayerID:                 player.ID,
			PlayerCountry:            player.CountryCode,
			PlayerUpdated:            player.UpdatedAt,
			SkipAchievements:         payload.SkipAchievements,
			ForceAchievementsRefresh: payload.ForceAchievementsRefresh,
			OldAchievementCount:      player.AchievementCount,
			OldAchievementCount100:   player.AchievementCount100,
			OldAchievementCountApps:  player.AchievementCountApps,
		},
		PlayersGroupsMessage{
			Player:          player,
			SkipGroupUpdate: payload.SkipGroupUpdate,
			UserAgent:       payload.UserAgent,
		},
		PlayersAliasesMessage{
			PlayerID:      player.ID,
			PlayerRemoved: player.Removed,
		},
		PlayersWishlistMessage{
			PlayerID: player.ID,
		},
	}

	for _, v := range produces {
		err = produce(v.Queue(), v)
		if err != nil {
			log.ErrS(err)
			sendToRetryQueue(message)
			break
		}
	}

	if message.ActionTaken {
		return
	}

	//
	message.Ack()
}
func updatePlayerSummary(player *mongo.Player) error {

	summary, err := steam.GetSteam().GetPlayer(player.ID)
	err = steam.AllowSteamCodes(err)
	if err != nil {
		return err
	}

	// Avatar
	if strings.Contains(summary.ProfileURL, "/id/") {
		player.VanityURL = path.Base(summary.ProfileURL)
	}

	continent, err := i18n.CountryCodeToContinent(summary.CountryCode)
	if err != nil {
		log.Err(err.Error(), zap.Int64("player", player.ID), zap.String("country", summary.CountryCode))
	}

	player.Avatar = summary.AvatarHash
	player.CountryCode = summary.CountryCode
	player.ContinentCode = continent
	player.StateCode = summary.StateCode
	player.PersonaName = summary.PersonaName
	player.TimeCreated = time.Unix(summary.TimeCreated, 0)
	player.PrimaryGroupID = summary.PrimaryClanID
	player.CommunityVisibilityState = summary.CommunityVisibilityState

	return err
}

func updatePlayerRecentGames(player *mongo.Player, payload PlayerMessage) error {

	// Get data
	oldAppsSlice, err := mongo.GetRecentApps(player.ID, 0, 0, nil)
	if err != nil {
		return err
	}

	newAppsSlice, err := steam.GetSteam().GetRecentlyPlayedGames(player.ID)
	err = steam.AllowSteamCodes(err)
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

	err = mongo.ReplaceRecentApps(appsToAdd)
	if err != nil {
		return err
	}

	//
	if !payload.SkipAchievements && !payload.ForceAchievementsRefresh {
		if player.UpdatedAt.After(time.Now().Add(time.Hour * 24 * 13 * -1)) { // Just under 2 weeks
			for _, v := range newAppsSlice {
				err = ProducePlayerAchievements(
					player.ID, v.AppID, false,
					player.AchievementCount, player.AchievementCount100, player.AchievementCountApps,
				)
				if err != nil {
					log.ErrS(err)
				}
			}
			err = ProducePlayerAchievements(
				player.ID, 0, false,
				player.AchievementCount, player.AchievementCount100, player.AchievementCountApps,
			)
			if err != nil {
				log.ErrS(err)
			}
		}
	}

	return nil
}

func updatePlayerFriends(player *mongo.Player) error {

	newFriendsSlice, err := steam.GetSteam().GetFriendList(player.ID)
	err = steam.AllowSteamCodes(err, 401, 404)
	if err != nil {
		return err
	}

	//
	player.FriendsCount = len(newFriendsSlice)

	// Get data
	oldFriendsSlice, err := mongo.GetFriends(player.ID, 0, 0, nil, nil)
	if err != nil {
		return err
	}

	newFriendsMap := map[int64]steamapi.Friend{}
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
		"_id":          1,
		"avatar":       1,
		"games_count":  1,
		"persona_name": 1,
		"level":        1,
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

	return mongo.ReplacePlayerFriends(friendsToAddSlice)
}

func updatePlayerLevel(player *mongo.Player) error {

	level, err := steam.GetSteam().GetSteamLevel(player.ID)
	err = steam.AllowSteamCodes(err)
	if err != nil {
		return err
	}

	if level > player.Level {
		player.Level = level
	}

	return nil
}

func updatePlayerBans(player *mongo.Player) error {

	response, err := steam.GetSteam().GetPlayerBans(player.ID)
	err = steam.AllowSteamCodes(err)
	if err == steamapi.ErrProfileMissing {
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

func updatePlayerComments(player *mongo.Player) error {

	resp, err := steam.GetSteam().GetComments(player.ID, 1, 0)
	err = steam.AllowSteamCodes(err)
	if err != nil {
		return err
	}

	player.CommentsCount = resp.TotalCount

	return nil
}

func savePlayerRow(player mongo.Player) error {

	_, err := mongo.ReplaceOne(mongo.CollectionPlayers, bson.D{{"_id", player.ID}}, player)
	return err
}

func updatePlayerFriendRows(player mongo.Player) error {

	update := bson.D{
		{"avatar", player.Avatar},
		{"name", player.PersonaName},
		{"games", player.GamesCount}, // Not the latest value, updated in sub queue
		{"level", player.Level},
	}

	_, err := mongo.UpdateManySet(mongo.CollectionPlayerFriends, bson.D{{"friend_id", player.ID}}, update)
	return err
}

func savePlayerToInflux(player mongo.Player) (err error) {

	fields := map[string]interface{}{
		influxHelper.InfPlayersComments.String(): player.CommentsCount,
		influxHelper.InfPlayersFriends.String():  player.FriendsCount,
		influxHelper.InfPlayersLevel.String():    player.Level,
		// Others stored in sub queues
	}

	return savePlayerStatsToInflux(player.ID, fields)
}

// Helper used in other consumers
func savePlayerStatsToInflux(playerId int64, fields map[string]interface{}) error {

	point := influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementPlayers),
		Tags: map[string]string{
			"player_id": strconv.FormatInt(playerId, 10),
		},
		Fields:    fields,
		Time:      time.Now(),
		Precision: "h",
	}

	_, err := influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
	return err
}

// Helper used in other consumers
func sendPlayerWebsocket(playerID int64, key string, message *rabbit.Message) {

	wsPayload := PlayerPayload{
		ID:    strconv.FormatInt(playerID, 10),
		Queue: key,
	}

	err := ProduceWebsocket(wsPayload, websockets.PagePlayer)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
	}
}
