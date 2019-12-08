package queue

import (
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type appPlayerMessage struct {
	baseMessage
	Message appPlayerMessageInner `json:"message"`
}

type appPlayerMessageInner struct {
	IDs []int `json:"ids"`
}

type appPlayerQueue struct {
}

func (q appPlayerQueue) processMessages(msgs []amqp.Delivery) {

	message := appPlayerMessage{}

	err := helpers.Unmarshal(msgs[0].Body, &message)
	if err != nil {
		log.Err(err, msgs[0].Body)
		ackFail(msgs[0], &message)
		return
	}

	payload := consumers.AppPlayerMessage{}
	payload.IDs = message.Message.IDs

	err = consumers.ProduceAppPlayers(payload)
	if err != nil {
		log.Err(err, msgs[0].Body)
		ackRetry(msgs[0], &message)
	} else {
		message.ack(msgs[0])
	}
}
