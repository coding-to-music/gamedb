package queue

import (
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/logging"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

type RabbitQueue string

func (rq RabbitQueue) String() string {
	return string(rq)
}

const (
	QueueApps         RabbitQueue = "Steam_Apps"          // Only takes IDs
	QueueAppsData     RabbitQueue = "Steam_Apps_Data"     //
	QueueChangesData  RabbitQueue = "Steam_Changes_Data"  //
	QueueDelaysData   RabbitQueue = "Steam_Delays_Data"   //
	QueuePackages     RabbitQueue = "Steam_Packages"      // Only takes IDs
	QueuePackagesData RabbitQueue = "Steam_Packages_Data" //
	QueueProfiles     RabbitQueue = "Steam_Profiles"      // Only takes IDs
	QueueProfilesData RabbitQueue = "Steam_Profiles_Data" //
)

var (
	consumers = map[RabbitQueue]rabbitConsumer{}

	errInvalidQueue = errors.New("invalid queue")
	errEmptyMessage = errors.New("empty message")

	rabbitDSN string

	consumerConnection   *amqp.Connection
	consumerCloseChannel chan *amqp.Error

	producerConnection   *amqp.Connection
	producerCloseChannel chan *amqp.Error
)

type queueInterface interface {
	getProduceQueue() RabbitQueue
	getConsumeQueue() RabbitQueue
	getRetryData() RabbitMessageDelay
	process(msg amqp.Delivery) (ack bool, requeue bool, err error)
}

func init() {

	consumerCloseChannel = make(chan *amqp.Error)
	producerCloseChannel = make(chan *amqp.Error)

	qs := []rabbitConsumer{
		{Message: RabbitMessageApp{}},
		{Message: RabbitMessageChanges{}},
		//{Message: RabbitMessageDelay{}},
		{Message: RabbitMessagePackage{}},
		{Message: RabbitMessageProfile{}},
	}

	for _, v := range qs {
		consumers[v.Message.getConsumeQueue()] = v
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

	for _, v := range consumers {
		go v.consume()
	}
}

func Produce(queue RabbitQueue, data []byte) (err error) {

	for _, v := range consumers {
		if queue == v.Message.getProduceQueue() {
			return v.produce(data)
		}
	}

	return errInvalidQueue
}

type rabbitConsumer struct {
	Message   queueInterface
	Attempt   int
	StartTime time.Time // Time first placed in delay queue
	EndTime   time.Time // Time to retry from delay queue
}

func (s rabbitConsumer) getQueue(conn *amqp.Connection, queue RabbitQueue) (ch *amqp.Channel, qu amqp.Queue, err error) {

	ch, err = conn.Channel()
	logging.Error(err)

	err = ch.Qos(10, 0, true)
	logging.Error(err)

	qu, err = ch.QueueDeclare(queue.String(), true, false, false, false, nil)
	logging.Error(err)

	return ch, qu, err
}

func (s rabbitConsumer) produce(data []byte) (err error) {

	logging.Info("Producing to: " + s.Message.getProduceQueue().String())

	// Connect
	if producerConnection == nil {

		producerConnection, err = amqp.Dial(rabbitDSN)
		if consumerConnection == nil {
			logging.Error(errors.New("rabbit not found"))
			return
		}
		producerConnection.NotifyClose(producerCloseChannel)
		if err != nil {
			return err
		}
	}

	//
	ch, qu, err := s.getQueue(producerConnection, s.Message.getProduceQueue())
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

func (s rabbitConsumer) consume() {

	logging.InfoL("Consuming from: " + s.Message.getConsumeQueue().String())

	var breakFor = false
	var err error

	for {

		// Connect
		if consumerConnection == nil {

			consumerConnection, err = amqp.Dial(rabbitDSN)
			if consumerConnection == nil {
				logging.Error(errors.New("rabbit not found"))
				return
			}
			consumerConnection.NotifyClose(consumerCloseChannel)
			if err != nil {
				logging.Error(err)
				return
			}
		}

		//
		ch, qu, err := s.getQueue(consumerConnection, s.Message.getConsumeQueue())
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

				// Ack/Nack/Requeue
				if ack {
					err = msg.Ack(false)
					logging.Error(err)
				} else {

					if requeue {
						logging.Info("Requeuing")
						err = s.requeueMessage(msg)
						logging.Error(err)
					}

					err = msg.Nack(false, false)
					logging.Error(err)
				}

				// Might be getting rate limited
				if err == steam.ErrNullResponse {
					logging.Info("Null response, sleeping for 10 seconds")
					time.Sleep(time.Second * 10)
				}
			}

			if breakFor {
				break
			}
		}

		//conn.Close()
		err = ch.Close()
		logging.Error(err)
	}
}

func (s rabbitConsumer) requeueMessage(msg amqp.Delivery) error {

	delayeMessage := rabbitConsumer{
		Attempt:   s.Attempt,
		StartTime: s.StartTime,
		EndTime:   s.EndTime,
		Message: RabbitMessageDelay{
			OriginalMessage: string(msg.Body),
			OriginalQueue:   s.Message.getConsumeQueue(),
		},
	}

	delayeMessage.IncrementAttempts()

	data, err := json.Marshal(delayeMessage)
	if err != nil {
		return err
	}

	err = Produce(QueueDelaysData, data)
	logging.Error(err)

	return nil
}

func (s *rabbitConsumer) IncrementAttempts() {

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
