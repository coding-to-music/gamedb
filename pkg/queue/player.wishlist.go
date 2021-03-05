package queue

import (
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type PlayersWishlistMessage struct {
	PlayerID int64 `json:"player_id"`
}

func (m PlayersWishlistMessage) Queue() rabbit.QueueName {
	return QueuePlayersWishlist
}

func playersWishlistHandler(message *rabbit.Message) {

	payload := PlayersWishlistMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	// Websocket
	defer sendPlayerWebsocket(payload.PlayerID, "wishlist", message)

	//
	resp, err := steam.GetSteam().GetWishlist(payload.PlayerID)
	err = steam.AllowSteamCodes(err, 500)
	if err == steamapi.ErrWishlistNotFound {

		message.Ack()
		return

	} else if err != nil {

		steam.LogSteamError(err, zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Convert response to ints
	var newAppMap = map[int]steamapi.WishlistItem{}
	for k, v := range resp.Items {
		newAppMap[int(k)] = v
	}

	// Old
	oldAppsSlice, err := mongo.GetPlayerWishlistAppsByPlayer(payload.PlayerID, 0, 0, nil, nil)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
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

	err = mongo.DeletePlayerWishlistApps(payload.PlayerID, toDelete)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Add
	var toAddIDs []int
	var toAdd []mongo.PlayerWishlistApp
	for appID, v := range newAppMap {
		if _, ok := oldAppsMap[appID]; !ok {
			toAddIDs = append(toAddIDs, appID)
			toAdd = append(toAdd, mongo.PlayerWishlistApp{
				PlayerID: payload.PlayerID,
				AppID:    appID,
				Order:    v.Priority,
			})
		}
	}

	// Fill in data from SQL
	apps, err := mongo.GetAppsByID(toAddIDs, bson.M{"_id": 1, "name": 1, "icon": 1, "release_state": 1, "release_date": 1, "release_date_unix": 1, "prices": 1})
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
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

	err = mongo.ReplacePlayerWishlistApps(toAdd)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Update player row
	var update = bson.D{
		{"wishlist_apps_count", len(resp.Items)},
	}

	_, err = mongo.UpdateOne(mongo.CollectionPlayers, bson.D{{"_id", payload.PlayerID}}, update)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Clear caches
	var items = []string{
		memcache.ItemPlayer(payload.PlayerID).Key,
	}

	err = memcache.Client().Delete(items...)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Update Elastic
	// err = ProducePlayerSearch(nil, payload.PlayerID)
	// if err != nil {
	// 	log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
	// 	sendToRetryQueue(message)
	// 	return
	// }

	//
	message.Ack()
}
