package framework

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
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
	isOpen        bool
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

				queue.isOpen = false

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

	if queue.isOpen {
		return nil
	}

	operation := func() (err error) {

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

		queue.isOpen = true

		return nil
	}

	policy := backoff.NewExponentialBackOff()
	policy.MaxElapsedTime = 0
	policy.InitialInterval = 5 * time.Second

	return backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err) })
}

func (queue *Queue) Produce(message Message) error {

	// Headers
	for _, message := range message.messages {

		message.Headers = queue.prepareHeaders(message.Headers)

		err := queue.channel.Publish("", string(queue.name), false, false, amqp.Publishing{
			Headers:      message.Headers,
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         message.Body,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (queue *Queue) ProduceInterface(message interface{}) error {

	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return queue.channel.Publish("", string(queue.name), false, false, amqp.Publishing{
		Headers:      queue.prepareHeaders(nil),
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         b,
	})
}

const (
	headerAttempt    = "attempt"
	headerFirstSeen  = "first-seen"
	headerLastSeen   = "last-seen"
	headerFirstQueue = "first-queue"
	headerLastQueue  = "last-queue"
	headerForce      = "force"
)

func (queue Queue) prepareHeaders(headers amqp.Table) amqp.Table {

	if headers == nil {
		headers = amqp.Table{}
	}

	//
	attemptSet := false
	attempt, ok := headers[headerAttempt]
	if ok {
		if val, ok2 := attempt.(int32); ok2 {
			headers[headerAttempt] = val + 1
			attemptSet = true
		}
	}
	if !attemptSet {
		headers[headerAttempt] = 1
	}

	//
	_, ok = headers[headerFirstSeen]
	if !ok {
		headers[headerFirstSeen] = time.Now().Unix()
	}

	//
	headers[headerLastSeen] = time.Now().Unix()

	//
	_, ok = headers[headerFirstQueue]
	if !ok {
		headers[headerFirstQueue] = queue.name
	}

	//
	headers[headerLastQueue] = queue.name

	//
	oldForce, ok := headers[headerForce]
	if ok {
		headers[headerForce] = oldForce
	} else {
		headers[headerForce] = false
	}

	return headers
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
			if !queue.connection.connection.IsClosed() && queue.isOpen {
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
		}

	}(msgs)

	return nil
}
