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

func NewConnection(name string, config amqp.Config) (c *Connection, err error) {

	connection := &Connection{
		config: config,
		name:   name,
	}

	err = connection.connect()
	if err != nil {
		return c, err
	}

	go func() {
		for {
			select {
			case amqpErr, open := <-connection.closeChan:

				connection.connection = nil

				if open {
					log.Warning("Rabbit connection closed", amqpErr)
				} else {
					log.Warning("Rabbit connection closed")
				}

				time.Sleep(time.Second * 10)

				err := connection.connect()
				log.Err("Failed to reconnect connection", err)
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

	log.Info("Creating Rabbit connection: " + connection.name)

	operation := func() (err error) {

		// Connect
		connection.connection, err = amqp.DialConfig(config.RabbitDSN(), connection.config)
		if err != nil {
			return err
		}

		// Set new close channel
		connection.closeChan = make(chan *amqp.Error)
		_ = connection.connection.NotifyClose(connection.closeChan)

		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.MaxElapsedTime = 0
	policy.InitialInterval = 5 * time.Second

	return backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info("Trying to connect to Rabbit", err) })
}
