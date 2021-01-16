package queue

import (
	"math"
	"sort"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type AppSameownersMessage struct {
	AppID int `json:"app_id"`
}

func (m AppSameownersMessage) Queue() rabbit.QueueName {
	return QueueAppsSameowners
}

func appSameownersHandler(message *rabbit.Message) {

	payload := AppSameownersMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	log.Info("Same owners", zap.Int("app", payload.AppID))

	ownerRows, err := mongo.GetAppOwners(payload.AppID)
	if err != nil {
		log.Err(err.Error(), zap.Int("app", payload.AppID))
		sendToRetryQueue(message)
		return
	}

	log.Info("Same owners", zap.Int("owners", len(ownerRows)))

	if len(ownerRows) == 0 {

		// Update app row
		var filter = bson.D{{"_id", payload.AppID}}
		var update = bson.D{{"related_owners_app_ids_date", time.Now()}}

		_, err = mongo.UpdateOne(mongo.CollectionApps, filter, update)
		if err != nil {
			log.Err(err.Error(), zap.Int("app", payload.AppID))
			sendToRetryQueue(message)
			return
		}

		message.Ack()
		return
	}

	var playerIDs []int64
	for _, v := range ownerRows {
		playerIDs = append(playerIDs, v.PlayerID)
	}

	const batch1 = 100
	var countMap = map[int]*mongo.RelatedAppOwner{}

	for k, chunk := range helpers.ChunkInt64s(playerIDs, batch1) {

		if k%10 == 0 { // Every 10
			log.Info("Same owners", zap.Int("offset", k*batch1))
		}

		apps, err := mongo.GetPlayerAppsByPlayers(chunk, bson.M{"_id": -1, "app_id": 1})
		if err != nil {
			log.Err(err.Error(), zap.Int("app", payload.AppID))
			sendToRetryQueue(message)
			return
		}

		for _, v := range apps {

			if _, ok := countMap[v.AppID]; ok {
				countMap[v.AppID].Count++
			} else {
				countMap[v.AppID] = &mongo.RelatedAppOwner{AppID: v.AppID, Count: 1}
			}
		}
	}

	log.Info("Same owners", zap.Int("games", len(countMap)))

	// Set the order by how popular the game is
	var appIDs []int
	for k := range countMap {
		appIDs = append(appIDs, k)
	}

	const batch2 = 1000
	for k, chunk := range helpers.ChunkInts(appIDs, batch2) {

		if k%10 == 0 { // Every 10
			log.Info("Same owners, app owners", zap.Int("offset", k*batch2))
		}

		apps, err := mongo.GetAppsByID(chunk, bson.M{"owners": 1})
		if err != nil {
			log.Err(err.Error(), zap.Int("app", payload.AppID))
			sendToRetryQueue(message)
			return
		}

		for _, v := range apps {
			if countMap[v.ID].Count > 0 && v.Owners > 0 {
				countMap[v.ID].Order = float64(countMap[v.ID].Count) / math.Pow(float64(v.Owners), 0.5) // Higher removed popular apps
			}
		}
	}

	// Get top 100
	var countSlice []mongo.RelatedAppOwner
	for _, v := range countMap {
		countSlice = append(countSlice, *v)
	}

	sort.Slice(countSlice, func(i, j int) bool {
		return countSlice[i].Order > countSlice[j].Order
	})

	if len(countSlice) > 100 {
		countSlice = countSlice[0:100]
	}

	// Update app row
	var filter = bson.D{{"_id", payload.AppID}}
	var update = bson.D{
		{"related_owners_app_ids", countSlice},
		{"related_owners_app_ids_date", time.Now()},
	}

	_, err = mongo.UpdateOne(mongo.CollectionApps, filter, update)
	if err != nil {
		log.Err(err.Error(), zap.Int("app", payload.AppID))
		sendToRetryQueue(message)
		return
	}

	// Clear cache
	err = memcache.Delete(memcache.ItemApp(payload.AppID).Key)
	if err != nil {
		log.Err(err.Error(), zap.Int("app", payload.AppID))
		sendToRetryQueue(message)
		return
	}

	// Update in Elastic
	err = ProduceAppSearch(nil, payload.AppID)
	if err != nil {
		log.Err(err.Error(), zap.Int("app", payload.AppID))
		sendToRetryQueue(message)
		return
	}

	log.Info("Done")

	//
	message.Ack()
}
