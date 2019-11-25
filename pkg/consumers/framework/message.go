package framework

import (
	"sync"

	"github.com/streadway/amqp"
)

type message struct {
	queue       *queue
	message     *amqp.Delivery
	actionTaken bool
	sync.Mutex
}

func (message *message) ack() error {

	message.Lock()
	defer message.Unlock()

	if message.actionTaken {
		return nil
	}

	message.actionTaken = true

	return message.message.Ack(false)
}

func (message message) sendToQueue(queues ...*queue) error {

	message.Lock()
	defer message.Unlock()

	if message.actionTaken {
		return nil
	}

	message.actionTaken = true

	if len(queues) == 0 {
		queues = []*queue{message.queue}
	}

	for _, queue := range queues {

		err := queue.produce(message)
		if err != nil {
			return err
		}
	}

	return message.message.Ack(false)
}
