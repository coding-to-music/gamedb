package framework

import (
	"sync"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type Message struct {
	Queue       *Queue
	Messages    []*amqp.Delivery // Slice for bulk acking
	actionTaken bool
	sync.Mutex
}

func (message *Message) Ack() {

	message.Lock()
	defer message.Unlock()

	if message.actionTaken {
		return
	}

	var err error

	if len(message.Messages) > 1 {
		var last = message.Messages[len(message.Messages)-1]
		err = last.Ack(true)
	} else if len(message.Messages) == 1 {
		err = message.Messages[0].Ack(false)
	}

	if err != nil {
		log.Err(err)
	} else {
		message.actionTaken = true
	}
}

func (message Message) SendToQueue(queues ...*Queue) {

	// Send to back of current queue if none specified
	if len(queues) == 0 {
		queues = []*Queue{message.Queue}
	}

	//
	var err error

	for _, queue := range queues {
		err = queue.Produce(message)
		log.Err(err)
	}

	if err == nil {
		message.Ack()
	}
}
