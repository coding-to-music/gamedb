package queue

import (
	"strconv"

	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/streadway/amqp"
)

type RabbitMessageApp struct {
	//
	BaseMessage

	// Returned from PICS
	PICSAppInfo RabbitMessageProduct

	// JSON must match the Updater app
	Payload struct {
		ID   []int `json:"IDs"`
		Time int64 `json:"Time"`
	}
}

type AppQueue struct {
}

func (d AppQueue) process(msg amqp.Delivery) (requeue bool) {

	var err error

	// Get message payload
	rabbitMessage := RabbitMessageApp{}

	err = helpers.Unmarshal(msg.Body, &rabbitMessage)
	if err != nil {
		log.Err(err)
		return false
	}

	message := rabbitMessage.PICSAppInfo

	logInfo("Consuming app: " + strconv.Itoa(message.ID))

	return false
}
