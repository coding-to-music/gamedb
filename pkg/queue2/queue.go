package queue2

import (
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/streadway/amqp"
)

type queue struct {
	connection    *connection
	queue         *amqp.Queue
	channel       *amqp.Channel
	name          string
	prefetchCount int
	sync.Mutex
}

func NewQueue(connection *connection, name string, prefetchCount int) (*queue, error) {

	qu := &queue{
		connection:    connection,
		name:          name,
		prefetchCount: prefetchCount,
	}

	return qu, qu.init()
}

func (queue *queue) init() error {

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

			if len(msgSlice) >= q.batchSize {

				switch v := q.queue.(type) {
				case *steamQueue:
					v.SteamClient = q.SteamClient
					q.queue = v
				case *delayQueue:
					v.BaseQueue = q
					q.queue = v
				}

				q.queue.processMessages(msgSlice)
				msgSlice = []amqp.Delivery{}
			}
		}

	}(msgs)
}
