package queue2

import (
	"strconv"
	"time"

	"github.com/gamedb/website/helpers"
	"github.com/streadway/amqp"
)

type DelayQueue struct {
	baseQueue
}

func (q DelayQueue) process(msg amqp.Delivery) {

	time.Sleep(time.Second)

	var err error
	var payload = baseMessage{}

	err = helpers.Unmarshal(msg.Body, &payload)
	if err != nil {
		logError(err)
		return
	}

	// Limits
	if payload.MaxTime > 0 && payload.FirstSeen.Add(payload.MaxTime).Unix() < time.Now().Unix() {

		logInfo("Message removed from delay queue (" + payload.MaxTime.String() + "): " + string(msg.Body))
		payload.ack(msg)
		return
	}

	if payload.MaxAttempts > payload.Attempt {

		logInfo("Message removed from delay queue (" + strconv.Itoa(payload.MaxAttempts) + "): " + string(msg.Body))
		payload.ack(msg)
		return
	}

	//
	var queue QueueName

	if payload.NextAttempt.Unix() <= time.Now().Unix() {

		logInfo("Sending back")
		queue = payload.OriginalQueue

	} else {

		logInfo("Delaying for " + payload.NextAttempt.Sub(time.Now()).String())
		queue = QueueDelaysGo
	}

	err = produce(queue, payload)
	if err != nil {
		logError(err)
		return
	}

	if err == nil {
		err = msg.Ack(false)
		logError(err)
	}
}
