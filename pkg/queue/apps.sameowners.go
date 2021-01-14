package queue

import (
	"sort"

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

	ownerRows, err := mongo.GetAppOwners(payload.AppID)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToFailQueue(message)
		return
	}

	if len(ownerRows) == 0 {
		message.Ack()
		return
	}

	var playerIDs []int64
	for _, v := range ownerRows {
		playerIDs = append(playerIDs, v.PlayerID)
	}

	apps, err := mongo.GetPlayerAppsByPlayers(playerIDs, bson.M{"_id": -1, "app_id": 1})
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToFailQueue(message)
		return
	}

	var countMap = map[int]int{}
	for _, v := range apps {
		if _, ok := countMap[v.AppID]; ok {
			countMap[v.AppID]++
		} else {
			countMap[v.AppID] = 1
		}
	}

	var countSlice []helpers.TupleInt
	for k, v := range countMap {
		countSlice = append(countSlice, helpers.TupleInt{Key: k, Value: v})
	}

	sort.Slice(countSlice, func(i, j int) bool {
		return countSlice[i].Value > countSlice[j].Value
	})

	if len(countSlice) > 100 {
		countSlice = countSlice[0:100]
	}

	// Update app row
	update := bson.D{{"related_owners_app_ids", countSlice}}

	_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", payload.AppID}}, update)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	// Clear cache
	err = memcache.Delete(memcache.ItemApp(payload.AppID).Key)
	if err != nil {
		log.ErrS(err, payload.AppID)
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
