package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/spf13/viper"
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
	queues = map[string]rabbitMessageBase{}

	errInvalidQueue = errors.New("invalid queue")
	errEmptyMessage = errors.New("empty message")

	dsn string
)

type queueInterface interface {
	getQueueName() (string)
	getRetryData() (RabbitMessageDelay)
	process(msg amqp.Delivery) (ack bool, requeue bool, err error)
}

func init() {
	qs := []rabbitMessageBase{
		//{Message: RabbitMessageApp{}},
		{Message: RabbitMessageChanges{}},
		//{Message: RabbitMessageDelay{}},
		//{Message: RabbitMessagePackage{}},
		//{Message: RabbitMessageProfile{}},
	}

	for _, v := range qs {
		queues[v.Message.getQueueName()] = v
	}
}

func Init() {

	username := viper.GetString("RABBIT_USER")
	password := viper.GetString("RABBIT_PASS")
	host := viper.GetString("RABBIT_HOST")
	port := viper.GetString("RABBIT_PORT")

	dsn = "amqp://" + username + ":" + password + "@" + host + ":" + port

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

type rabbitMessageBase struct {
	Attempt   int
	StartTime time.Time // Time first placed in delay queue
	EndTime   time.Time // Time to retry from delay queue
	Message   queueInterface
}

func (s rabbitMessageBase) getConnection() (conn *amqp.Connection, ch *amqp.Channel, q amqp.Queue, closeChannel chan *amqp.Error, err error) {

	closeChannel = make(chan *amqp.Error)

	conn, err = amqp.Dial(dsn)
	conn.NotifyClose(closeChannel)
	if err != nil {
		logger.Error(err)
	}

	ch, err = conn.Channel()
	if err != nil {
		logger.Error(err)
	}

	q, err = ch.QueueDeclare(s.Message.getQueueName(), true, false, false, false, nil)
	if err != nil {
		logger.Error(err)
	}

	return conn, ch, q, closeChannel, err
}

func (s rabbitMessageBase) produce(data []byte) (err error) {

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

func (s rabbitMessageBase) consume() {

	fmt.Println(s.Message.getQueueName())

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

				ack, requeue, err := s.Message.process(msg)
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

func (s rabbitMessageBase) requeueMessage(msg amqp.Delivery) error {

	delayeMessage := rabbitMessageBase{
		Attempt:   s.Attempt,
		StartTime: s.StartTime,
		EndTime:   s.EndTime,
		Message: RabbitMessageDelay{
			OriginalMessage: string(msg.Body),
			OriginalQueue:   s.Message.getQueueName(),
		},
	}

	delayeMessage.IncrementAttempts()

	data, err := json.Marshal(delayeMessage)
	if err != nil {
		return err
	}

	Produce(QueueDelaysData, data)

	return nil
}

func (s *rabbitMessageBase) IncrementAttempts() {

	s.Attempt++

	// Update end time
	var min float64 = 1
	var max float64 = 600

	var seconds = math.Pow(1.3, float64(s.Attempt))
	var minmaxed = math.Min(min+seconds, max)
	var rounded = math.Round(minmaxed)

	s.EndTime = s.StartTime.Add(time.Second * time.Duration(rounded))
}
