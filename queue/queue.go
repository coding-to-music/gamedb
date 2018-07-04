package queue

import (
	"encoding/json"
	"errors"
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
	queues map[string]queue

	errInvalidQueue = errors.New("invalid queue")
)

func init() {

	qs := []queue{
		{Name: QueueChangesData, Callback: processChange},
		//{PICSName: QueueAppsData, Callback: processApp},
		{Name: QueuePackagesData, Callback: processPackage},
		//{PICSName: QueueProfiles, Callback: processPlayer},
		{Name: QueueDelaysData, Callback: processDelay},
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

	return errInvalidQueue
}

type queue struct {
	Name     string
	Callback func(msg amqp.Delivery) (ack bool, requeue bool, err error)
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

	q, err = ch.QueueDeclare(s.Name, true, false, false, false, nil)
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

				ack, requeue, err := s.Callback(msg)
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

func (s queue) requeueMessage(msg amqp.Delivery) error {

	delayMessage := RabbitMessageDelay{}
	delayMessage.Attempt = 1
	delayMessage.StartTime = time.Now()
	delayMessage.SetEndTime()
	delayMessage.Queue = s.Name
	delayMessage.Message = string(msg.Body)

	data, err := json.Marshal(delayMessage)
	if err != nil {
		return err
	}

	Produce(QueueDelaysData, data)

	return nil
}
