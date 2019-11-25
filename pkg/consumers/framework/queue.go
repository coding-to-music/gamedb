package framework

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type queueName string

type queueHandler func(message message)

type queue struct {
	connection    *connection
	queue         *amqp.Queue
	channel       *amqp.Channel
	closeChan     chan *amqp.Error
	handler       queueHandler
	name          queueName
	prefetchCount int
	batchSize     int
	sync.Mutex
}

func NewQueue(connection *connection, name queueName, prefetchCount int, batchSize int, handler queueHandler) (*queue, error) {

	queue := &queue{
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

func (queue *queue) connect() error {

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

func (queue *queue) produce(message message) error {

	// Sort headers
	if message.message == nil {
		message.message = &amqp.Delivery{}
	}

	if message.message.Headers == nil {
		message.message.Headers = amqp.Table{}
	}

	attempt, ok := message.message.Headers[headerAttempt]
	if ok {
		message.message.Headers[headerAttempt] = attempt.(int) + 1
	} else {
		message.message.Headers[headerAttempt] = 1
	}

	_, ok = message.message.Headers[headerFirstSeen]
	if !ok {
		message.message.Headers[headerFirstSeen] = time.Now().Unix()
	}

	message.message.Headers[headerLastSeen] = time.Now().Unix()

	_, ok = message.message.Headers[headerFirstQueue]
	if !ok {
		message.message.Headers[headerFirstQueue] = queue
	}

	message.message.Headers[headerLastQueue] = time.Now().Unix()

	oldForce, ok := message.message.Headers[headerForce]
	if ok {
		message.message.Headers[headerForce] = oldForce
	} else {
		message.message.Headers[headerForce] = false
	}

	//
	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return queue.channel.Publish("", queue.name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         b,
	})
}

func (queue *queue) consume() error {

	name := config.Config.Environment.Get() + "-" + config.GetSteamKeyTag()

	msgs, err := queue.channel.Consume(queue.queue.Name, name, false, false, false, false, nil)
	if err != nil {
		return err
	}

	// In a anon function so can return at anytime
	go func(msgs <-chan amqp.Delivery) {

		var msgSlice []message

		for {
			msg := <-msgs

			message := message{}
			message.message = &msg
			message.queue = queue

			if len(msgSlice) >= queue.batchSize {

				queue.processMessages(msgSlice)
				msgSlice = []amqp.Delivery{}
			}
		}

	}(msgs)
}
