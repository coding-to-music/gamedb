package consumers

import (
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/consumers/framework"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

const (
	queueApps        framework.QueueName = "GDB_Apps"
	queueAppPlayer   framework.QueueName = "GDB_App_Players"
	queueBundles     framework.QueueName = "GDB_Bundles"
	queueChanges     framework.QueueName = "GDB_Changes"
	queueGroups      framework.QueueName = "GDB_Groups"
	queueGroupsNew   framework.QueueName = "GDB_Groups_New"
	queuePackages    framework.QueueName = "GDB_Packages"
	queuePlayers     framework.QueueName = "GDB_Profiles"
	queuePlayerRanks framework.QueueName = "GDB_Player_Ranks"
	queueSteam       framework.QueueName = "GDB_Steam"

	queueDelay  framework.QueueName = "GDB_Delay"
	queueFailed framework.QueueName = "GDB_Failed"
	queueTest   framework.QueueName = "GDB_Test"
)

var (
	queues = map[string]map[framework.QueueName]*framework.Queue{
		framework.Consumer: {},
		framework.Producer: {},
	}
	handlers = map[framework.QueueName]framework.Handler{
		queueApps:        appHandler,
		queueAppPlayer:   nil,
		queueBundles:     bundleHandler,
		queueChanges:     nil,
		queueGroups:      nil,
		queueGroupsNew:   nil,
		queuePackages:    nil,
		queuePlayers:     nil,
		queuePlayerRanks: playerRanksHandler,
		queueSteam:       nil,

		queueDelay: nil,
		queueTest:  nil,
	}
)

func InitProducers() {

	heartbeat := time.Minute
	if config.IsLocal() {
		heartbeat = time.Hour
	}

	connection, err := framework.NewConnection("producer", amqp.Config{Heartbeat: heartbeat})
	if err != nil {
		log.Info(err)
		return
	}

	for k, v := range handlers {

		updateHeaders := true
		if k == queueDelay {
			updateHeaders = false
		}

		q, err := framework.NewQueue(connection, k, 10, 1, v, updateHeaders)
		if err != nil {
			log.Critical(string(k), err)
		} else {
			queues[framework.Producer][k] = q
		}
	}

	// Start consuming
	for _, queue := range queues[framework.Consumer] {
		err = queue.Consume()
		log.Err(err)
	}
}

func InitConsumers() {

	heartbeat := time.Minute
	if config.IsLocal() {
		heartbeat = time.Hour
	}

	connection, err := framework.NewConnection("consumer", amqp.Config{Heartbeat: heartbeat})
	if err != nil {
		log.Info(err)
		return
	}

	for k, v := range handlers {
		if v != nil {

			updateHeaders := true
			if k == queueDelay {
				updateHeaders = false
			}

			q, err := framework.NewQueue(connection, k, 10, 1, v, updateHeaders)
			if err != nil {
				log.Critical(string(k), err)
			} else {
				queues[framework.Consumer][k] = q
			}
		}
	}

	// Start consuming
	for _, queue := range queues[framework.Consumer] {
		err = queue.Consume()
		log.Err(err)
	}
}

func sendToFailQueue(message framework.Message) {
	message.SendToQueue(queues[framework.Producer][queueFailed])
}

func sendToBackOfQueue(message framework.Message) {
	message.SendToQueue(message.Queue)
}

func sendToRetryQueue(message framework.Message) {
	message.SendToQueue(queues[framework.Producer][queueDelay])
}
