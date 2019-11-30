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

const (
	HeaderAttempt    = "attempt"
	HeaderFirstSeen  = "first-seen"
	HeaderLastSeen   = "last-seen"
	HeaderFirstQueue = "first-queue"
	HeaderLastQueue  = "last-queue"
)

func (message Message) Attempt() (i int32) {
	i = 1
	if val, ok := message.Messages[0].Headers[HeaderAttempt]; ok {
		if val2, ok2 := val.(int32); ok2 {
			i = val2
		}
	}
	return i
}

func (message Message) FirstSeen() (i int64) {
	i = 0
	if val, ok := message.Messages[0].Headers[HeaderFirstSeen]; ok {
		if val2, ok2 := val.(int64); ok2 {
			i = val2
		}
	}
	return i
}

func (message Message) FirstQueue() (i string) {
	i = ""
	if val, ok := message.Messages[0].Headers[HeaderLastSeen]; ok {
		if val2, ok2 := val.(string); ok2 {
			i = val2
		}
	}
	return i
}

func (message Message) LastSeen() (i int64) {
	i = 0
	if val, ok := message.Messages[0].Headers[HeaderFirstQueue]; ok {
		if val2, ok2 := val.(int64); ok2 {
			i = val2
		}
	}
	return i
}

func (message Message) LastQueue() (i string) {
	i = ""
	if val, ok := message.Messages[0].Headers[HeaderLastQueue]; ok {
		if val2, ok2 := val.(string); ok2 {
			i = val2
		}
	}
	return i
}
