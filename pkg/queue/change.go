package queue

import (
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type changeMessage struct {
	baseMessage
	Message changeMessageInner `json:"message"`
}

type changeMessageInner struct {
	AppIDs     map[uint32]uint32 `json:"app_ids"`
	PackageIDs map[uint32]uint32 `json:"package_ids"`
}

type changeQueue struct {
}

func (q changeQueue) processMessages(msgs []amqp.Delivery) {

	message := changeMessage{}

	err := helpers.Unmarshal(msgs[0].Body, &message)
	if err != nil {
		log.Err(err, msgs[0].Body)
		ackFail(msgs[0], &message)
		return
	}

	//
	payload := consumers.ChangesMessage{}
	payload.AppIDs = message.Message.AppIDs
	payload.PackageIDs = message.Message.PackageIDs

	err = consumers.ProduceChanges(payload)
	if err != nil {
		log.Err(err, msgs[0].Body)
		ackRetry(msgs[0], &message)
	} else {
		message.ack(msgs[0])
	}
}
