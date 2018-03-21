package queue

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Jleagle/go-helpers/logger"
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

func Produce(queue string, data []byte) (err error) {

	if val, ok := queues[queue]; ok {
		return val.produce(data)
	}

	return errors.New("no such queue")
}

type queue struct {
	Name     string
	Callback func(msg amqp.Delivery) (err error)
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

	err = ch.Publish("", q.Name, false, false, amqp.Publishing{DeliveryMode: amqp.Persistent, ContentType: "application/json", Body: data})
	if err != nil {
		logger.Error(err)
	}

	return nil

}

func (s queue) consume() {

	ticker := time.NewTicker(5 * time.Second)

	go func() {

		var working bool

		for _ := range ticker.C {

			if working == false {
				working = true

				fmt.Println("Getting " + s.Name + " messages")

				conn, ch, q, _, err := s.getConnection()

				msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
				if err != nil {
					logger.Error(err)
				}

				for v := range msgs {

					err := s.Callback(v)
					if err != nil {
						logger.Error(err)
					}

				}

				conn.Close()
				ch.Close()

				working = false
			}
		}
	}()
}
