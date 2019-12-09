package consumers

import (
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/consumers/framework"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
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

	QueueDefinitions = []queue{
		{name: queueApps, consumer: appHandler},
		{name: queueAppsRegular},
		{name: queueAppPlayers, consumer: appPlayersHandler},
		{name: queueBundles, consumer: bundleHandler},
		{name: queueChanges, consumer: changesHandler},
		{name: queueGroups, consumer: groupsHandler},
		{name: queueGroupsNew, consumer: newGroupsHandler},
		{name: queuePackages, consumer: packageHandler},
		{name: queuePackagesRegular},
		{name: queuePlayers, consumer: playerHandler},
		{name: queuePlayersRegular},
		{name: queuePlayerRanks, consumer: playerRanksHandler},
		{name: queueDelay, consumer: delayHandler, skipHeaders: true},
		{name: queueSteam, consumer: nil},
		{name: queueFailed, consumer: nil},
	}

	QueueSteamDefinitions = []queue{
		{name: queueSteam, consumer: steamHandler},
		{name: queueApps, consumer: nil},
		{name: queuePackages, consumer: nil},
		{name: queuePlayers, consumer: nil},
		{name: queueChanges, consumer: nil},
	}
)

type queue struct {
	name          framework.QueueName
	consumer      framework.Handler
	skipHeaders   bool
	prefetchCount int
	batchSize     int
}

func Init(definitions []queue, consume bool) {

	heartbeat := time.Minute
	if config.IsLocal() {
		heartbeat = time.Hour
	}

	connection, err := framework.NewConnection("Producer", amqp.Config{Heartbeat: heartbeat})
	if err != nil {
		log.Info(err)
		return
	}

	for _, queue := range definitions {

		q, err := framework.NewChannel(connection, queue.name, queue.prefetchCount, queue.batchSize, queue.consumer, !queue.skipHeaders)
		if err != nil {
			log.Critical(string(queue.name), err)
		} else {
			channels[framework.Producer][queue.name] = q
		}
	}

	// Consume
	if consume {

		connection, err := framework.NewConnection("Consumer", amqp.Config{Heartbeat: heartbeat})
		if err != nil {
			log.Info(err)
			return
		}

		for _, queue := range definitions {
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
func ProduceApp(payload AppMessage) error {
	return channels[framework.Producer][queueApps].ProduceInterface(payload)
}

func ProduceAppRegular(payload AppMessage) error {
	return channels[framework.Producer][queueAppsRegular].ProduceInterface(payload)
}

func ProduceAppPlayers(payload AppPlayerMessage) error {

	if len(payload.IDs) == 0 {
		return nil
	}

	return channels[framework.Producer][queueAppPlayers].ProduceInterface(payload)
}

func ProduceBundle(payload BundleMessage) error {
	return channels[framework.Producer][queueBundles].ProduceInterface(payload)
}

func ProduceChanges(payload ChangesMessage) error {
	return channels[framework.Producer][queueChanges].ProduceInterface(payload)
}

func ProduceGroup(payload GroupMessage) error {
	return channels[framework.Producer][queueGroups].ProduceInterface(payload)
}

func ProduceGroupNew(payload GroupMessage) error {
	return channels[framework.Producer][queueGroupsNew].ProduceInterface(payload)
}

func ProducePackage(payload PackageMessage) error {

	if !sql.IsValidPackageID(payload.ID) {
		return sql.ErrInvalidPackageID
	}

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

	if len(payload.AppIDs) == 0 && len(payload.PackageIDs) == 0 {
		return nil
	}

	return channels[framework.Producer][queueSteam].ProduceInterface(payload)
}
