package queue

import (
	"sort"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type AppSameownersMessage struct {
	ID int `json:"id"`
}

func (m AppSameownersMessage) Queue() rabbit.QueueName {
	return QueueAppsSameowners
}

func appSameownersHandler(message *rabbit.Message) {

	message.Ack(false)
	return

	payload := AppSameownersMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToFailQueue(message)
		return
	}

	ownerRows, err := mongo.GetAppOwners(payload.ID)
	if err != nil {
		log.Err(err, payload.ID)
		sendToFailQueue(message)
		return
	}

	if len(ownerRows) == 0 {
		message.Ack(false)
		return
	}

	var playerIDs []int64
	for _, v := range ownerRows {
		playerIDs = append(playerIDs, v.PlayerID)
	}

	apps, err := mongo.GetPlayersApps(playerIDs, bson.M{"_id": -1, "app_id": 1})
	if err != nil {
		log.Err(err, payload.ID)
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

	_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", payload.ID}}, bson.D{{"related_owners_app_ids", countSlice}})
	if err != nil {
		log.Err(err, payload.ID)
		sendToRetryQueue(message)
		return
	}

	err = memcache.Delete(memcache.MemcacheApp(payload.ID).Key)
	if err != nil {
		log.Err(err, payload.ID)
		sendToRetryQueue(message)
		return
	}

	message.Ack(false)
}
