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

type Handler func(message []Message)

type Queue struct {
	Name          QueueName
	connection    *Connection
	queue         *amqp.Queue
	channel       *amqp.Channel
	closeChan     chan *amqp.Error
	handler       Handler
	isOpen        bool
	prefetchCount int
	batchSize     int
	updateHeaders bool
	sync.Mutex
}

func NewQueue(connection *Connection, name QueueName, prefetchCount int, batchSize int, handler Handler, updateHeaders bool) (*Queue, error) {

	queue := &Queue{
		connection:    connection,
		Name:          name,
		prefetchCount: prefetchCount,
		batchSize:     batchSize,
		closeChan:     make(chan *amqp.Error),
		handler:       handler,
		updateHeaders: updateHeaders,
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

			channel, err := queue.connection.connection.Channel()
			if err != nil {
				return err
			}

			err = channel.Qos(queue.prefetchCount, 0, false)
			if err != nil {
				return err
			}

			_ = channel.NotifyClose(queue.closeChan)

			queue.channel = channel
		}

		if queue.queue == nil {

			qu, err := queue.channel.QueueDeclare(string(queue.Name), true, false, false, false, nil)
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
	if queue.updateHeaders {
		message.Message.Headers = queue.prepareHeaders(message.Message.Headers)
	}

	//
	return queue.channel.Publish("", string(queue.Name), false, false, amqp.Publishing{
		Headers:      message.Message.Headers,
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         message.Message.Body,
	})
}

func (queue *Queue) ProduceInterface(message interface{}) error {

	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	headers := amqp.Table{}
	if queue.updateHeaders {
		headers = queue.prepareHeaders(headers)
	}

	return queue.channel.Publish("", string(queue.Name), false, false, amqp.Publishing{
		Headers:      headers,
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         b,
	})
}

func (queue Queue) prepareHeaders(headers amqp.Table) amqp.Table {

	if headers == nil {
		headers = amqp.Table{}
	}

	//
	attemptSet := false
	attempt, ok := headers[HeaderAttempt]
	if ok {
		if val, ok2 := attempt.(int32); ok2 {
			headers[HeaderAttempt] = val + 1
			attemptSet = true
		}
	}
	if !attemptSet {
		headers[HeaderAttempt] = 1
	}

	//
	_, ok = headers[HeaderFirstSeen]
	if !ok {
		headers[HeaderFirstSeen] = time.Now().Unix()
	}

	//
	headers[HeaderLastSeen] = time.Now().Unix()

	//
	_, ok = headers[HeaderFirstQueue]
	if !ok {
		headers[HeaderFirstQueue] = string(queue.Name)
	}

	//
	headers[HeaderLastQueue] = string(queue.Name)

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

		var messages []Message

		for {
			if !queue.connection.connection.IsClosed() && queue.isOpen {
				select {
				case msg := <-msgs:
					messages = append(messages, Message{
						Queue:   queue,
						Message: &msg,
					})
				}

				if len(messages) >= queue.batchSize {

					if queue.handler != nil {
						queue.handler(messages)
						messages = nil
					}
				}
			}
		}

	}(msgs)

	return nil
}
