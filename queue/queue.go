package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/streadway/amqp"
)

const (
	QueueApps         = "Steam_Apps"
	QueueAppsData     = "Steam_Apps_Data"
	QueueChangesData  = "Steam_Changes_Data"
	QueueDelaysData   = "Steam_Delays_Data"
	QueuePackages     = "Steam_Packages"
	QueuePackagesData = "Steam_Packages_Data"
	QueueProfiles     = "Steam_Profiles"
	QueueProfilesData = "Steam_Profiles_Data"
)

var (
	queues = map[string]queueInterface{}

	errInvalidQueue = errors.New("invalid queue")
	errEmptyMessage = errors.New("empty message")
)

type queueInterface interface {
	getQueueName() (string)
	getRetryData() (RabbitMessageDelay)
	process(msg amqp.Delivery) (ack bool, requeue bool, err error)
	consume()
	produce(data []byte) (err error)
}

func init() {
	qs := []queueInterface{
		RabbitMessageChanges{},
		//RabbitMessageDelay{},
		//RabbitMessagePackage{},
	}

	for _, v := range qs {
		queues[v.getQueueName()] = v
	}
}

func RunConsumers() {

	for _, v := range queues {

		fmt.Println(v.getQueueName() + "X")

		go v.consume()
	}
}

func Produce(queue string, data []byte) (err error) {

	if val, ok := queues[queue]; ok {
		return val.produce(data)
	}

	return errInvalidQueue
}

type baseQueue struct {
	Attempt   int
	StartTime time.Time // Time first placed in queues
	EndTime   time.Time // Time to retry from delay queue
}

func (s baseQueue) getConnection() (conn *amqp.Connection, ch *amqp.Channel, q amqp.Queue, closeChannel chan *amqp.Error, err error) {

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

	q, err = ch.QueueDeclare(s.getQueueName(), true, false, false, false, nil)
	if err != nil {
		logger.Error(err)
	}

	return conn, ch, q, closeChannel, err
}

func (s baseQueue) produce(data []byte) (err error) {

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
	})
	if err != nil {
		logger.Error(err)
	}

	return nil
}

func (s baseQueue) consume() {

	fmt.Println(s.StartTime)

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

				// todo, send to channel to pickup on process?
				//if s.chanx == nil {
				//	s.chanx = make(chan amqp.Delivery)
				//}

				ack, requeue, err := s.process(msg)
				if err != nil {
					logger.Error(err)
				}

				if ack {
					msg.Ack(false)
				} else {

					if requeue {
						err = s.requeueMessage(msg)
						if err != nil {
							logger.Error(err)
						}
					}

					msg.Nack(false, false)
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

func (s baseQueue) requeueMessage(msg amqp.Delivery) error {

	delayMessage := RabbitMessageDelay{}
	delayMessage.Attempt = 1
	delayMessage.StartTime = time.Now()
	delayMessage.SetEndTime()
	delayMessage.OriginalMessage = string(msg.Body)

	data, err := json.Marshal(delayMessage)
	if err != nil {
		return err
	}

	Produce(QueueDelaysData, data)

	return nil
}
