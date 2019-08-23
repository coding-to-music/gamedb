package queue

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/streadway/amqp"
)

type delayQueue struct {
	BaseQueue baseQueue
}

func (q delayQueue) processMessages(msgs []amqp.Delivery) {

	time.Sleep(time.Second / 10)

	msg := msgs[0]

	var err error
	var payload = baseMessage{}

	err = helpers.Unmarshal(msg.Body, &payload)
	if err != nil {
		logError(err)
		return
	}

	// Limits
	if q.BaseQueue.getMaxTime() > 0 && payload.FirstSeen.Add(q.BaseQueue.getMaxTime()).Unix() < time.Now().Unix() {

		logInfo("Message removed from delay queue (Over " + q.BaseQueue.getMaxTime().String() + " / " + payload.FirstSeen.Add(q.BaseQueue.getMaxTime()).String() + "): " + string(msg.Body))
		payload.fail(msg)
		return
	}

	if q.BaseQueue.maxAttempts > 0 && payload.Attempt > q.BaseQueue.maxAttempts {

		logInfo("Message removed from delay queue (" + strconv.Itoa(payload.Attempt) + "/" + strconv.Itoa(q.BaseQueue.maxAttempts) + " attempts): " + string(msg.Body))
		payload.fail(msg)
		return
	}

	if payload.OriginalQueue == queueGoDelays {

		logInfo("Message removed from delay queue (Stuck in delay queue): " + string(msg.Body))
		payload.fail(msg)
		return
	}

	//
	var queue queueName

	if payload.getNextAttempt().Unix() <= time.Now().Unix() {

		logInfo("Sending back to " + string(payload.OriginalQueue))
		queue = payload.OriginalQueue

	} else {

		// logInfo("Sending " + msg.MessageId + " back in " + payload.getNextAttempt().Sub(time.Now()).String())
		queue = queueGoDelays
	}

	err = produce(payload, queue)
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
