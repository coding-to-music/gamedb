package framework

import (
	"sync"
	"time"

	"github.com/cenkalti/backoff/v3"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type ConnType string

const (
	Consumer ConnType = "consumer"
	Producer ConnType = "producer"
)

type Connection struct {
	dial       string
	connection *amqp.Connection
	config     amqp.Config
	closeChan  chan *amqp.Error
	connType   ConnType
	sync.Mutex
}

func NewConnection(dial string, conType ConnType, config amqp.Config) (c *Connection, err error) {

	connection := &Connection{
		dial:     dial,
		config:   config,
		connType: conType,
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
				if err != nil {
					log.Err("Failed to reconnect connection", err, log.OptionNoStack)
				}
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

	log.Info("Creating Rabbit connection (" + connection.connType + ")")

	operation := func() (err error) {

		// Connect
		connection.connection, err = amqp.DialConfig(connection.dial, connection.config)
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

	err := backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Warning("Trying to connect to Rabbit", err) })
	if err == nil {
		log.Info("Rabbit conn connected (" + connection.connType + ")")
	}
	return err
}
