package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type GroupPrimariesMessage struct {
	GroupID          string `json:"group_id"`
	GroupType        string `json:"group_type"`
	CurrentPrimaries int    `json:"current_primaries"`
}

func (m GroupPrimariesMessage) Queue() rabbit.QueueName {
	return QueueGroupsPrimaries
}

func groupPrimariesHandler(message *rabbit.Message) {

	payload := GroupPrimariesMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	prims, err := mongo.CountDocuments(mongo.CollectionPlayers, bson.D{{"primary_clan_id_string", payload.GroupID}}, 0)
	if err != nil {
		log.ErrS(err, payload.GroupID)
		sendToRetryQueue(message)
		return
	}

	if payload.CurrentPrimaries == int(prims) {
		message.Ack()
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
		log.ErrS(err, payload.GroupID)
		sendToRetryQueue(message)
		return
	}

	// Clear group cache
	err = memcache.Client().Delete(memcache.ItemGroup(payload.GroupID).Key)
	if err != nil {
		log.ErrS(err, payload.GroupID)
		sendToRetryQueue(message)
		return
	}

	// Update Elastic
	err = ProduceGroupSearch(nil, payload.GroupID, payload.GroupType)
	if err != nil {
		log.ErrS(err, payload.GroupID)
		sendToRetryQueue(message)
		return
	}

	message.Ack()
}
