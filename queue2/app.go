package queue

import (
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

func (q AppQueue) process(msg amqp.Delivery, queue QueueName) (requeue bool) {

	var err error
	var payload = BaseMessage{
		Message: AppMessage{},
	}

	err = helpers.Unmarshal(msg.Body, &payload)
	if err != nil {
		log.Err(err)
		return false
	}

	logInfo("Consuming app")

	return false
}
