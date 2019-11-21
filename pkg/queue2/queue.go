package queue2

import (
	"github.com/streadway/amqp"
)

func NewQueue(connection connection) {

}

type queue struct {
	queue   amqp.Queue
	channel amqp.Channel
}
