package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type AppItemsMessage struct {
	AppID     int    `json:"id"`
	OldDigect string `json:"old_digect"`
}

func (m AppItemsMessage) Queue() rabbit.QueueName {
	return QueueAppsItems
}

func appItemsHandler(message *rabbit.Message) {

	payload := AppItemsMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToFailQueue(message)
		return
	}

	// Get new items
	meta, err := steam.GetSteam().GetItemDefMeta(payload.AppID)
	if err != nil {
		steam.LogSteamError(err, string(message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	if meta.Response.Digest == "" || meta.Response.Digest == payload.OldDigect {
		message.Ack()
		return
	}

	archive, err := steam.GetSteam().GetItemDefArchive(payload.AppID, meta.Response.Digest)
	if err != nil {
		steam.LogSteamError(err, string(message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	// Create new rows to update
	var newItemsMap = map[int]bool{}
	var newDocuments []mongo.AppItem
	for _, v := range archive {

		newItemsMap[int(v.ItemDefID)] = true

		appItem := mongo.AppItem{
			AppID:            int(v.AppID),
			Bundle:           v.Bundle,
			Commodity:        bool(v.Commodity),
			DateCreated:      v.DateCreated,
			Description:      v.Description,
			DisplayType:      v.DisplayType,
			DropInterval:     int(v.DropInterval),
			DropMaxPerWindow: int(v.DropMaxPerWindow),
			Hash:             v.Hash,
			IconURL:          v.IconURL,
			IconURLLarge:     v.IconURLLarge,
			ItemDefID:        int(v.ItemDefID),
			ItemQuality:      string(v.ItemQuality),
			Marketable:       bool(v.Marketable),
			Modified:         v.Modified,
			Name:             v.Name,
			Price:            v.Price,
			Promo:            v.Promo,
			Quantity:         int(v.Quantity),
			Timestamp:        v.Timestamp,
			Tradable:         bool(v.Tradable),
			Type:             v.Type,
			WorkshopID:       int64(v.WorkshopID),
			// Exchange:         v.Exchange,
			// Tags:             v.Tags,
		}
		appItem.SetExchange(v.Exchange)
		appItem.SetTags(v.Tags)

		newDocuments = append(newDocuments, appItem)
	}

	// Get items to delete
	var filter = bson.D{{"app_id", payload.AppID}}

	if len(newItemsMap) > 0 {
		var keys []int
		for k := range newItemsMap {
			keys = append(keys, k)
		}
		filter = append(filter, bson.E{Key: "item_def_id", Value: bson.M{"$nin": keys}})
	}

	resp, err := mongo.GetAppItems(0, 0, filter, bson.M{"item_def_id": 1})
	if err != nil {
		log.Err(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	var itemIDsToDelete []int
	for _, v := range resp {
		itemIDsToDelete = append(itemIDsToDelete, v.ItemDefID)
	}

	err = mongo.DeleteAppItems(payload.AppID, itemIDsToDelete)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	// Update all new items (must be after delete)
	// Always save them all incase they change
	err = mongo.ReplaceAppItems(newDocuments)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	// Update app row
	var update = bson.D{
		{"items", len(archive)},
		{"items_digest", meta.Response.Digest},
	}

	_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", payload.AppID}}, update)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	// Clear caches
	var items = []string{
		memcache.MemcacheApp(payload.AppID).Key,
		memcache.MemcacheMongoCount(mongo.CollectionAppItems.String(), bson.D{{"app_id", payload.AppID}}).Key,
	}

	err = memcache.Delete(items...)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	// Update in Elastic
	err = ProduceAppSearch(nil, payload.AppID)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	//
	message.Ack()
}
