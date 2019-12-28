package framework

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type (
	QueueName string
	Handler   func(message []*Message)
)

type Channel struct {
	Name          QueueName
	connection    *Connection
	channel       *amqp.Channel
	closeChan     chan *amqp.Error
	handler       Handler
	isOpen        bool
	prefetchCount int
	batchSize     int
	updateHeaders bool
	connectLock   sync.Mutex
}

func NewChannel(connection *Connection, name QueueName, prefetchCount int, batchSize int, handler Handler, updateHeaders bool) (c *Channel, err error) {

	channel := &Channel{
		connection:    connection,
		Name:          name,
		prefetchCount: prefetchCount,
		batchSize:     batchSize,
		handler:       handler,
		updateHeaders: updateHeaders,
	}

	err = channel.connect()
	if err != nil {
		return c, err
	}

	// For producer channels
	if connection.connType == Producer {
		go func() {
			for {
				select {
				case amqpErr, _ := <-channel.closeChan:
					channel.onDisconnect(amqpErr)
				}
			}
		}()
	}

	return channel, nil
}

func (channel *Channel) connect() error {

	channel.connectLock.Lock()
	defer channel.connectLock.Unlock()

	if channel.isOpen {
		return nil
	}

	operation := func() (err error) {

		if channel.connection.connection == nil || channel.connection.connection.IsClosed() {
			return errors.New("waiting for connecting before channel")
		}

		// Connect
		c, err := channel.connection.connection.Channel()
		if err != nil {
			return err
		}

		// Set pre-fetch
		if channel.prefetchCount > 0 {
			err = c.Qos(channel.prefetchCount, 0, false)
			if err != nil {
				return err
			}
		}

		// Set new close channel
		channel.closeChan = make(chan *amqp.Error)
		_ = c.NotifyClose(channel.closeChan)

		channel.channel = c

		// Queue
		_, err = channel.channel.QueueDeclare(string(channel.Name), true, false, false, false, nil)
		if err != nil {
			return err
		}

		//
		channel.isOpen = true

		log.Info("Rabbit chan connected (" + string(channel.connection.connType) + "/" + string(channel.Name) + ")")

		return nil
	}

	policy := backoff.NewExponentialBackOff()
	policy.MaxElapsedTime = 0
	policy.InitialInterval = 5 * time.Second

	return backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Warning("Connecting to channel: ", err) })
}

func (channel *Channel) produceMessage(message *Message) error {

	// Headers
	if channel.updateHeaders {
		message.Message.Headers = channel.prepareHeaders(message.Message.Headers)
	}

	//
	return channel.channel.Publish("", string(channel.Name), false, false, amqp.Publishing{
		Headers:      message.Message.Headers,
		DeliveryMode: message.Message.DeliveryMode,
		ContentType:  message.Message.ContentType,
		Body:         message.Message.Body,
	})
}

func (channel *Channel) Produce(message interface{}) error {

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

func (channel *Channel) onDisconnect(amqpErr *amqp.Error) {

	channel.isOpen = false

	log.Warning("Rabbit channel closed ("+channel.Name+")", amqpErr, log.OptionNoStack)

	err := channel.connect()
	if err != nil {
		log.Err("Failed to reconnect channel", err, log.OptionNoStack)
	}

	time.Sleep(time.Second * 20)
}

func (channel Channel) prepareHeaders(headers amqp.Table) amqp.Table {

	if headers == nil {
		headers = amqp.Table{}
	}

	//
	attemptSet := false
	attempt, ok := headers[HeaderAttempt]
	if ok {
		if val, ok2 := attempt.(int); ok2 {
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

func (channel *Channel) Consume() {

	for {

		if !channel.isOpen || channel.channel == nil {
			log.Info("Can't consume when channel is nil/closed")
			time.Sleep(time.Second * 5)
			continue
		}

		tag := config.Config.Environment.Get() + "-" + config.GetSteamKeyTag()

		msgs, err := channel.channel.Consume(string(channel.Name), tag, false, false, false, false, nil)
		if err != nil {
			log.Err("Getting Rabbit channel chan", err, log.OptionNoStack)
			continue
		}

		// In a anon function so can return at anytime
		func() {

			var messages []*Message

			for {
				select {
				case amqpErr, _ := <-channel.closeChan:

					channel.onDisconnect(amqpErr)
					return

				case msg, open := <-msgs:
					if open && channel.connection.connection != nil && !channel.connection.connection.IsClosed() {
						messages = append(messages, &Message{
							Channel: channel,
							Message: &msg,
						})
					}
				}

				if len(messages) > 0 && len(messages) >= channel.batchSize {

					if channel.handler != nil {

						// Fill in batch info
						for k := range messages {
							messages[k].BatchTotal = len(messages)
							messages[k].BatchItem = k + 1
						}

						//
						channel.handler(messages)
						messages = nil
					}
				}
			}
		}()

		time.Sleep(time.Second * 5)
	}
}

func (channel *Channel) Inspect() (amqp.Queue, error) {

	return channel.channel.QueueInspect(string(channel.Name))
}
