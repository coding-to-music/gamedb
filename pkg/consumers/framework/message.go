package framework

import (
	"sync"

	"github.com/streadway/amqp"
)

type Message struct {
	queue       *Queue
	messages    []*amqp.Delivery // Slice for bulk acking
	actionTaken bool
	sync.Mutex
}

func (message *Message) ack() error {

	message.Lock()
	defer message.Unlock()

	if message.actionTaken {
		return nil
	}

	message.actionTaken = true

	if len(message.messages) > 1 {
		var last = message.messages[len(message.messages)-1]
		return last.Ack(true)
	} else if len(message.messages) == 1 {
		return message.messages[0].Ack(false)
	}
	return nil
}

func (message Message) SendToQueue(queues ...*Queue) error {

	message.Lock()
	defer message.Unlock()

	// Prodice to new queue
	if len(queues) == 0 {
		queues = []*Queue{message.queue}
	}

	for _, queue := range queues {

		err := queue.produce(message)
		if err != nil {
			return err
		}
	}

	return message.Ack()
}
