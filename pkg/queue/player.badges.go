package queue

import (
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type PlayerBadgesMessage struct {
	PlayerID     int64  `json:"player_id"`
	PlayerName   string `json:"player_name"`
	PlayerAvatar string `json:"player_avatar"`
}

func (m PlayerBadgesMessage) Queue() rabbit.QueueName {
	return QueuePlayersBadges
}

// Always updates player_app rows as the playtime will change
func playerBadgesHandler(message *rabbit.Message) {

	payload := PlayerBadgesMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		zap.L().Error(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToFailQueue(message)
		return
	}

	// Websocket
	defer sendPlayerWebsocket(payload.PlayerID, "badge", message)

	//
	response, err := steam.GetSteam().GetBadges(payload.PlayerID)
	err = steam.AllowSteamCodes(err)
	if err != nil {
		zap.L().Error(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	// Save badges
	var playerBadgeSlice []mongo.PlayerBadge
	var appIDSlice []int
	var foilBadgeCount int

	for _, badge := range response.Badges {

		if badge.BorderColor {
			foilBadgeCount++
		}

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
			PlayerID:            payload.PlayerID,
			PlayerIcon:          payload.PlayerAvatar,
			PlayerName:          payload.PlayerName,
		})
	}
	appIDSlice = helpers.UniqueInt(appIDSlice)

	// Make map of app rows
	var appRowsMap = map[int]mongo.App{}
	appRows, err := mongo.GetAppsByID(appIDSlice, bson.M{"_id": 1, "name": 1, "icon": 1})
	if err != nil {
		zap.L().Error(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	for _, v := range appRows {
		appRowsMap[v.ID] = v
	}

	// Finish badges slice
	for k, v := range playerBadgeSlice {

		if v.IsSpecial() {
			if badge, ok := helpers.BuiltInSpecialBadges[v.BadgeID]; ok {
				playerBadgeSlice[k].AppName = badge.Name
			}
		} else {
			if app, ok := appRowsMap[v.AppID]; ok {
				playerBadgeSlice[k].AppName = app.Name
				playerBadgeSlice[k].BadgeIcon = app.Icon
			}
		}
	}

	// Save to Mongo
	err = mongo.ReplacePlayerBadges(playerBadgeSlice)
	if err != nil {
		zap.L().Error(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	// Update player row
	update := bson.D{
		{"badges_count", len(response.Badges)},
		{"badges_foil_count", foilBadgeCount},
		{"badge_stats", mongo.ProfileBadgeStats{
			PlayerXP:                   response.PlayerXP,
			PlayerLevel:                response.PlayerLevel,
			PlayerXPNeededToLevelUp:    response.PlayerXPNeededToLevelUp,
			PlayerXPNeededCurrentLevel: response.PlayerXPNeededCurrentLevel,
			PercentOfLevel:             response.GetPercentOfLevel(),
		}},
	}

	_, err = mongo.UpdateOne(mongo.CollectionPlayers, bson.D{{"_id", payload.PlayerID}}, update)
	if err != nil {
		zap.L().Error(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	// Save to Influx
	fields := map[string]interface{}{
		"badges":      len(response.Badges),
		"badges_foil": foilBadgeCount,
	}

	err = savePlayerStatsToInflux(payload.PlayerID, fields)
	if err != nil {
		zap.L().Error(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	// Clear player cache
	err = memcache.Delete(memcache.MemcachePlayer(payload.PlayerID).Key)
	if err != nil {
		zap.L().Error(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	// Update Elastic
	err = ProducePlayerSearch(nil, payload.PlayerID)
	if err != nil {
		zap.L().Error(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	//
	message.Ack()
}
