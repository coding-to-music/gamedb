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

func (message *Message) Ack() (err error) {

	message.Lock()
	defer message.Unlock()

	if message.actionTaken {
		return nil
	}

	if len(message.messages) > 1 {
		var last = message.messages[len(message.messages)-1]
		err = last.Ack(true)
	} else if len(message.messages) == 1 {
		err = message.messages[0].Ack(false)
	}

	if err == nil {
		message.actionTaken = true
	}

	return err
}

func (message Message) SendToQueue(queues ...*Queue) error {

	// Prodice to new queue
	if len(queues) == 0 {
		queues = []*Queue{message.queue}
	}

	for _, queue := range queues {

		err := queue.Produce(message)
		if err != nil {
			return err
		}
	}

	return message.Ack()
}
