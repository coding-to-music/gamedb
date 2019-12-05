package consumers

import (
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/consumers/framework"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

const (
	queueApps            framework.QueueName = "GDB_Apps"
	queueAppsRegular     framework.QueueName = "GDB_Apps_Regular"
	queueAppPlayer       framework.QueueName = "GDB_App_Players"
	queueBundles         framework.QueueName = "GDB_Bundles"
	queueChanges         framework.QueueName = "GDB_Changes"
	queueGroups          framework.QueueName = "GDB_Groups"
	queueGroupsNew       framework.QueueName = "GDB_Groups_New"
	queuePackages        framework.QueueName = "GDB_Packages"
	queuePackagesRegular framework.QueueName = "GDB_Packages_Regular"
	queuePlayers         framework.QueueName = "GDB_Players"
	queuePlayerRanks     framework.QueueName = "GDB_Player_Ranks"
	queueSteam           framework.QueueName = "GDB_Steam"

	queueDelay  framework.QueueName = "GDB_Delay"
	queueFailed framework.QueueName = "GDB_Failed"
	queueTest   framework.QueueName = "GDB_Test"
)

var (
	channels = map[string]map[framework.QueueName]*framework.Channel{
		framework.Consumer: {},
		framework.Producer: {},
	}

	queueDefinitions = []queue{
		{name: queueApps, consumer: appHandler},
		{name: queueAppsRegular},
		{name: queueAppPlayer},
		{name: queueBundles, consumer: bundleHandler},
		{name: queueChanges},
		{name: queueGroups},
		{name: queueGroupsNew},
		{name: queuePackages},
		{name: queuePackagesRegular},
		{name: queuePlayers},
		{name: queuePlayerRanks, consumer: playerRanksHandler},
		{name: queueSteam},
		{name: queueDelay, consumer: delayHandler, skipHeaders: true},
		{name: queueFailed},
		{name: queueTest},
	}
)

type queue struct {
	name          framework.QueueName
	consumer      framework.Handler
	skipHeaders   bool
	prefetchCount int
	batchSize     int
}

func Init(consume bool) {

	heartbeat := time.Minute
	if config.IsLocal() {
		heartbeat = time.Hour
	}

	connection, err := framework.NewConnection("producer", amqp.Config{Heartbeat: heartbeat})
	if err != nil {
		log.Info(err)
		return
	}

	for _, queue := range queueDefinitions {

		q, err := framework.NewChannel(connection, queue.name, queue.prefetchCount, queue.batchSize, queue.consumer, !queue.skipHeaders)
		if err != nil {
			log.Critical(string(queue.name), err)
		} else {
			channels[framework.Producer][queue.name] = q
		}
	}

	// Consume
	if consume {

		connection, err := framework.NewConnection("consumer", amqp.Config{Heartbeat: heartbeat})
		if err != nil {
			log.Info(err)
			return
		}

		for _, queue := range queueDefinitions {
			if queue.consumer != nil {

				q, err := framework.NewChannel(connection, queue.name, queue.prefetchCount, queue.batchSize, queue.consumer, !queue.skipHeaders)
				if err != nil {
					log.Critical(string(queue.name), err)
					continue
				}

				channels[framework.Consumer][queue.name] = q

				err = q.Consume()
				if err != nil {
					log.Critical(string(queue.name), err)
					continue
				}
			}
		}
	}
}

func sendToFailQueue(message *framework.Message) {
	message.SendToQueue(channels[framework.Producer][queueFailed])
}

func sendToBackOfQueue(message *framework.Message) {
	message.SendToQueue(message.Channel)
}

func sendToRetryQueue(message *framework.Message) {
	message.SendToQueue(channels[framework.Producer][queueDelay])
}

func sendToFirstQueue(message *framework.Message) {

	queue := message.FirstQueue()

	if queue == "" {
		queue = queueFailed
	}

	message.SendToQueue(channels[framework.Producer][queue])
}
