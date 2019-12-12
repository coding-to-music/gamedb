package consumers

import (
	"errors"
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/consumers/framework"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/streadway/amqp"
)

const (
	QueueApps            framework.QueueName = "GDB_Apps"
	QueueAppsRegular     framework.QueueName = "GDB_Apps_Regular"
	QueueAppPlayers      framework.QueueName = "GDB_App_Players"
	QueueBundles         framework.QueueName = "GDB_Bundles"
	QueueChanges         framework.QueueName = "GDB_Changes"
	QueueGroups          framework.QueueName = "GDB_Groups"
	QueueGroupsNew       framework.QueueName = "GDB_Groups_New"
	QueuePackages        framework.QueueName = "GDB_Packages"
	QueuePackagesRegular framework.QueueName = "GDB_Packages_Regular"
	QueuePlayers         framework.QueueName = "GDB_Players"
	QueuePlayersRegular  framework.QueueName = "GDB_Players_Regular"
	QueuePlayerRanks     framework.QueueName = "GDB_Player_Ranks"
	QueueSteam           framework.QueueName = "GDB_Steam"

	QueueDelay  framework.QueueName = "GDB_Delay"
	QueueFailed framework.QueueName = "GDB_Failed"
)

var (
	Channels = map[string]map[framework.QueueName]*framework.Channel{
		framework.Consumer: {},
		framework.Producer: {},
	}

	QueueDefinitions = []queue{
		{name: QueueApps, consumer: appHandler},
		{name: QueueAppsRegular},
		{name: QueueAppPlayers, consumer: appPlayersHandler},
		{name: QueueBundles, consumer: bundleHandler},
		{name: QueueChanges, consumer: changesHandler},
		{name: QueueGroups, consumer: groupsHandler},
		{name: QueueGroupsNew, consumer: newGroupsHandler},
		{name: QueuePackages, consumer: packageHandler},
		{name: QueuePackagesRegular},
		{name: QueuePlayers, consumer: playerHandler},
		{name: QueuePlayersRegular},
		{name: QueuePlayerRanks, consumer: playerRanksHandler},
		{name: QueueDelay, consumer: delayHandler, skipHeaders: true},
		{name: QueueSteam, consumer: nil},
		{name: QueueFailed, consumer: nil},
	}

	QueueSteamDefinitions = []queue{
		{name: QueueSteam, consumer: steamHandler},
		{name: QueueApps, consumer: nil},
		{name: QueuePackages, consumer: nil},
		{name: QueuePlayers, consumer: nil},
		{name: QueueChanges, consumer: nil},
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

	producerConnection, err := framework.NewConnection("Producer", amqp.Config{Heartbeat: heartbeat})
	if err != nil {
		log.Info(err)
		return
	}

	for _, queue := range definitions {

		q, err := framework.NewChannel(producerConnection, queue.name, 10, queue.batchSize, queue.consumer, !queue.skipHeaders)
		if err != nil {
			log.Critical(string(queue.name), err)
		} else {
			Channels[framework.Producer][queue.name] = q
		}
	}

	// Consume
	if consume {

		consumerConnection, err := framework.NewConnection("Consumer", amqp.Config{Heartbeat: heartbeat})
		if err != nil {
			log.Info(err)
			return
		}

		for _, queue := range definitions {
			if queue.consumer != nil {

				q, err := framework.NewChannel(consumerConnection, queue.name, 10, queue.batchSize, queue.consumer, !queue.skipHeaders)
				if err != nil {
					log.Critical(string(queue.name), err)
					continue
				}

				Channels[framework.Consumer][queue.name] = q

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
	message.SendToQueue(Channels[framework.Producer][QueueFailed])
}

func sendToRetryQueue(message *framework.Message) {
	message.SendToQueue(Channels[framework.Producer][QueueDelay])
}

func sendToFirstQueue(message *framework.Message) {

	queue := message.FirstQueue()

	if queue == "" {
		queue = QueueFailed
	}

	message.SendToQueue(Channels[framework.Producer][queue])
}

// Producers
func ProduceApp(payload AppMessage) error {

	if !helpers.IsValidAppID(payload.ID) {
		return sql.ErrInvalidAppID
	}

	return Channels[framework.Producer][QueueApps].ProduceInterface(payload)
}

func ProduceAppRegular(payload AppMessage) error {
	return Channels[framework.Producer][QueueAppsRegular].ProduceInterface(payload)
}

func ProduceAppPlayers(payload AppPlayerMessage) error {

	if len(payload.IDs) == 0 {
		return nil
	}

	return Channels[framework.Producer][QueueAppPlayers].ProduceInterface(payload)
}

func ProduceBundle(payload BundleMessage) error {
	return Channels[framework.Producer][QueueBundles].ProduceInterface(payload)
}

func ProduceChanges(payload ChangesMessage) error {
	return Channels[framework.Producer][QueueChanges].ProduceInterface(payload)
}

func ProduceGroup(payload GroupMessage) error {
	return Channels[framework.Producer][QueueGroups].ProduceInterface(payload)
}

func ProduceGroupNew(payload GroupMessage) error {
	return Channels[framework.Producer][QueueGroupsNew].ProduceInterface(payload)
}

func ProducePackage(payload PackageMessage) error {

	if !sql.IsValidPackageID(payload.ID) {
		return sql.ErrInvalidPackageID
	}

	return Channels[framework.Producer][QueuePackages].ProduceInterface(payload)
}

func ProducePackageRegular(payload PackageMessage) error {
	return Channels[framework.Producer][QueuePackagesRegular].ProduceInterface(payload)
}

func ProducePlayer(payload PlayerMessage) error {

	if !helpers.IsValidPlayerID(payload.ID) {
		return errors.New("invalid player id: " + strconv.FormatInt(payload.ID, 10))
	}

	return Channels[framework.Producer][QueuePlayers].ProduceInterface(payload)
}

func ProducePlayerRegular(payload PlayerMessage) error {
	return Channels[framework.Producer][QueuePlayersRegular].ProduceInterface(payload)
}

func ProducePlayerRank(payload PlayerRanksMessage) error {
	return Channels[framework.Producer][QueuePlayerRanks].ProduceInterface(payload)
}

func ProduceSteam(payload SteamMessage) error {

	if len(payload.AppIDs) == 0 && len(payload.PackageIDs) == 0 {
		return nil
	}

	return Channels[framework.Producer][QueueSteam].ProduceInterface(payload)
}
