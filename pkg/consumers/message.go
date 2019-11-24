package consumers

import (
	"sync"
	"time"

	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/streadway/amqp"
)

type message struct {
	message     *amqp.Delivery
	actionTaken bool
	sync.Mutex
}

func (message message) ack() error {

	message.Lock()
	defer message.Unlock()

	if message.actionTaken {
		return nil
	}

	message.actionTaken = true

	return message.message.Ack(false)
}

func (message message) failQueue() error {
	return message.moveQueue(queue.queueFailed)
}

// todo, send to originalQueue+"_Retry"
func (message message) retryQueue() error {
	return message.moveQueue(queue.queueDelays)
}

func (message message) moveQueue(queue queue.queueName) error {

	message.Lock()
	defer message.Unlock()

	if message.actionTaken {
		return nil
	}

	message.actionTaken = true

	err := queue.produce(message, queue)
	if err != nil {
		return err
	}

	return message.message.Ack(false)
}

const (
	headerAttempt    = "attempt"
	headerFirstSeen  = "first-seen"
	headerLastSeen   = "last-seen"
	headerFirstQueue = "first-queue"
	headerLastQueue  = "last-queue"
	headerForce      = "force"
)

func (message message) setupHeadersForProduce(queue string, force bool) {

	if message.message == nil {
		message.message = &amqp.Delivery{}
	}

	if message.message.Headers == nil {
		message.message.Headers = amqp.Table{}
	}

	attempt, ok := message.message.Headers[headerAttempt]
	if ok {
		message.message.Headers[headerAttempt] = attempt.(int) + 1
	} else {
		message.message.Headers[headerAttempt] = 1
	}

	_, ok = message.message.Headers[headerFirstSeen]
	if !ok {
		message.message.Headers[headerFirstSeen] = time.Now().Unix()
	}

	message.message.Headers[headerLastSeen] = time.Now().Unix()

	_, ok = message.message.Headers[headerFirstQueue]
	if !ok {
		message.message.Headers[headerFirstQueue] = queue
	}

	message.message.Headers[headerLastQueue] = time.Now().Unix()

	oldForce, ok := message.message.Headers[headerForce]
	if ok {
		message.message.Headers[headerForce] = oldForce
	} else {
		message.message.Headers[headerForce] = force
	}
}
