package framework

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type (
	QueueName string
	Handler   func(message []Message)
)

type Channel struct {
	Name          QueueName
	connection    Connection
	channel       *amqp.Channel
	closeChan     chan *amqp.Error
	handler       Handler
	isOpen        bool
	prefetchCount int
	batchSize     int
	updateHeaders bool
	sync.Mutex
}

func NewChannel(connection Connection, name QueueName, prefetchCount int, batchSize int, handler Handler, updateHeaders bool) (c Channel, err error) {

	channel := Channel{
		connection:    connection,
		Name:          name,
		prefetchCount: prefetchCount,
		batchSize:     batchSize,
		closeChan:     make(chan *amqp.Error),
		handler:       handler,
		updateHeaders: updateHeaders,
	}

	err = channel.connect()
	if err != nil {
		return c, err
	}

	go func() {
		for {
			var err error
			// var open bool
			select {
			case err, _ = <-channel.closeChan:

				// if open {
				// 	log.Warning("Rabbit channel closed", err)
				// } else {
				// 	channel.isOpen = false
				// 	log.Warning("Rabbit channel closed")
				// }

				time.Sleep(time.Second * 10)

				err = channel.connect()
				log.Err("Channel connecting", err)
			}
		}
	}()

	return channel, nil
}

func (channel *Channel) connect() error {

	channel.Lock()
	defer channel.Unlock()

	if channel.isOpen {
		return nil
	}

	operation := func() (err error) {

		if channel.channel == nil {

			c, err := channel.connection.connection.Channel()
			if err != nil {
				return err
			}

			err = c.Qos(channel.prefetchCount, 0, false)
			if err != nil {
				return err
			}

			_ = c.NotifyClose(channel.closeChan)

			channel.channel = c
		}

		_, err = channel.channel.QueueDeclare(string(channel.Name), true, false, false, false, nil)
		if err != nil {
			return err
		}

		channel.isOpen = true

		return nil
	}

	policy := backoff.NewExponentialBackOff()
	policy.MaxElapsedTime = 0
	policy.InitialInterval = 5 * time.Second

	return backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err) })
}

func (channel *Channel) Produce(message Message) error {

	if channel == nil {
		return errors.New("queue has been removed")
	}

	// Headers
	if channel.updateHeaders {
		message.Message.Headers = channel.prepareHeaders(message.Message.Headers)
	}

	//
	return channel.channel.Publish("", string(channel.Name), false, false, amqp.Publishing{
		Headers:      message.Message.Headers,
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         message.Message.Body,
	})
}

func (channel *Channel) ProduceInterface(message interface{}) error {

	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	headers := amqp.Table{}
	if channel.updateHeaders {
		headers = channel.prepareHeaders(headers)
	}

	return channel.channel.Publish("", string(channel.Name), false, false, amqp.Publishing{
		Headers:      headers,
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         b,
	})
}

func (channel Channel) prepareHeaders(headers amqp.Table) amqp.Table {

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
		headers[HeaderFirstQueue] = string(channel.Name)
	}

	//
	headers[HeaderLastQueue] = string(channel.Name)

	return headers
}

func (channel *Channel) Consume() error {

	tag := config.Config.Environment.Get() + "-" + config.GetSteamKeyTag()

	msgs, err := channel.channel.Consume(string(channel.Name), tag, false, false, false, false, nil)
	if err != nil {
		return err
	}

	// In a anon function so can return at anytime
	go func(msgs <-chan amqp.Delivery) {

		var messages []Message

		for {
			select {
			case msg, open := <-msgs:
				if open && !channel.connection.connection.IsClosed() && channel.isOpen {
					messages = append(messages, Message{
						Channel: *channel,
						Message: &msg,
					})
				}
			}

			if len(messages) >= channel.batchSize {

				if channel.handler != nil {
					channel.handler(messages)
					messages = nil
				}
			}
		}

	}(msgs)

	return nil
}
