package queue

import (
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type groupQueueAPI struct {
}

func (q groupQueueAPI) processMessages(msgs []amqp.Delivery) {

	message := appMessage{}

	err := helpers.Unmarshal(msgs[0].Body, &message)
	if err != nil {
		log.Err(err, msgs[0].Body)
		ackFail(msgs[0], &message)
		return
	}

	payload := consumers.AppMessage{}
	payload.ID = message.Message.ID
	payload.ChangeNumber = message.Message.ChangeNumber
	payload.VDF = message.Message.VDF

	err = consumers.ProduceApp(payload)
	if err != nil {
		log.Err(err, msgs[0].Body)
		ackRetry(msgs[0], &message)
	} else {
		message.ack(msgs[0])
	}
}
