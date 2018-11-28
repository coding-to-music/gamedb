package queue

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/logging"
	"github.com/streadway/amqp"
)

type RabbitMessageDelay struct {
	rabbitConsumer
	OriginalQueue   RabbitQueue
	OriginalMessage string
}

func (d RabbitMessageDelay) getConsumeQueue() RabbitQueue {
	return QueueDelaysData
}

func (d RabbitMessageDelay) getProduceQueue() RabbitQueue {
	return ""
}

func (d RabbitMessageDelay) getRetryData() RabbitMessageDelay {
	return RabbitMessageDelay{}
}

func (d RabbitMessageDelay) process(msg amqp.Delivery) (requeue bool, err error) {

	if len(msg.Body) == 0 {
		return false, errEmptyMessage
	}

	delayMessage := RabbitMessageDelay{}

	err = helpers.Unmarshal(msg.Body, &delayMessage)
	if err != nil {
		return false, err
	}

	if len(delayMessage.OriginalMessage) == 0 {
		return false, errEmptyMessage
	}

	if delayMessage.EndTime.UnixNano() > time.Now().UnixNano() {

		// Re-delay
		logging.Info("Re-delay: attemp: " + strconv.Itoa(delayMessage.Attempt))

		delayMessage.IncrementAttempts()

		bytes, err := json.Marshal(delayMessage)
		if err != nil {
			return false, err
		}

		err = Produce(delayMessage.getConsumeQueue(), bytes)

	} else {

		// Add to original queue
		logging.Info("Re-trying after attempt: " + strconv.Itoa(delayMessage.Attempt))

		err = Produce(delayMessage.getConsumeQueue(), []byte(delayMessage.OriginalMessage))
	}

	if err != nil {
		return true, err
	}

	return false, nil
}
