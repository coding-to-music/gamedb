package queue

import (
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/gamedb/website/logging"
	"github.com/spf13/viper"
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

	rabbitDSN string

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

	user := viper.GetString("RABBIT_USER")
	pass := viper.GetString("RABBIT_PASS")
	host := viper.GetString("RABBIT_HOST")
	port := viper.GetString("RABBIT_PORT")

	rabbitDSN = "amqp://" + user + ":" + pass + "@" + host + ":" + port
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
	logging.Error(err)

	qu, err = ch.QueueDeclare(s.Message.getQueueName(), true, false, false, false, nil)
	logging.Error(err)

	return ch, qu, err
}

func (s rabbitMessageBase) produce(data []byte) (err error) {

	logging.Info("Producing to: " + s.Message.getQueueName())

	// Connect
	if producerConnection == nil {

		producerConnection, err = amqp.Dial(rabbitDSN)
		producerConnection.NotifyClose(producerCloseChannel)
		if err != nil {
			return err
		}
	}

	//
	ch, qu, err := s.getQueue(producerConnection)
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.Publish("", qu.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         data,
	})
	logging.Error(err)

	return nil
}

func (s rabbitMessageBase) consume() {

	logging.InfoL("Consuming from: " + s.Message.getQueueName())

	var breakFor = false
	var err error

	for {

		// Connect
		if consumerConnection == nil {

			consumerConnection, err = amqp.Dial(rabbitDSN)
			consumerConnection.NotifyClose(consumerCloseChannel)
			if err != nil {
				logging.Error(err)
				return
			}
		}

		//
		ch, qu, err := s.getQueue(consumerConnection)
		if err != nil {
			logging.Error(err)
			return
		}

		msgs, err := ch.Consume(qu.Name, "", false, false, false, false, nil)
		if err != nil {
			logging.Error(err)
			return
		}

		for {
			select {
			case err = <-consumerCloseChannel:
				breakFor = true
				break

			case msg := <-msgs:

				ack, requeue, err := s.Message.process(msg)
				logging.Error(err)

				if ack {
					msg.Ack(false)
				} else {

					if requeue {
						err = s.requeueMessage(msg)
						logging.Error(err)
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

type SteamKitJob struct {
	SequentialCount int    `json:"SequentialCount"`
	StartTime       string `json:"StartTime"`
	ProcessID       int    `json:"ProcessID"`
	BoxID           int    `json:"BoxID"`
	Value           int64  `json:"Value"`
}
