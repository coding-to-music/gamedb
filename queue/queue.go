package queue

import (
	"errors"
	"os"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/streadway/amqp"
)

const (
	enableConsumers  = true
	Namespace        = "Steam_"
	UpdaterNamespace = Namespace + "Updater_"
	headerTry        = "try"
)

const (
	QueueApps         = "Updater_Apps"
	QueueAppsData     = "Updater_Apps_Data"
	QueuePackages     = "Updater_Packages"
	QueuePackagesData = "Updater_Packages_Data"
	QueueChanges      = "Updater_Changes"
	QueuePlayers      = "Players"
	QueueDelays       = "Delays"
)

var (
	queues map[string]queue
)

func init() {

	qs := []queue{
		{Name: QueueChanges, Callback: processChange},
		//{PICSName: QueueAppsData, Callback: processApp},
		{Name: QueuePackagesData, Callback: processPackage},
		//{PICSName: QueuePlayers, Callback: processPlayer},
		{Name: QueueDelays, Callback: processDelay},
	}

	queues = make(map[string]queue)
	for _, v := range qs {
		queues[v.Name] = v
	}
}

func RunConsumers() {

	if enableConsumers {
		for _, v := range queues {
			go v.consume()
		}
	}
}

type ProduceOptions struct {
	QueueConstant string
	Data          []byte
	Try           uint32
}

func Produce(queue string, data []byte) (err error) {

	if val, ok := queues[queue]; ok {
		return val.produce()
	}

	return errors.New("no such queue")
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

	q, err = ch.QueueDeclare(Namespace+s.Name, true, false, false, false, nil)
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
						requeueMessage(msg, q.Name)
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
