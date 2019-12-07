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
	queueAppPlayers      framework.QueueName = "GDB_App_Players"
	queueBundles         framework.QueueName = "GDB_Bundles"
	queueChanges         framework.QueueName = "GDB_Changes"
	queueGroups          framework.QueueName = "GDB_Groups"
	queueGroupsNew       framework.QueueName = "GDB_Groups_New"
	queuePackages        framework.QueueName = "GDB_Packages"
	queuePackagesRegular framework.QueueName = "GDB_Packages_Regular"
	queuePlayers         framework.QueueName = "GDB_Players"
	queuePlayersRegular  framework.QueueName = "GDB_Players_Regular"
	queuePlayerRanks     framework.QueueName = "GDB_Player_Ranks"
	queueSteam           framework.QueueName = "GDB_Steam"

	queueDelay  framework.QueueName = "GDB_Delay"
	queueFailed framework.QueueName = "GDB_Failed"
)

var (
	channels = map[string]map[framework.QueueName]*framework.Channel{
		framework.Consumer: {},
		framework.Producer: {},
	}

	queueDefinitions = []queue{
		{name: queueApps, consumer: appHandler},
		{name: queueAppsRegular},
		{name: queueAppPlayers},
		{name: queueBundles, consumer: bundleHandler},
		{name: queueChanges},
		{name: queueGroups},
		{name: queueGroupsNew},
		{name: queuePackages},
		{name: queuePackagesRegular},
		{name: queuePlayers},
		{name: queuePlayersRegular},
		{name: queuePlayerRanks, consumer: playerRanksHandler},
		{name: queueSteam},
		{name: queueDelay, consumer: delayHandler, skipHeaders: true},
		{name: queueFailed},
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

// Message helpers
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

// Producers
func ProduceApp(payload BundleMessage) error {
	return channels[framework.Producer][queueApps].ProduceInterface(payload)
}

func ProduceAppRegular(payload AppMessage) error {
	return channels[framework.Producer][queueAppsRegular].ProduceInterface(payload)
}

func ProduceAppPlayers(payload AppPlayerMessage) error {
	return channels[framework.Producer][queueAppPlayers].ProduceInterface(payload)
}

func ProduceBundle(payload BundleMessage) error {
	return channels[framework.Producer][queueBundles].ProduceInterface(payload)
}

func ProduceChanges(payload ChangesMessage) error {
	return channels[framework.Producer][queueChanges].ProduceInterface(payload)
}

func ProduceGroup(payload interface{}) error {
	return channels[framework.Producer][queueGroups].ProduceInterface(payload)
}

func ProducePackage(payload PackageMessage) error {
	return channels[framework.Producer][queuePackages].ProduceInterface(payload)
}

func ProducePackageRegular(payload PackageMessage) error {
	return channels[framework.Producer][queuePackagesRegular].ProduceInterface(payload)
}

func ProducePlayer(payload PlayerMessage) error {
	return channels[framework.Producer][queuePlayers].ProduceInterface(payload)
}

func ProducePlayerRegular(payload PlayerMessage) error {
	return channels[framework.Producer][queuePlayersRegular].ProduceInterface(payload)
}

func ProducePlayerRank(payload PlayerRanksMessage) error {
	return channels[framework.Producer][queuePlayerRanks].ProduceInterface(payload)
}

func ProduceSteam(payload SteamMessage) error {
	return channels[framework.Producer][queueSteam].ProduceInterface(payload)
}
