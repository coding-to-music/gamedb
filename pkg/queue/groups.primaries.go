package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type GroupPrimariesMessage struct {
	GroupID string `json:"group_id"`
}

func (m GroupPrimariesMessage) Queue() rabbit.QueueName {
	return QueueGroupsPrimaries
}

func groupPrimariesHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := GroupPrimariesMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		prims, err := mongo.CountDocuments(mongo.CollectionPlayers, bson.D{{"primary_clan_id_string", payload.GroupID}}, 0)
		if err != nil {
			log.Err(err, payload.GroupID)
			sendToRetryQueue(message)
			return
		}

		filter := bson.D{
			{"_id", payload.GroupID},
		}

		update := bson.D{
			{"primaries", int(prims)},
		}

		_, err = mongo.UpdateOne(mongo.CollectionGroups, filter, update)
		if err != nil {
			log.Err(err, payload.GroupID)
			sendToRetryQueue(message)
			continue
		}

		// Clear player cache
		err = memcache.Delete(memcache.MemcacheGroup(payload.GroupID).Key)
		if err != nil {
			log.Err(err, payload.GroupID)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
