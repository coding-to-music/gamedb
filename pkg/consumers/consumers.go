package consumers

import (
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/consumers/framework"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type queueName string

const (
	QueueSteam queueName = "GameDB_Go_Steam"

	queueApps      queueName = "GameDB_Go_Apps"
	queueAppPlayer queueName = "GameDB_Go_App_Players"
	queueBundles   queueName = "GameDB_Go_Bundles"
	queueChanges   queueName = "GameDB_Go_Changes"
	queueDelays    queueName = "GameDB_Go_Delays"
	queueFailed    queueName = "GameDB_Go_Failed"
	queueGroups    queueName = "GameDB_Go_Groups"
	queueGroups2   queueName = "GameDB_Go_Groups2"
	queueGroupsNew queueName = "GameDB_Go_Groups_New"
	queuePackages  queueName = "GameDB_Go_Packages"
	queuePlayers   queueName = "GameDB_Go_Profiles"
	queuePlayers2  queueName = "GameDB_Go_Profiles2"
	queueTest      queueName = "GameDB_Go_Test"
)

var (
	consumerConnection *framework.connection
	producerConnection *framework.connection
)

func init() {

	heartbeat := time.Minute
	if config.IsLocal() {
		heartbeat = time.Hour
	}

	var err error

	consumerConnection, err = framework.NewConnection(amqp.Config{Heartbeat: heartbeat})
	if err != nil {
		log.Info(err)
		return
	}

	framework.NewQueue(consumerConnection, "apps", 10, 1)

	producerConnection, err = framework.NewConnection(amqp.Config{Heartbeat: heartbeat})
	if err != nil {
		log.Info(err)
		return
	}
}
