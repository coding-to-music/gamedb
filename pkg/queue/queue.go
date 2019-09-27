package queue

import (
	"encoding/json"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/go-durationfmt"
	"github.com/Philipp15b/go-steam"
	"github.com/cenkalti/backoff"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type queueName string

//noinspection GoUnusedConst
const (
	queueApps      queueName = "GameDB_Go_Apps"
	queueAppPlayer queueName = "GameDB_Go_App_Players"
	queueBundles   queueName = "GameDB_Go_Bundles"
	queueChanges   queueName = "GameDB_Go_Changes"
	queueDelays    queueName = "GameDB_Go_Delays"
	queueFailed    queueName = "GameDB_Go_Failed"
	queueGroups    queueName = "GameDB_Go_Groups"
	queueGroupsNew queueName = "GameDB_Go_Groups_New"
	queuePackages  queueName = "GameDB_Go_Packages"
	queuePlayers   queueName = "GameDB_Go_Profiles"
	QueueSteam     queueName = "GameDB_Go_Steam"

	//
	maxBytesToStore int = 1024 * 10
)

var (
	consumeLock sync.Mutex
	produceLock sync.Mutex

	consumerConnection *amqp.Connection
	producerConnection *amqp.Connection

	consumerConnectionChannel = make(chan *amqp.Error)
	producerConnectionChannel = make(chan *amqp.Error)

	QueueRegister = map[queueName]baseQueue{
		queueApps: {
			Name:  queueApps,
			queue: &appQueue{},
		},
		queueBundles: {
			Name:  queueBundles,
			queue: &bundleQueue{},
		},
		queueChanges: {
			Name:  queueChanges,
			queue: &changeQueue{},
		},
		queueDelays: {
			Name:  queueDelays,
			queue: &delayQueue{},
		},
		queueGroups: {
			Name:  queueGroups,
			queue: &groupQueueScrape{},
		},
		queueGroupsNew: {
			Name:    queueGroupsNew,
			queue:   &groupQueueAPI{},
			maxTime: time.Hour * 24 * 7,
		},
		queuePackages: {
			Name:  queuePackages,
			queue: &packageQueue{},
		},
		queuePlayers: {
			Name:  queuePlayers,
			queue: &playerQueue{},
		},
		queueAppPlayer: {
			Name:  queueAppPlayer,
			queue: &appPlayerQueue{},
		},
		QueueSteam: {
			Name:       QueueSteam,
			queue:      &steamQueue{},
			DoNotScale: true,
		},
	}
)

func init() {

	// Reconnect to Rabbit on producer disconnect
	// Consumer connection is handled elsewhere
	go func() {
		for {
			var err error
			select {
			case err = <-producerConnectionChannel:
				logWarning("Consumer connection closed", err)
				logInfo("Getting new producer connection")

				producerConnection, err = getConnection()
				if err != nil {
					logCritical("Connecting to Rabbit: " + err.Error())
					continue
				}
				producerConnection.NotifyClose(producerConnectionChannel)
			}
		}
	}()
}

type messageInterface interface {
	produce(queue queueName)
	ackRetry() // Don't call this directly
	ackFail()  // Don't call this directly
}

type baseMessage struct {
	FirstSeen     time.Time `json:"first_seen"`
	LastSeen      time.Time `json:"last_seen"`
	Attempt       int       `json:"attempt"`
	OriginalQueue queueName `json:"original_queue"`
	Force         bool      `json:"force"`
	actionTaken   bool      `json:"-"`
	sync.Mutex    `json:"-"`
}

func (payload *baseMessage) produce(queue queueName) {

	if payload.OriginalQueue == "" {
		payload.OriginalQueue = queue
	}

	if payload.FirstSeen.IsZero() {
		payload.FirstSeen = time.Now()
	}

	if queue != queueDelays && queue != queueFailed {
		payload.LastSeen = time.Now()
		payload.Attempt++
	}
}

func (payload baseMessage) getNextAttempt() time.Time {

	var min = time.Second * 2
	var max = time.Hour

	var seconds float64
	seconds = math.Pow(1.5, float64(payload.Attempt))
	seconds = math.Max(seconds, min.Seconds())
	seconds = math.Min(seconds, max.Seconds())

	return payload.LastSeen.Add(time.Second * time.Duration(int64(seconds)))
}

// Remove from queue
func (payload *baseMessage) ack(msg amqp.Delivery) {

	payload.Lock()
	defer payload.Unlock()

	if payload.actionTaken {
		return
	}
	payload.actionTaken = true

	err := msg.Ack(false)
	logError(err)
}

func (payload *baseMessage) ackMulti(msg amqp.Delivery) {

	payload.Lock()
	defer payload.Unlock()

	if payload.actionTaken {
		return
	}
	payload.actionTaken = true

	err := msg.Ack(true)
	logError(err)
}

// Send to failed queue
func (payload *baseMessage) ackFail() {

	payload.Lock()
	defer payload.Unlock()

	if payload.actionTaken {
		return
	}
	payload.actionTaken = true
}

func ackFail(msg amqp.Delivery, message messageInterface) {

	message.ackFail()

	err := produce(message, queueFailed)
	if err != nil {
		logError(err)
		return
	}

	err = msg.Ack(false)
	if err != nil {
		logError(err)
		return
	}
}

func (payload *baseMessage) ackRetry() {

	payload.Lock()
	defer payload.Unlock()

	if payload.actionTaken {
		return
	}
	payload.actionTaken = true

	totalStr, err := durationfmt.Format(payload.getNextAttempt().Sub(payload.FirstSeen), "%mm %ss")
	if err != nil {
		logError(err)
	}

	leftStr, err := durationfmt.Format(payload.getNextAttempt().Sub(time.Now()), "%mm %ss")
	if err != nil {
		logError(err)
	}

	logInfo("Adding to delay queue for " + leftStr + ", " + totalStr + " total, attempt " + strconv.Itoa(payload.Attempt))
}

func ackRetry(msg amqp.Delivery, message messageInterface) {

	message.ackRetry()

	err := produce(message, queueDelays)
	if err != nil {
		logError(err)
		return
	}

	err = msg.Ack(false)
	if err != nil {
		logError(err)
		return
	}
}

type queueInterface interface {
	processMessages(msgs []amqp.Delivery)
}

type baseQueue struct {
	Name        queueName
	DoNotScale  bool
	SteamClient *steam.Client // Just used for Steam queue
	queue       queueInterface
	qos         int
	batchSize   int
	maxAttempts int
	maxTime     time.Duration
}

func (q baseQueue) getQOS() int {

	if q.qos != 0 {
		return q.qos
	}

	return 10
}

func (q baseQueue) getMaxTime() time.Duration {

	if q.maxTime != 0 {
		return q.maxTime
	}

	return time.Hour * 24 * 7
}

func (q baseQueue) ConsumeMessages() {

	var err error

	for {

		func() {

			// Connect
			err = func() error {

				consumeLock.Lock()
				defer consumeLock.Unlock()

				if consumerConnection == nil {

					log.Info("Getting new consumer connection")

					consumerConnection, err = getConnection()
					if err != nil {
						return err
					}
					consumerConnection.NotifyClose(consumerConnectionChannel)
				}

				return nil
			}()

			if err != nil {
				logCritical("Connecting to Rabbit: " + err.Error())
				return
			}

			//
			ch, qu, err := getQueue(consumerConnection, q.Name, q.getQOS())
			if err != nil {
				logError(err)
				return
			}

			defer func(ch *amqp.Channel) {
				err = ch.Close()
				logError(err)
			}(ch)

			tag := config.Config.Environment.Get() + "-" + config.GetSteamKeyTag()

			msgs, err := ch.Consume(qu.Name, tag, false, false, false, false, nil)
			if err != nil {
				logError(err)
				return
			}

			// In a anon function so can return at anytime
			func(msgs <-chan amqp.Delivery, q baseQueue) {

				var msgSlice []amqp.Delivery

				for {
					select {
					case err = <-consumerConnectionChannel:
						logWarning("Consumer connection closed", err)
						consumerConnection = nil
						return
					case msg := <-msgs:
						msgSlice = append(msgSlice, msg)
					}

					if len(msgSlice) >= q.batchSize {

						switch v := q.queue.(type) {
						case *steamQueue:
							v.SteamClient = q.SteamClient
							q.queue = v
						case *delayQueue:
							v.BaseQueue = q
							q.queue = v
						}

						q.queue.processMessages(msgSlice)
						msgSlice = []amqp.Delivery{}
					}
				}

			}(msgs, q)

			logWarning("Rabbit consumer connection has disconnected")

		}()
	}
}

func produce(message messageInterface, queue queueName) (err error) {

	// Connect
	err = func() error {

		produceLock.Lock()
		defer produceLock.Unlock()

		if producerConnection == nil {

			log.Info("Getting new producer connection")

			producerConnection, err = getConnection()
			if err != nil {
				logCritical("Connecting to Rabbit: " + err.Error())
				return err
			}
			producerConnection.NotifyClose(producerConnectionChannel)
		}

		return nil
	}()

	if err != nil {
		return err
	}

	//
	message.produce(queue) // Sets queue, firstSeen, lastSeen, attempt

	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	//
	ch, qu, err := getQueue(producerConnection, queue, QueueRegister[queue].getQOS())
	if err != nil {
		return err
	}

	defer func() {
		err = ch.Close()
		log.Err(err)
	}()

	return ch.Publish("", qu.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         b,
	})
}

func getConnection() (conn *amqp.Connection, err error) {

	operation := func() (err error) {

		amqpConfig := amqp.Config{}
		if config.IsLocal() {
			amqpConfig.Heartbeat = time.Hour
		}
		conn, err = amqp.DialConfig(config.RabbitDSN(), amqpConfig)

		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.MaxElapsedTime = 0
	policy.InitialInterval = 5 * time.Second

	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { logInfo(err) })

	return conn, err
}

func getQueue(conn *amqp.Connection, queue queueName, qos int) (ch *amqp.Channel, qu amqp.Queue, err error) {

	ch, err = conn.Channel()
	if err != nil {
		return
	}

	err = ch.Qos(qos, 0, false)
	if err != nil {
		return
	}

	qu, err = ch.QueueDeclare(string(queue), true, false, false, false, nil)

	return ch, qu, err
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

func logCritical(interfaces ...interface{}) {
	log.Critical(append(interfaces, log.LogNameConsumers)...)
}
