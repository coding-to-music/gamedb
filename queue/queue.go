package queue

import (
	"encoding/json"
	"errors"
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

	consumerConnection   *amqp.Connection
	consumerCloseChannel chan *amqp.Error

	producerConnection   *amqp.Connection
	producerCloseChannel chan *amqp.Error
)

type queueInterface interface {
	getQueueName() (string)
	getRetryData() (RabbitMessageDelay)
	process(msg amqp.Delivery) (ack bool, requeue bool, err error)
}

func init() {

	consumerCloseChannel = make(chan *amqp.Error)
	producerCloseChannel = make(chan *amqp.Error)

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
	Message   queueInterface
	Attempt   int
	StartTime time.Time // Time first placed in delay queue
	EndTime   time.Time // Time to retry from delay queue
}

func (s rabbitMessageBase) getQueue(conn *amqp.Connection) (ch *amqp.Channel, qu amqp.Queue, err error) {

	ch, err = conn.Channel()
	if err != nil {
		logger.Error(err)
	}

	qu, err = ch.QueueDeclare(s.Message.getQueueName(), true, false, false, false, nil)
	if err != nil {
		logger.Error(err)
	}

	return ch, qu, err
}

func (s rabbitMessageBase) produce(data []byte) (err error) {

	logger.Info("Producing to: " + s.Message.getQueueName())

	// Connect
	if producerConnection == nil {

		producerConnection, err = amqp.Dial(dsn)
		producerConnection.NotifyClose(producerCloseChannel)
		if err != nil {
			return err
		}
	}

	//
	ch, qu, err := s.getQueue(producerConnection)
	defer ch.Close()
	if err != nil {
		return err
	}

	err = ch.Publish("", qu.Name, false, false, amqp.Publishing{
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

	logger.Info("Consuming from: " + s.Message.getQueueName())

	var breakFor = false
	var err error

	for {

		// Connect
		if consumerConnection == nil {

			consumerConnection, err = amqp.Dial(dsn)
			consumerConnection.NotifyClose(consumerCloseChannel)
			if err != nil {
				logger.Error(err)
				return
			}
		}

		//
		ch, qu, err := s.getQueue(consumerConnection)
		if err != nil {
			logger.Error(err)
			return
		}

		msgs, err := ch.Consume(qu.Name, "", false, false, false, false, nil)
		if err != nil {
			logger.Error(err)
			return
		}

		for {
			select {
			case err = <-consumerCloseChannel:
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

		//conn.Close()
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

	// Increment attemp
	s.Attempt++

	// Update end time
	var min float64 = 1
	var max float64 = 600

	var seconds = math.Pow(1.3, float64(s.Attempt))
	var minmaxed = math.Min(min+seconds, max)
	var rounded = math.Round(minmaxed)

	s.EndTime = s.StartTime.Add(time.Second * time.Duration(rounded))
}
