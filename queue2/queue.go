package queue2

import (
	"encoding/json"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/log"
	"github.com/streadway/amqp"
)

type QueueName string

const (
	// C#
	QueueCSApps     QueueName = "GameDB_CS_Apps"
	QueueCSPackages QueueName = "GameDB_CS_Packages"
	QueueCSProfiles QueueName = "GameDB_CS_Profiles"

	// Go
	QueueGoApps     QueueName = "GameDB_Go_Apps"
	QueueGoBundles  QueueName = "GameDB_Go_Bundles"
	QueueGoChanges  QueueName = "GameDB_Go_Changes"
	QueueGoDelays   QueueName = "GameDB_Go_Delays"
	QueueGoPackages QueueName = "GameDB_Go_Packages"
	QueueGoProfiles QueueName = "GameDB_Go_Profiles"
)

var (
	consumeLock sync.Mutex
	produceLock sync.Mutex

	errInvalidQueue = errors.New("invalid queue")
	errEmptyMessage = errors.New("empty message")

	consumerConnection *amqp.Connection
	producerConnection *amqp.Connection

	consumerCloseChannel = make(chan *amqp.Error)
	producerCloseChannel = make(chan *amqp.Error)

	queues = map[QueueName]baseQueue{
		QueueGoApps:   baseQueue{queue: &AppQueue{}},
		QueueGoDelays: baseQueue{queue: &DelayQueue{}},
	}
)

type baseMessage struct {
	Message interface{}

	// Retry info
	FirstSeen     time.Time
	Attempt       int
	NextAttempt   time.Time
	OriginalQueue QueueName

	// Limits
	MaxAttempts int
	MaxTime     time.Duration
}

func (payload *baseMessage) init() {

	if payload.FirstSeen.IsZero() {
		payload.FirstSeen = time.Now()
	}

}

// Remove from queue
func (payload baseMessage) ack(msg amqp.Delivery) {

	err := msg.Ack(false)
	logError(err)
}

// Send to delay queue
func (payload baseMessage) delay(msg amqp.Delivery) {

	logInfo("Adding to delay queue")

	payload.Attempt++

	// var min float64 = 1
	// var max float64 = 600

	var seconds = math.Pow(1.3, float64(payload.Attempt))
	// var minmaxed = math.Min(min+seconds, max)
	// var rounded = math.Round(minmaxed)

	payload.NextAttempt = payload.FirstSeen.Add(time.Second * time.Duration(int64(seconds)))

	err := produce(QueueGoDelays, payload)
	logError(err)

	if err == nil {
		err = msg.Ack(false)
		logError(err)
	}
}

type queueInterface interface {
	setQueueName(QueueName)
	process(msg amqp.Delivery)
	consume()
}

type baseQueue struct {
	queue     queueInterface
	name      QueueName
	batchSize int // Not in use yet
}

func (q *baseQueue) setQueueName(name QueueName) {
	q.name = name
}

func (q baseQueue) consume() {

	var err error

	for {

		// Connect
		err = func() error {

			consumeLock.Lock()
			defer consumeLock.Unlock()

			if consumerConnection == nil {

				consumerConnection, err = makeAConnection()
				if err != nil {
					log.Critical("Connecting to Rabbit: " + err.Error())
					return err
				}
				consumerConnection.NotifyClose(consumerCloseChannel)
			}

			return nil
		}()

		if err != nil {
			logError(err)
			return
		}

		//
		ch, qu, err := getQueue(consumerConnection, q.name)
		if err != nil {
			logError(err)
			return
		}

		msgs, err := ch.Consume(qu.Name, "", false, false, false, false, nil)
		if err != nil {
			logError(err)
			return
		}

		// In a anon function so can return at anytime
		func(msgs <-chan amqp.Delivery, q baseQueue) {

			for {
				select {
				case err = <-consumerCloseChannel:
					log.Warning(err)
					return
				case msg := <-msgs:
					q.queue.process(msg)
				}
			}

		}(msgs, q)

		// We only get here if the amqp connection gets closed

		err = ch.Close()
		logError(err)
	}
}

func RunConsumers() {
	for k, v := range queues {
		v.setQueueName(k)
		go v.consume()
	}
}

func produce(queue QueueName, msg baseMessage) (err error) {

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// log.Info("Producing to: " + q.Message.getProduceQueue().String())s

	// Connect
	err = func() error {

		produceLock.Lock()
		defer produceLock.Unlock()

		if producerConnection == nil {

			producerConnection, err = makeAConnection()
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
	ch, qu, err := getQueue(producerConnection, queue)
	if err != nil {
		return err
	}

	// Close channel
	if ch != nil {
		defer func(ch *amqp.Channel) {
			err := ch.Close()
			logError(err)
		}(ch)
	}

	return ch.Publish("", qu.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         b,
	})
}

func makeAConnection() (conn *amqp.Connection, err error) {

	operation := func() (err error) {

		log.Info("Connecting to Rabbit")

		conn, err = amqp.Dial(config.Config.RabbitDSN())
		logError(err) // Logging here as no max elasped time
		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.MaxElapsedTime = 0

	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { logInfo(err) })

	return conn, err
}

func getQueue(conn *amqp.Connection, queue QueueName) (ch *amqp.Channel, qu amqp.Queue, err error) {

	ch, err = conn.Channel()
	if err != nil {
		return
	}

	err = ch.Qos(10, 0, false)
	if err != nil {
		return
	}

	qu, err = ch.QueueDeclare(string(queue), true, false, false, false, nil)

	return ch, qu, err
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
