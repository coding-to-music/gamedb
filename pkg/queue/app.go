package queue

import (
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type appMessage struct {
	baseMessage
	Message appMessageInner `json:"message"`
}

type appMessageInner struct {
	ID           int                    `json:"id"`
	ChangeNumber int                    `json:"change_number,omitempty"`
	VDF          map[string]interface{} `json:"vdf,omitempty"`
}

type appQueue struct {
}

func (q appQueue) processMessages(msgs []amqp.Delivery) {

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
