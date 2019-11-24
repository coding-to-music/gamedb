package consumers

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type queue struct {
	connection    *connection
	queue         *amqp.Queue
	channel       *amqp.Channel
	closeChan     chan *amqp.Error
	name          string
	prefetchCount int
	batchSize     int
	sync.Mutex
}

func NewQueue(connection *connection, name string, prefetchCount int, batchSize int) (*queue, error) {

	qu := &queue{
		connection:    connection,
		name:          name,
		prefetchCount: prefetchCount,
		batchSize:     batchSize,
	}

	return qu, qu.connect()
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

		qu, err := queue.channel.QueueDeclare(queue.name, true, false, false, false, nil)
		if err != nil {
			return err
		}

		queue.queue = &qu
	}

	return nil
}

func (queue *queue) listen() {
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
}

func (queue *queue) produce(message message) error {

	message.setupHeadersForProduce(queue.name, true)

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

		var msgSlice []amqp.Delivery

		for {
			select {
			case msg := <-msgs:
				msgSlice = append(msgSlice, msg)
			}

			if len(msgSlice) >= queue.batchSize {

				queue.processMessages(msgSlice)
				msgSlice = []amqp.Delivery{}
			}
		}

	}(msgs)
}
