package queue2

import (
	"sync"

	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/streadway/amqp"
)

const (
	headerAttempt    = "attempt"
	headerFirstSeen  = "first-seen"
	headerLastSeen   = "last-seen"
	headerFirstQueue = "first-queue"
	headerLastQueue  = "first-queue"
	headerForce      = "force"
)

type message struct {
	msg         *amqp.Delivery
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

	return message.msg.Ack(false)
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

	return message.msg.Ack(false)
}

func (message message) getTryCount() int {

	val, ok := message.msg.Headers["try"]
	if ok {
		return val.(int)
	}
	return 0
}

func amqpMessage(msg *amqp.Delivery) *message {
	return &message{
		msg: msg,
	}
}
