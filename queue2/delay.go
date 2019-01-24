package queue

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gamedb/website/helpers"
	"github.com/streadway/amqp"
)

type DelayMessage struct {
	BaseMessage
	OriginalQueue   QueueName
	OriginalMessage string
}

type DelayQueue struct {
	BaseQueue
}

func (q DelayQueue) process(msg amqp.Delivery, queue QueueName) (requeue bool) {

	if len(msg.Body) == 0 {
		return false
	}

	delayMessage := RabbitMessageDelay{}

	err = helpers.Unmarshal(msg.Body, &delayMessage)
	if err != nil {
		return false
	}

	if len(delayMessage.OriginalMessage) == 0 {
		return false
	}

	if delayMessage.EndTime.UnixNano() > time.Now().UnixNano() {

		// Re-delay
		logInfo("Re-delay: attemp: " + strconv.Itoa(delayMessage.Attempt))

		delayMessage.IncrementAttempts()

		b, err := json.Marshal(delayMessage)
		if err != nil {
			return false
		}

		err = Produce(delayMessage.getConsumeQueue(), b)
		logError(err)

	} else {

		// Add to original queue
		logInfo("Re-trying after attempt: " + strconv.Itoa(delayMessage.Attempt))

		err = Produce(delayMessage.getConsumeQueue(), []byte(delayMessage.OriginalMessage))
	}

	if err != nil {
		return true
	}

	return false
}
