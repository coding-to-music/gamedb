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

func appSameownersHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		message.Ack(false)
		continue

		payload := AppSameownersMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		ownerRows, err := mongo.GetAppOwners(payload.ID)
		if err != nil {
			log.Err(err, payload.ID)
			sendToFailQueue(message)
			continue
		}

		if len(ownerRows) == 0 {
			message.Ack(false)
			continue
		}

		var playerIDs []int64
		for _, v := range ownerRows {
			playerIDs = append(playerIDs, v.PlayerID)
		}

		apps, err := mongo.GetPlayersApps(playerIDs, bson.M{"_id": -1, "app_id": 1})
		if err != nil {
			log.Err(err, payload.ID)
			sendToFailQueue(message)
			continue
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
			continue
		}

		err = memcache.Delete(memcache.MemcacheApp(payload.ID).Key)
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
