package framework

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type QueueName string

type Handler func(message Message)

type Queue struct {
	connection    *Connection
	queue         *amqp.Queue
	channel       *amqp.Channel
	closeChan     chan *amqp.Error
	handler       Handler
	name          QueueName
	prefetchCount int
	batchSize     int
	sync.Mutex
}

func NewQueue(connection *Connection, name QueueName, prefetchCount int, batchSize int, handler Handler) (*Queue, error) {

	queue := &Queue{
		connection:    connection,
		name:          name,
		prefetchCount: prefetchCount,
		batchSize:     batchSize,
		closeChan:     make(chan *amqp.Error),
		handler:       handler,
	}

	err := queue.connect()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			var err error
			select {
			case err = <-queue.closeChan:

				log.Warning("Rabbit channel closed", err)

				time.Sleep(time.Second * 10)

				err = queue.connect()
				log.Err(err)
			}
		}
	}()

	return queue, nil
}

func (queue *Queue) connect() error {

	queue.Lock()
	defer queue.Unlock()

	if queue.channel == nil {

		ch, err := queue.connection.connection.Channel()
		if err != nil {
			return err
		}

		err = ch.Qos(queue.prefetchCount, 0, false)
		if err != nil {
			return err
		}

		_ = ch.NotifyClose(queue.closeChan)

		queue.channel = ch
	}

	if queue.queue == nil {

		qu, err := queue.channel.QueueDeclare(string(queue.name), true, false, false, false, nil)
		if err != nil {
			return err
		}

		queue.queue = &qu
	}

	return nil
}

const (
	headerAttempt    = "attempt"
	headerFirstSeen  = "first-seen"
	headerLastSeen   = "last-seen"
	headerFirstQueue = "first-queue"
	headerLastQueue  = "last-queue"
	headerForce      = "force"
)

func (queue *Queue) produce(message Message) error {

	// Headers
	for _, message := range message.messages {

		// if message == nil {
		// 	message = &amqp.Delivery{}
		// }

		if message.Headers == nil {
			message.Headers = amqp.Table{}
		}

		//
		attempt, ok := message.Headers[headerAttempt]
		if ok {
			message.Headers[headerAttempt] = attempt.(int32) + 1
		} else {
			message.Headers[headerAttempt] = 1
		}

		//
		_, ok = message.Headers[headerFirstSeen]
		if !ok {
			message.Headers[headerFirstSeen] = time.Now().Unix()
		}

		//
		message.Headers[headerLastSeen] = time.Now().Unix()

		//
		_, ok = message.Headers[headerFirstQueue]
		if !ok {
			// message.Headers[headerFirstQueue] = queue
		}

		//
		message.Headers[headerLastQueue] = time.Now().Unix()

		//
		oldForce, ok := message.Headers[headerForce]
		if ok {
			message.Headers[headerForce] = oldForce
		} else {
			message.Headers[headerForce] = false
		}

		//
		b, err := json.Marshal(message)
		if err != nil {
			return err
		}

		err = queue.channel.Publish("", string(queue.name), false, false, amqp.Publishing{
			Headers:      message.Headers,
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         b,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (queue *Queue) Consume() error {

	name := config.Config.Environment.Get() + "-" + config.GetSteamKeyTag()

	msgs, err := queue.channel.Consume(queue.queue.Name, name, false, false, false, false, nil)
	if err != nil {
		return err
	}

	// In a anon function so can return at anytime
	go func(msgs <-chan amqp.Delivery) {

		message := Message{}
		message.queue = queue

		for {
			select {
			case msg := <-msgs:
				message.messages = append(message.messages, &msg)
			}

			if len(message.messages) >= queue.batchSize {

				if queue.handler != nil {
					queue.handler(message)
					message.messages = nil
				}
			}
		}

	}(msgs)

	return nil
}
