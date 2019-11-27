package consumers

import (
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/consumers/framework"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

const (
	queueApps      framework.QueueName = "GameDB_Go_Apps"
	queueAppPlayer framework.QueueName = "GameDB_Go_App_Players"
	queueBundles   framework.QueueName = "GameDB_Go_Bundles"
	queueChanges   framework.QueueName = "GameDB_Go_Changes"
	queueFailed    framework.QueueName = "GameDB_Go_Failed"
	queueGroups    framework.QueueName = "GameDB_Go_Groups"
	queueGroups2   framework.QueueName = "GameDB_Go_Groups2"
	queueGroupsNew framework.QueueName = "GameDB_Go_Groups_New"
	queuePackages  framework.QueueName = "GameDB_Go_Packages"
	queuePlayers   framework.QueueName = "GameDB_Go_Profiles"
	queuePlayers2  framework.QueueName = "GameDB_Go_Profiles2"
	queueSteam     framework.QueueName = "GameDB_Go_Steam"
	queueTest      framework.QueueName = "GameDB_Go_Test"
)

var (
	queues = map[string]map[framework.QueueName]*framework.Queue{
		framework.Consumer: {},
		framework.Producer: {},
	}
)

func Init() {

	heartbeat := time.Minute
	if config.IsLocal() {
		heartbeat = time.Hour
	}

	consumerConnection, err := framework.NewConnection(amqp.Config{Heartbeat: heartbeat})
	if err != nil {
		log.Info(err)
		return
	}

	producerConnection, err := framework.NewConnection(amqp.Config{Heartbeat: heartbeat})
	if err != nil {
		log.Info(err)
		return
	}

	queueHandlers := map[framework.QueueName]framework.Handler{
		queueApps:      appHandler,
		queueAppPlayer: nil,
		queueBundles:   bundleHandler,
		queueChanges:   nil,
		queueFailed:    nil,
		queueGroups:    nil,
		queueGroups2:   nil,
		queueGroupsNew: nil,
		queuePackages:  nil,
		queuePlayers:   nil,
		queueTest:      nil,
		queuePlayers2:  nil,
		// queueSteam:     nil,
	}

	for k, v := range queueHandlers {
		if v != nil {

			// Producer
			q, err := framework.NewQueue(producerConnection, k, 10, 1, v)
			if err == nil {
				queues[framework.Producer][k] = q
			}

			if err != nil {
				log.Err(string(k), err)
			}

			// Consumer
			q, err = framework.NewQueue(consumerConnection, k, 10, 1, v)
			if err == nil {
				queues[framework.Consumer][k] = q
			}

			if err != nil {
				log.Err(string(k), err)
			}
		}
	}

	// Start consuming
	for _, queue := range queues[framework.Consumer] {
		err = queue.Consume()
		log.Err(err)
	}
}
