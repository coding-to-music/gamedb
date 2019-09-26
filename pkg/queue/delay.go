package queue

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/streadway/amqp"
)

type delayMessage struct {
	baseMessage
	Message interface{} `json:"message"`
}

type delayQueue struct {
	BaseQueue baseQueue
}

func (q delayQueue) processMessages(msgs []amqp.Delivery) {

	time.Sleep(time.Second / 10)

	msg := msgs[0]

	var err error
	var message = delayMessage{}

	err = helpers.Unmarshal(msg.Body, &message)
	if err != nil {
		logError(err)
		return
	}

	// Limits
	if q.BaseQueue.getMaxTime() > 0 && message.FirstSeen.Add(q.BaseQueue.getMaxTime()).Unix() < time.Now().Unix() {

		logInfo("Message removed from delay queue (Over " + q.BaseQueue.getMaxTime().String() + " / " + message.FirstSeen.Add(q.BaseQueue.getMaxTime()).String() + "): " + string(msg.Body))
		message.fail(msg)
		return
	}

	if q.BaseQueue.maxAttempts > 0 && message.Attempt > q.BaseQueue.maxAttempts {

		logInfo("Message removed from delay queue (" + strconv.Itoa(message.Attempt) + "/" + strconv.Itoa(q.BaseQueue.maxAttempts) + " attempts): " + string(msg.Body))
		message.fail(msg)
		return
	}

	if message.OriginalQueue == queueDelays {

		logInfo("Message removed from delay queue (Stuck in delay queue): " + string(msg.Body))
		message.fail(msg)
		return
	}

	//
	var queue queueName

	if message.getNextAttempt().Unix() <= time.Now().Unix() {

		logInfo("Sending back to " + string(message.OriginalQueue))
		queue = message.OriginalQueue

	} else {

		// logInfo("Sending " + msg.MessageId + " back in " + message.getNextAttempt().Sub(time.Now()).String())
		queue = queueDelays
	}

	err = produce(&message, queue)
	if err != nil {
		logError(err)
		return
	}

	err = msg.Ack(false)
	if err != nil {
		logError(err)
		return
	}
}
