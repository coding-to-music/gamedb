package queue

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/streadway/amqp"
)

const (
	namespace = "STEAM_"

	ChangeQueue  = "Changes"
	AppQueue     = "Apps"
	PackageQueue = "Packages"
	PlayerQueue  = "Players"
)

var (
	queues          map[string]queue
	enableConsumers bool = true
)

func init() {

	qs := []queue{
		{Name: ChangeQueue, Callback: processChange},
		{Name: AppQueue, Callback: processApp},
		{Name: PackageQueue, Callback: processPackage},
		{Name: PlayerQueue, Callback: processPlayer},
	}

	queues = make(map[string]queue)
	for _, v := range qs {
		queues[v.Name] = v
	}
}

func RunConsumers() {

	for _, v := range queues {
		go v.consume()
	}
}

// todo, use interface so we can set the payload time in here?
func Produce(queue string, data []byte) (err error) {

	if val, ok := queues[queue]; ok {
		return val.produce(data)
	}

	return errors.New("no such queue")
}

type queue struct {
	Name     string
	Callback func(msg amqp.Delivery) (ack bool, requeue bool)
}

func (s queue) getConnection() (conn *amqp.Connection, ch *amqp.Channel, q amqp.Queue, closeChannel chan *amqp.Error, err error) {

	closeChannel = make(chan *amqp.Error)

	conn, err = amqp.Dial(os.Getenv("STEAM_AMQP"))
	conn.NotifyClose(closeChannel)
	if err != nil {
		logger.Error(err)
	}

	ch, err = conn.Channel()
	if err != nil {
		logger.Error(err)
	}

	q, err = ch.QueueDeclare(namespace+s.Name, true, false, false, false, nil)
	if err != nil {
		logger.Error(err)
	}

	return conn, ch, q, closeChannel, err
}

func (s queue) produce(data []byte) (err error) {

	conn, ch, q, _, err := s.getConnection()
	defer conn.Close()
	defer ch.Close()
	if err != nil {
		return err
	}

	err = ch.Publish("", q.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         data,
		Headers: amqp.Table{
			"trys": 0,
		},
	})
	if err != nil {
		logger.Error(err)
	}

	return nil

}

func (s queue) consume() {

	var breakFor = false

	for {
		conn, ch, q, closeChan, err := s.getConnection()
		if err != nil {
			logger.Error(err)
			return
		}

		msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
		if err != nil {
			logger.Error(err)
			return
		}

		for {
			select {
			case err = <-closeChan:
				breakFor = true
				break

			case msg := <-msgs:

				// Increment consumed header
				if msg.Headers == nil {
					msg.Headers = amqp.Table{}
				}

				trys, ok := msg.Headers["trys"].(int)
				if ok {
					msg.Headers["trys"] = trys + 1
				} else {
					msg.Headers["trys"] = 1
				}

				if trys > 1 {
					fmt.Println(trys)
				}

				// Process message
				ack, requeue := s.Callback(msg)
				if ack {
					msg.Ack(false)
				} else {

					if requeue {
						go func() {
							time.Sleep(time.Second * 10 * time.Duration(trys)) // Exponential backoff
							msg.Nack(false, requeue)
						}()
					} else {
						msg.Nack(false, requeue)
					}
				}
			}

			if breakFor {
				break
			}
		}

		conn.Close()
		ch.Close()
	}
}
