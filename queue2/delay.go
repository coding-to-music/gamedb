package queue2

import (
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/streadway/amqp"
)

type DelayQueue struct {
	baseQueue
}

func (q DelayQueue) process(msg amqp.Delivery) {

	var err error
	var payload = baseMessage{}

	err = helpers.Unmarshal(msg.Body, &payload)
	if err != nil {
		log.Err(err)
		return
	}

	err = produce(QueueDelaysGo, payload)
	if err != nil {
		log.Err(err)
		return
	}

	if err == nil {
		err = msg.Ack(false)
		log.Err(err)
	}

	// delayMessage := RabbitMessageDelay{}
	//
	// err = helpers.Unmarshal(msg.Body, &delayMessage)
	// if err != nil {
	// 	return false
	// }
	//
	// if len(delayMessage.OriginalMessage) == 0 {
	// 	return false
	// }
	//
	// if delayMessage.EndTime.UnixNano() > time.Now().UnixNano() {
	//
	// 	// Re-delay
	// 	logInfo("Re-delay: attemp: " + strconv.Itoa(delayMessage.Attempt))
	//
	// 	delayMessage.IncrementAttempts()
	//
	// 	b, err := json.Marshal(delayMessage)
	// 	if err != nil {
	// 		return false
	// 	}
	//
	// 	err = produce(delayMessage.getConsumeQueue(), b)
	// 	logError(err)
	//
	// } else {
	//
	// 	// Add to original queue
	// 	logInfo("Re-trying after attempt: " + strconv.Itoa(delayMessage.Attempt))
	//
	// 	err = produce(delayMessage.getConsumeQueue(), []byte(delayMessage.OriginalMessage))
	// }
	//
	// if err != nil {
	// 	return true
	// }
	//
	// return false
}
