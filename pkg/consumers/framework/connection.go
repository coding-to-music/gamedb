package framework

import (
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

const (
	Consumer = "c"
	Producer = "p"
)

type Connection struct {
	connection *amqp.Connection
	name       string
	config     amqp.Config
	closeChan  chan *amqp.Error
	sync.Mutex
}

func NewConnection(name string, config amqp.Config) (c Connection, err error) {

	connection := Connection{
		config:    config,
		name:      name,
		closeChan: make(chan *amqp.Error),
	}

	err = connection.connect()
	if err != nil {
		return c, err
	}

	go func() {
		for {

			var err error
			var open bool

			select {
			case err, open = <-connection.closeChan:

				if open {
					log.Warning("Rabbit connection closed", err)
				} else {
					log.Warning("Rabbit connection closed")
				}

				time.Sleep(time.Second * 10)

				err = connection.connect()
				log.Err("Connection connecting", err)
			}
		}
	}()

	return connection, nil
}

func (connection *Connection) connect() error {

	connection.Lock()
	defer connection.Unlock()

	if connection.connection != nil && !connection.connection.IsClosed() {
		return nil
	}

	log.Info("Creating Rabbit connection")

	operation := func() (err error) {

		connection.connection, err = amqp.DialConfig(config.RabbitDSN(), connection.config)
		if err != nil {
			return err
		}

		_ = connection.connection.NotifyClose(connection.closeChan)

		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.MaxElapsedTime = 0
	policy.InitialInterval = 5 * time.Second

	return backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err) })
}
