package queue

import (
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type bundleMessage struct {
	baseMessage
	Message bundleMessageInner `json:"message"`
}

type bundleMessageInner struct {
	ID    int `json:"id"`
	AppID int `json:"app_id"`
}

type bundleQueue struct {
}

func (q bundleQueue) processMessages(msgs []amqp.Delivery) {

	message := bundleMessage{}

	err := helpers.Unmarshal(msgs[0].Body, &message)
	if err != nil {
		log.Err(err, msgs[0].Body)
		ackFail(msgs[0], &message)
		return
	}

	payload := consumers.BundleMessage{}
	payload.ID = message.Message.ID

	err = consumers.ProduceBundle(payload)
	if err != nil {
		log.Err(err, msgs[0].Body)
		ackRetry(msgs[0], &message)
	} else {
		message.ack(msgs[0])
	}
}
