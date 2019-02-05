package queue

import (
	"strconv"
	"time"

	"github.com/gamedb/website/helpers"
	"github.com/streadway/amqp"
)

type delayQueue struct {
	baseQueue
}

func (q delayQueue) processMessage(msg amqp.Delivery) {

	time.Sleep(time.Second / 10)

	var err error
	var payload = baseMessage{}

	err = helpers.Unmarshal(msg.Body, &payload)
	if err != nil {
		logError(err)
		return
	}

	// Limits
	if payload.MaxTime > 0 && payload.FirstSeen.Add(payload.MaxTime).Unix() < time.Now().Unix() {

		logInfo("Message removed from delay queue (Over " + payload.MaxTime.String() + "): " + string(msg.Body))
		payload.stop(msg)
		return
	}

	if payload.MaxAttempts > 0 && payload.Attempt > payload.MaxAttempts {

		logInfo("Message removed from delay queue (" + strconv.Itoa(payload.Attempt) + "/" + strconv.Itoa(payload.MaxAttempts) + " attempts): " + string(msg.Body))
		payload.stop(msg)
		return
	}

	//
	var queue queueName

	if payload.getNextAttempt().Unix() <= time.Now().Unix() {

		logInfo("Sending back")
		queue = payload.OriginalQueue

	} else {

		// logInfo("Sending back in " + payload.NextAttempt.Sub(time.Now()).String())
		queue = queueGoDelays
	}

	err = produce(payload, queue)
	if err != nil {
		logError(err)
		return
	}

	if err == nil {
		err = msg.Ack(false)
		logError(err)
	}
}
