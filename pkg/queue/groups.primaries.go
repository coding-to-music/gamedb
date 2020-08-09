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
	Group mongo.Group `json:"group"`
}

func (m GroupPrimariesMessage) Queue() rabbit.QueueName {
	return QueueGroupsPrimaries
}

func groupPrimariesHandler(message *rabbit.Message) {

	payload := GroupPrimariesMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToFailQueue(message)
		return
	}

	prims, err := mongo.CountDocuments(mongo.CollectionPlayers, bson.D{{"primary_clan_id_string", payload.Group.ID}}, 0)
	if err != nil {
		log.Err(err, payload.Group.ID)
		sendToRetryQueue(message)
		return
	}

	if payload.Group.Primaries == int(prims) {
		message.Ack()
		return
	}

	filter := bson.D{
		{"_id", payload.Group.ID},
	}

	update := bson.D{
		{"primaries", int(prims)},
	}

	_, err = mongo.UpdateOne(mongo.CollectionGroups, filter, update)
	if err != nil {
		log.Err(err, payload.Group.ID)
		sendToRetryQueue(message)
		return
	}

	// Clear player cache
	err = memcache.Delete(memcache.MemcacheGroup(payload.Group.ID).Key)
	if err != nil {
		log.Err(err, payload.Group.ID)
		sendToRetryQueue(message)
		return
	}

	// Update Elastic
	payload.Group.Primaries = int(prims)

	err = ProduceGroupSearch(payload.Group)
	if err != nil {
		log.Err(err, payload.Group.ID)
		sendToRetryQueue(message)
		return
	}

	message.Ack()
}
