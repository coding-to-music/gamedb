package consumers

import (
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type connection struct {
	connection *amqp.Connection
	config     amqp.Config
	closeChan  chan *amqp.Error
	sync.Mutex
}

func NewConnection(config amqp.Config) (*connection, error) {

	conn := &connection{
		config:    config,
		closeChan: make(chan *amqp.Error),
	}

	err := conn.connect()
	if err != nil {
		return nil, err
	}

	conn.listen()

	return conn, nil
}

func (connection *connection) connect() error {

	connection.Lock()
	defer connection.Unlock()

	if !connection.connection.IsClosed() {
		return nil
	}

	log.Info("Creating Rabbit connection")

	operation := func() (err error) {

		connection.connection, err = amqp.DialConfig(config.RabbitDSN(), connection.config)
		if err != nil {
			return nil
		}

		connection.connection.NotifyClose(connection.closeChan)

		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.MaxElapsedTime = 0
	policy.InitialInterval = 5 * time.Second

	return backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err) })
}

func (connection *connection) listen() {
	go func() {
		for {
			var err error
			select {
			case err = <-connection.closeChan:

				log.Warning("Rabbit connection closed", err)

				time.Sleep(time.Second * 10)

				err = connection.connect()
				log.Err(err)
			}
		}
	}()
}
