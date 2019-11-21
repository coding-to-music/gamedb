package queue2

import (
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

func NewConnection(config amqp.Config) (*connection, error) {

	conn := &connection{
		config:    config,
		closeChan: make(chan *amqp.Error),
	}

	return conn, conn.connect()
}

type connection struct {
	conn      *amqp.Connection
	config    amqp.Config
	closeChan chan *amqp.Error
}

func (connection *connection) connect() error {

	log.Info("Creating Rabbit connection")

	operation := func() (err error) {

		connection.conn, err = amqp.DialConfig(config.RabbitDSN(), connection.config)
		if err != nil {
			return nil
		}

		connection.conn.NotifyClose(connection.closeChan)

		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.MaxElapsedTime = 0
	policy.InitialInterval = 5 * time.Second

	return backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err) })
}
