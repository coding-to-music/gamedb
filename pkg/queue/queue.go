package queue

import (
	"errors"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/streadway/amqp"
)

const (
	QueueApps        rabbit.QueueName = "GDB_Apps"
	QueueAppsDaily   rabbit.QueueName = "GDB_Apps_Daily"
	QueueAppPlayers  rabbit.QueueName = "GDB_App_Players"
	QueueBundles     rabbit.QueueName = "GDB_Bundles"
	QueueChanges     rabbit.QueueName = "GDB_Changes"
	QueueGroups      rabbit.QueueName = "GDB_Groups"
	QueuePackages    rabbit.QueueName = "GDB_Packages"
	QueuePlayers     rabbit.QueueName = "GDB_Players"
	QueuePlayerRanks rabbit.QueueName = "GDB_Player_Ranks"
	QueueSteam       rabbit.QueueName = "GDB_Steam"

	QueueDelay  rabbit.QueueName = "GDB_Delay"
	QueueFailed rabbit.QueueName = "GDB_Failed"
	QueueTest   rabbit.QueueName = "GDB_Test"
)

func init() {
	rabbit.SetLogInfo(func(i ...interface{}) {
		i = append(i, log.LogNameRabbit)
		log.Info(i...)
	})
	rabbit.SetLogWarning(func(i ...interface{}) {
		i = append(i, log.LogNameRabbit)
		log.Warning(i...)
	})
	rabbit.SetLogError(func(i ...interface{}) {
		i = append(i, log.LogNameRabbit)
		log.Err(i...)
	})
}

var (
	Channels = map[rabbit.ConnType]map[rabbit.QueueName]*rabbit.Channel{
		rabbit.Consumer: {},
		rabbit.Producer: {},
	}

	AllDefinitions = []queueDef{
		{name: QueueApps, consumer: appHandler},
		{name: QueueAppsDaily, consumer: appDailyHandler, batchSize: 10, prefetchSize: 100},
		{name: QueueAppPlayers, consumer: appPlayersHandler},
		{name: QueueBundles, consumer: bundleHandler},
		{name: QueueChanges, consumer: changesHandler},
		{name: QueueGroups, consumer: groupsHandler},
		{name: QueuePackages, consumer: packageHandler},
		{name: QueuePlayers, consumer: playerHandler},
		{name: QueuePlayerRanks, consumer: playerRanksHandler},
		{name: QueueDelay, consumer: delayHandler, skipHeaders: true},
		{name: QueueSteam, consumer: nil},
		{name: QueueFailed, consumer: nil},
		{name: QueueTest, consumer: testHandler},
	}

	QueueSteamDefinitions = []queueDef{
		{name: QueueSteam, consumer: steamHandler},
		{name: QueueApps, consumer: nil},
		{name: QueuePackages, consumer: nil},
		{name: QueuePlayers, consumer: nil},
		{name: QueueChanges, consumer: nil},
		{name: QueueDelay, consumer: nil, skipHeaders: true},
	}

	QueueCronsDefinitions = []queueDef{
		{name: QueueApps, consumer: nil},
		{name: QueueAppsDaily, consumer: nil},
		{name: QueueAppPlayers, consumer: nil},
		{name: QueueGroups, consumer: nil},
		{name: QueuePackages, consumer: nil},
		{name: QueuePlayers, consumer: nil},
		{name: QueuePlayerRanks, consumer: nil},
		{name: QueueSteam, consumer: nil},
		{name: QueueDelay, consumer: nil, skipHeaders: true},
	}

	ChatbotDefinitions = []queueDef{
		{name: QueuePlayers, consumer: nil},
	}
)

var discordClient *discordgo.Session

func SetDiscordClient(c *discordgo.Session) {
	discordClient = c
}

type queueDef struct {
	name         rabbit.QueueName
	consumer     rabbit.Handler
	skipHeaders  bool
	batchSize    int
	prefetchSize int
}

func Init(definitions []queueDef, consume bool) {

	heartbeat := time.Minute
	if config.IsLocal() {
		heartbeat = time.Hour
	}

	// Producers
	producerConnection, err := rabbit.NewConnection(config.RabbitDSN(), rabbit.Producer, amqp.Config{Heartbeat: heartbeat, Properties: map[string]interface{}{
		"connection_name": config.Config.Environment.Get() + "-" + string(rabbit.Consumer) + "-" + config.GetSteamKeyTag(),
	}})
	if err != nil {
		log.Info(err)
		return
	}

	for _, queue := range definitions {

		prefetchSize := 10
		if queue.prefetchSize > 0 {
			prefetchSize = queue.prefetchSize
		}

		q, err := rabbit.NewChannel(producerConnection, queue.name, "consumer-name", prefetchSize, queue.batchSize, queue.consumer, !queue.skipHeaders)
		if err != nil {
			log.Critical(string(queue.name), err)
		} else {
			Channels[rabbit.Producer][queue.name] = q
		}
	}

	// Consumers
	if consume {

		consumerConnection, err := rabbit.NewConnection(config.RabbitDSN(), rabbit.Consumer, amqp.Config{Heartbeat: heartbeat, Properties: map[string]interface{}{
			"connection_name": config.Config.Environment.Get() + "-" + string(rabbit.Consumer) + "-" + config.GetSteamKeyTag(),
		}})
		if err != nil {
			log.Info(err)
			return
		}

		for _, queue := range definitions {
			if queue.consumer != nil {

				prefetchSize := 10
				if queue.prefetchSize > 0 {
					prefetchSize = queue.prefetchSize
				}

				q, err := rabbit.NewChannel(consumerConnection, queue.name, "consumer-name", prefetchSize, queue.batchSize, queue.consumer, !queue.skipHeaders)
				if err != nil {
					log.Critical(string(queue.name), err)
					continue
				}

				Channels[rabbit.Consumer][queue.name] = q

				go q.Consume()
			}
		}
	}
}

