package queue

import (
	"strconv"

	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/streadway/amqp"
)

type AppMessage struct {
	// This is what we sent to updater
	Payload struct {
		ID   []int `json:"IDs"`
		Time int64 `json:"Time"`
	}
	PICSAppInfo RabbitMessageProduct
}

type AppQueue struct {
	BaseQueue
}

func (q AppQueue) process(msg amqp.Delivery) (requeue bool) {

	var err error

	// Get message payload
	rabbitMessage := BaseMessage{
		Message: AppMessage{},
	}

	err = helpers.Unmarshal(msg.Body, &rabbitMessage)
	if err != nil {
		log.Err(err)
		return false
	}

	message := rabbitMessage.PICSAppInfo

	logInfo("Consuming app: " + strconv.Itoa(message.ID))

	return false
}
