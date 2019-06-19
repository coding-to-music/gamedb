package queue

import (
	"errors"
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
)

type groupQueueAPI struct {
	baseQueue
}

func (q groupQueueAPI) processMessages(msgs []amqp.Delivery) {

	msg := msgs[0]

	var err error
	var payload = baseMessage{
		Message:       groupMessage{},
		OriginalQueue: queueGoGroupsNew,
	}

	err = helpers.Unmarshal(msg.Body, &payload)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	var message groupMessage
	err = mapstructure.Decode(payload.Message, &message)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	//
	if !helpers.IsValidGroupID(message.ID) {
		log.Err(errors.New("invalid group id: " + message.ID))
		payload.ack(msg)
		return
	}

	// See if it's been added
	group, err := mongo.GetGroup(message.ID)
	log.Err(err)
	if err == nil {
		log.Info("Putting group back into first queue")
		err = ProduceGroup([]string{message.ID})
		log.Err()
		payload.ack(msg)
		return
	}

	//
	err = updateGroupFromXML(message.ID, &group)
	if err != nil {
		if err.Error() == "expected element type <memberList> but have <html>" {
			payload.ack(msg)
			return
		} else {
			logError(err, message.ID)
			payload.ackRetry(msg)
			return
		}
	}

	//
	err = saveGroupToMongo(group)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	//
	err = saveGroupToInflux(group)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	//
	err = helpers.RemoveKeyFromMemCacheViaPubSub(
		helpers.MemcacheGroup(group.ID64).Key,
		helpers.MemcacheGroup(strconv.Itoa(group.ID)).Key,
	)
	if err != nil {
		logError(err, message.ID)
	}

	//
	err = sendGroupWebsocket([]string{message.ID})
	if err != nil {
		logError(err, message.ID)
	}

	//
	payload.ack(msg)
}
