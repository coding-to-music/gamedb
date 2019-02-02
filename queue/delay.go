package queue

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gamedb/website/helpers"
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

func (d RabbitMessageDelay) process(msg amqp.Delivery) (requeue bool) {

	if len(msg.Body) == 0 {
		return handleError(errEmptyMessage, false)
	}

	delayMessage := RabbitMessageDelay{}

	err := helpers.Unmarshal(msg.Body, &delayMessage)
	if err != nil {
		return handleError(err, false)
	}

	if len(delayMessage.OriginalMessage) == 0 {
		return handleError(errEmptyMessage, false)
	}

	if delayMessage.EndTime.UnixNano() > time.Now().UnixNano() {

		// Re-delay
		logInfo("Re-delay: attemp: " + strconv.Itoa(delayMessage.Attempt))

		delayMessage.IncrementAttempts()

		b, err := json.Marshal(delayMessage)
		if err != nil {
			return handleError(err, false)
		}

		err = Produce(delayMessage.getConsumeQueue(), b)
		logError(err)

	} else {

		// Add to original queue
		logInfo("Re-trying after attempt: " + strconv.Itoa(delayMessage.Attempt))

		err = Produce(delayMessage.getConsumeQueue(), []byte(delayMessage.OriginalMessage))
	}

	if err != nil {
		return handleError(err, true)
	}

	return false
}
