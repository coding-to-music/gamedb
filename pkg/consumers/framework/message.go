package framework

import (
	"sync"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type Message struct {
	Channel     Channel
	Message     *amqp.Delivery
	actionTaken bool
	sync.Mutex
}

func (message *Message) Ack() {
	message.ack(false)
}

func (message *Message) AckMultiple() {
	message.ack(true)
}

func (message *Message) ack(multiple bool) {

	message.Lock()
	defer message.Unlock()

	if message.actionTaken {
		return
	}

	err := message.Message.Ack(multiple)
	if err != nil {
		log.Err(err)
	} else {
		message.actionTaken = true
	}
}

func (message Message) SendToQueue(channels ...Channel) {

	// Send to back of current queue if none specified
	if len(channels) == 0 {
		channels = []Channel{message.Channel}
	}

	//
	var err error

	for _, channel := range channels {
		err = channel.Produce(message)
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
	found := false
	if val, ok := message.Message.Headers[HeaderAttempt]; ok {
		if val2, ok2 := val.(int32); ok2 {
			i = val2
			found = true
		}
	}

	// Write it onto the message so it's acurate when
	// produced by something outside of this library
	if !found {
		message.Message.Headers[HeaderAttempt] = i
	}

	return i
}

func (message Message) FirstSeen() (i time.Time) {
	if val, ok := message.Message.Headers[HeaderFirstSeen]; ok {
		if val2, ok2 := val.(int64); ok2 {
			i = time.Unix(val2, 0)
		}
	}
	return i
}

func (message Message) FirstQueue() (i QueueName) {
	i = ""
	if val, ok := message.Message.Headers[HeaderLastSeen]; ok {
		if val2, ok2 := val.(string); ok2 {
			i = QueueName(val2)
		}
	}
	return i
}

func (message Message) LastSeen() (i time.Time) {
	if val, ok := message.Message.Headers[HeaderFirstQueue]; ok {
		if val2, ok2 := val.(int64); ok2 {
			i = time.Unix(val2, 0)
		}
	}
	return i
}

func (message Message) LastQueue() (i QueueName) {
	i = ""
	if val, ok := message.Message.Headers[HeaderLastQueue]; ok {
		if val2, ok2 := val.(string); ok2 {
			i = QueueName(val2)
		}
	}
	return i
}
