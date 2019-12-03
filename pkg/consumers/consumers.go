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
	channels = map[string]map[framework.QueueName]framework.Channel{
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

		queueDelay: delayHandler,
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

		q, err := framework.NewChannel(connection, k, 10, 1, v, updateHeaders)
		if err != nil {
			log.Critical(string(k), err)
		} else {
			channels[framework.Producer][k] = q
		}
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

			q, err := framework.NewChannel(connection, k, 10, 1, v, updateHeaders)
			if err != nil {
				log.Critical(string(k), err)
				continue
			}

			channels[framework.Consumer][k] = q

			err = q.Consume()
			if err != nil {
				log.Critical(string(k), err)
				continue
			}
		}
	}
}

func sendToFailQueue(message framework.Message) {
	message.SendToQueue(channels[framework.Producer][queueFailed])
}

func sendToBackOfQueue(message framework.Message) {
	message.SendToQueue(message.Channel)
}

func sendToRetryQueue(message framework.Message) {
	message.SendToQueue(channels[framework.Producer][queueDelay])
}

func sendToFirstQueue(message framework.Message) {

	queue := message.FirstQueue()

	if queue == "" {
		queue = queueFailed
	}

	message.SendToQueue(channels[framework.Producer][queue])
}