// Message helpers
func sendToFailQueue(messages ...*rabbit.Message) {

	for _, message := range messages {
		err := message.SendToQueue(Channels[rabbit.Producer][QueueFailed])
		log.Err(err)
	}
}

func sendToRetryQueue(messages ...*rabbit.Message) {

	for _, message := range messages {
		err := message.SendToQueue(Channels[rabbit.Producer][QueueDelay])
		log.Err(err)
	}
}

func sendToLastQueue(message *rabbit.Message) {

	queue := message.LastQueue()

	if queue == "" {
		queue = QueueFailed
	}

	err := message.SendToQueue(Channels[rabbit.Producer][queue])
	log.Err(err)
}

// Producers
func ProduceApp(payload AppMessage) (err error) {

	if !helpers.IsValidAppID(payload.ID) {
		return mongo.ErrInvalidAppID
	}

	mc := memcache.GetClient()
	item := memcache.MemcacheAppInQueue(payload.ID)

	if payload.ChangeNumber == 0 && !config.IsLocal() {
		_, err = mc.Get(item.Key)
		if err == nil {
			return memcache.ErrInQueue
		}
	}

	err = produce(QueueApps, payload)
	if err == nil {
		err = mc.Set(&item)
	}

	return err
}

func ProduceAppsDaily(id int, name string) (err error) {
	return produce(QueueAppsDaily, AppDailyMessage{ID: id, Name: name})
}

func ProduceAppPlayers(payload AppPlayerMessage) (err error) {

	if len(payload.IDs) == 0 {
		return nil
	}

	return produce(QueueAppPlayers, payload)
}

func ProduceBundle(id int) (err error) {

	mc := memcache.GetClient()
	item := memcache.MemcacheBundleInQueue(id)

	if !config.IsLocal() {
		_, err = mc.Get(item.Key)
		if err == nil {
			return memcache.ErrInQueue
		}
	}

	err = produce(QueueBundles, BundleMessage{ID: id})
	if err == nil {
		err = mc.Set(&item)
	}

	return err
}

func ProduceChanges(payload ChangesMessage) (err error) {

	return produce(QueueChanges, payload)
}

func ProduceGroup(payload GroupMessage) (err error) {

	if payload.UserAgent != nil && helpers.IsBot(*payload.UserAgent) {
		return ErrIsBot
	}

	payload.ID, err = helpers.IsValidGroupID(payload.ID)
	if err != nil {
		return err
	}

	mc := memcache.GetClient()
	item := memcache.MemcacheGroupInQueue(payload.ID)

	if !config.IsLocal() {
		_, err = mc.Get(item.Key)
		if err == nil {
			return memcache.ErrInQueue
		}
	}

	err = produce(QueueGroups, payload)
	if err == nil {
		err = mc.Set(&item)
	}

	return err
}

func ProducePackage(payload PackageMessage) (err error) {

	if !helpers.IsValidPackageID(payload.ID) {
		return mongo.ErrInvalidPackageID
	}

	mc := memcache.GetClient()
	item := memcache.MemcachePackageInQueue(payload.ID)

	if payload.ChangeNumber == 0 && !config.IsLocal() {
		_, err = mc.Get(item.Key)
		if err == nil {
			return memcache.ErrInQueue
		}
	}

	err = produce(QueuePackages, payload)
	if err == nil {
		err = mc.Set(&item)
	}

	return err
}

var ErrIsBot = errors.New("bots can't update players")

func ProducePlayer(payload PlayerMessage) (err error) {

	if payload.UserAgent != nil && helpers.IsBot(*payload.UserAgent) {
		return ErrIsBot
	}

	payload.ID, err = helpers.IsValidPlayerID(payload.ID)
	if err != nil {
		return helpers.ErrInvalidPlayerID
	}

	mc := memcache.GetClient()
	item := memcache.MemcachePlayerInQueue(payload.ID)

	if !config.IsLocal() {
		_, err = mc.Get(item.Key)
		if err == nil {
			return memcache.ErrInQueue
		}
	}

	err = produce(QueuePlayers, payload)
	if err == nil {
		err = mc.Set(&item)
	}

	return err
}

func ProducePlayerRank(payload PlayerRanksMessage) (err error) {

	return produce(QueuePlayerRanks, payload)
}

func ProduceSteam(payload SteamMessage) (err error) {

	if len(payload.AppIDs) == 0 && len(payload.PackageIDs) == 0 {
		return nil
	}

	return produce(QueueSteam, payload)
}

func ProduceTest(id int) (err error) {

	return produce(QueueTest, TestMessage{ID: id})
}

func produce(q rabbit.QueueName, payload interface{}) error {

	if val, ok := Channels[rabbit.Producer][q]; ok {
		return val.Produce(payload)
	}

	return errors.New("channel does not exist")
}
