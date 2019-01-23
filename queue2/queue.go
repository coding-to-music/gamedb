package queue

import (
	"encoding/json"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/cenkalti/backoff"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/log"
	"github.com/streadway/amqp"
)

type QueueName string

const (
	// C#
	QueueApps     QueueName = "Steam_Apps"
	QueuePackages QueueName = "Steam_Packages"
	QueueProfiles QueueName = "Steam_Profiles"

	// Go
	QueueAppsData     QueueName = "Steam_Apps_Data"
	QueueBundlesData  QueueName = "Steam_Bundles_Data"
	QueueChangesData  QueueName = "Steam_Changes_Data"
	QueueDelaysData   QueueName = "Steam_Delays_Data"
	QueuePackagesData QueueName = "Steam_Packages_Data"
	QueueProfilesData QueueName = "Steam_Profiles_Data"
)

var (
	consumeLock = new(sync.Mutex)
	produceLock = new(sync.Mutex)

	errInvalidQueue = errors.New("invalid queue")
	errEmptyMessage = errors.New("empty message")

	consumerConnection *amqp.Connection
	producerConnection *amqp.Connection

	consumerCloseChannel = make(chan *amqp.Error)
	producerCloseChannel = make(chan *amqp.Error)

	consumers = map[QueueName]queueInterface{
		QueueAppsData:   AppQueue{},
		QueueDelaysData: DelayQueue{},
	}
)

type queueInterface interface {
	// getQueueName() QueueName
	setQueueName(QueueName)
	process(msg amqp.Delivery) (requeue bool)
	consume()
}

func RunConsumers() {
	for k, v := range consumers {
		v.setQueueName(k)
		go v.consume()
	}
}

func Produce(queue QueueName, data []byte) (err error) {

	for _, v := range consumers {
		if queue == v.Message.getProduceQueue() {
			return v.produce(data)
		}
	}

	return errInvalidQueue
}

type BaseMessage struct {
	FirstSeen   time.Time
	Attempt     int
	NextAttempt time.Time
	Message     interface{}
}

type BaseQueue struct {
	Name QueueName
}

func (q BaseQueue) setQueueName(name QueueName) {
	q.Name = name
}

func (q BaseQueue) makeAConnection() (conn *amqp.Connection, err error) {

	operation := func() (err error) {

		log.Info("Connecting to Rabbit")

		conn, err = amqp.Dial(config.Config.RabbitDSN())
		log.Err(err) // Logging here as no max elasped time
		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.MaxElapsedTime = 0

	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { logInfo(err) })

	return conn, err
}

func (q BaseQueue) getQueue(conn *amqp.Connection, queue QueueName) (ch *amqp.Channel, qu amqp.Queue, err error) {

	ch, err = conn.Channel()
	if err != nil {
		return
	}

	err = ch.Qos(10, 0, false)
	if err != nil {
		return
	}

	qu, err = ch.QueueDeclare(string(q.Name), true, false, false, false, nil)

	return ch, qu, err
}

func (q BaseQueue) produce(data []byte) (err error) {

	// log.Info("Producing to: " + q.Message.getProduceQueue().String())

	// Connect
	err = func() error {

		produceLock.Lock()
		defer produceLock.Unlock()

		if producerConnection == nil {

			producerConnection, err = q.makeAConnection()
			if err != nil {
				log.Critical("Connecting to Rabbit: " + err.Error())
				return err
			}
			producerConnection.NotifyClose(producerCloseChannel)
		}

		return nil
	}()

	if err != nil {
		return err
	}

	//
	ch, qu, err := q.getQueue(producerConnection, q.Message.getProduceQueue())
	if err != nil {
		return err
	}

	// Close channel
	if ch != nil {
		defer func(ch *amqp.Channel) {
			err := ch.Close()
			log.Err(err)
		}(ch)
	}

	return ch.Publish("", qu.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         data,
	})
}

func (q BaseQueue) consume() {

	var err error

	for {

		// Connect
		err = func() error {

			consumeLock.Lock()
			defer consumeLock.Unlock()

			if consumerConnection == nil {

				consumerConnection, err = q.makeAConnection()
				if err != nil {
					log.Critical("Connecting to Rabbit: " + err.Error())
					return err
				}
				consumerConnection.NotifyClose(consumerCloseChannel)
			}

			return nil
		}()

		if err != nil {
			log.Err(err)
			return
		}

		//
		ch, qu, err := q.getQueue(consumerConnection, q.Name)
		if err != nil {
			log.Err(err)
			return
		}

		msgs, err := ch.Consume(qu.Name, "", false, false, false, false, nil)
		if err != nil {
			log.Err(err)
			return
		}

		// In a anon function so can return at anytime
		func(msgs <-chan amqp.Delivery, q BaseQueue) {

			for {
				select {
				case err = <-consumerCloseChannel:
					log.Warning(err)
					return
				case msg := <-msgs:

					requeue, err := q.Message.process(msg)
					if err != nil {
						logError(err, q.Name)
					}

					// Might be getting rate limited
					if err == steam.ErrNullResponse {
						logInfo("Null response, sleeping for 10 seconds")
						time.Sleep(time.Second * 10)
					}

					// No point in retrying if Steam has issues
					if err == steam.ErrNullResponse {
						logInfo("HTML response, sleeping for 10 seconds")
						time.Sleep(time.Second * 10)
					}

					if requeue {
						logInfo("Requeuing")
						err = s.requeueMessage(msg)
						logError(err)
					}

					err = msg.Ack(false)
					logError(err)
				}
			}

		}(msgs, q)

		// We only get here if the amqp connection gets closed

		err = ch.Close()
		log.Err(err)
	}
}

func (q BaseQueue) requeueMessage(msg amqp.Delivery) error {

	delayeMessage := rabbitConsumer{
		Attempt:   q.Attempt,
		StartTime: q.StartTime,
		EndTime:   q.EndTime,
		Message: RabbitMessageDelay{
			OriginalMessage: string(msg.Body),
			OriginalQueue:   q.Message.getConsumeQueue(),
		},
	}

	delayeMessage.IncrementAttempts()

	b, err := json.Marshal(delayeMessage)
	if err != nil {
		return err
	}

	err = Produce(QueueDelaysData, b)
	log.Err(err)

	return nil
}

func (q *BaseQueue) IncrementAttempts() {

	// Increment attemp
	q.Attempt++

	// Update end time
	var min float64 = 1
	var max float64 = 600

	var seconds = math.Pow(1.3, float64(q.Attempt))
	var minmaxed = math.Min(min+seconds, max)
	var rounded = math.Round(minmaxed)

	q.EndTime = q.StartTime.Add(time.Second * time.Duration(rounded))
}

//
type SteamKitJob struct {
	SequentialCount int    `json:"SequentialCount"`
	StartTime       string `json:"StartTime"`
	ProcessID       int    `json:"ProcessID"`
	BoxID           int    `json:"BoxID"`
	Value           int64  `json:"Value"`
}

func logInfo(interfaces ...interface{}) {
	log.Info(append(interfaces, log.LogNameConsumers)...)
}

func logError(interfaces ...interface{}) {
	log.Err(append(interfaces, log.LogNameConsumers)...)
}

func logWarning(interfaces ...interface{}) {
	log.Warning(append(interfaces, log.LogNameConsumers)...)
}
