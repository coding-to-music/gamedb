package queue

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/streadway/amqp"
)

const (
	// Apps
	QueueApps             rabbit.QueueName = "GDB_Apps"
	QueueAppsAchievements rabbit.QueueName = "GDB_Apps.Achievements"
	QueueAppsYoutube      rabbit.QueueName = "GDB_Apps.Youtube"
	QueueAppsInflux       rabbit.QueueName = "GDB_Apps.Influx"
	QueueAppsDLC          rabbit.QueueName = "GDB_Apps.DLC"
	QueueAppsSameowners   rabbit.QueueName = "GDB_Apps.Sameowners"
	QueueAppsNews         rabbit.QueueName = "GDB_Apps.News"
	QueueAppsReviews      rabbit.QueueName = "GDB_Apps.Reviews"
	QueueAppsTwitch       rabbit.QueueName = "GDB_Apps.Twitch"
	QueueAppsMorelike     rabbit.QueueName = "GDB_Apps.Morelike"
	QueueAppsSteamspy     rabbit.QueueName = "GDB_Apps.Steamspy"
	QueueAppsSearch       rabbit.QueueName = "GDB_Apps.Search"

	// Packages
	QueuePackages       rabbit.QueueName = "GDB_Packages"
	QueuePackagesPrices rabbit.QueueName = "GDB_Packages.Prices"

	// Players
	QueuePlayers             rabbit.QueueName = "GDB_Players"
	QueuePlayersAchievements rabbit.QueueName = "GDB_Players.Achievements"
	QueuePlayersSearch       rabbit.QueueName = "GDB_Players.Search"

	// Other
	QueueAppPlayers  rabbit.QueueName = "GDB_App_Players"
	QueueBundles     rabbit.QueueName = "GDB_Bundles"
	QueueChanges     rabbit.QueueName = "GDB_Changes"
	QueueDelay       rabbit.QueueName = "GDB_Delay"
	QueueFailed      rabbit.QueueName = "GDB_Failed"
	QueueGroups      rabbit.QueueName = "GDB_Groups"
	QueuePlayerRanks rabbit.QueueName = "GDB_Player_Ranks"
	QueueSteam       rabbit.QueueName = "GDB_Steam"
	QueueTest        rabbit.QueueName = "GDB_Test"
	QueueWebsockets  rabbit.QueueName = "GDB_Websockets"
)

func init() {
	rabbit.SetLogInfo(func(i ...interface{}) {
		i = append(i, log.LogNameRabbit)
		// log.Info(i...)
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

	AllProducerDefinitions = []QueueDefinition{
		{name: QueueAppPlayers},
		{name: QueueApps},
		{name: QueueAppsDLC},
		{name: QueueAppsYoutube},
		{name: QueueAppsInflux},
		{name: QueueAppsNews},
		{name: QueueAppsAchievements},
		{name: QueueAppsSameowners},
		{name: QueueAppsReviews},
		{name: QueueAppsMorelike},
		{name: QueueAppsTwitch},
		{name: QueueAppsSteamspy},
		{name: QueueBundles},
		{name: QueueChanges},
		{name: QueueGroups},
		{name: QueuePackages},
		{name: QueuePackagesPrices},
		{name: QueuePlayers},
		{name: QueuePlayerRanks},
		{name: QueuePlayersAchievements},
		{name: QueueDelay, skipHeaders: true},
		{name: QueueAppsSearch},
		{name: QueuePlayersSearch},
		{name: QueueSteam},
		{name: QueueFailed},
		{name: QueueTest},
		{name: QueueWebsockets},
	}

	ConsumersDefinitions = []QueueDefinition{
		{name: QueueAppPlayers, consumer: appPlayersHandler},
		{name: QueueApps, consumer: appHandler},
		{name: QueueAppsInflux, consumer: appInfluxHandler},
		{name: QueueAppsDLC, consumer: appDLCHandler},
		{name: QueueAppsYoutube, consumer: appYoutubeHandler},
		{name: QueueAppsNews, consumer: appNewsHandler},
		{name: QueueAppsAchievements, consumer: appAchievementsHandler},
		{name: QueueAppsSameowners, consumer: appSameownersHandler},
		{name: QueueAppsReviews, consumer: appReviewsHandler},
		{name: QueueAppsMorelike, consumer: appMorelikeHandler},
		{name: QueueAppsTwitch, consumer: appTwitchHandler},
		{name: QueueAppsSteamspy, consumer: appSteamspyHandler},
		{name: QueueBundles, consumer: bundleHandler},
		{name: QueueChanges, consumer: changesHandler},
		{name: QueueGroups, consumer: groupsHandler},
		{name: QueuePackages, consumer: packageHandler},
		{name: QueuePackagesPrices, consumer: packagePriceHandler},
		{name: QueuePlayers, consumer: playerHandler},
		{name: QueuePlayerRanks, consumer: playerRanksHandler},
		{name: QueuePlayersAchievements, consumer: playerAchievementsHandler},
		{name: QueueDelay, consumer: delayHandler, skipHeaders: true},
		{name: QueueAppsSearch, consumer: appsSearchHandler},
		{name: QueuePlayersSearch, consumer: appsPlayersHandler},
		{name: QueueSteam},
		{name: QueueFailed},
		{name: QueueTest, consumer: testHandler},
		{name: QueueWebsockets},
	}

	WebserverDefinitions = []QueueDefinition{
		{name: QueueApps},
		{name: QueueAppsYoutube},
		{name: QueueAppsInflux},
		{name: QueueAppPlayers},
		{name: QueueAppsSearch},
		{name: QueueBundles},
		{name: QueueChanges},
		{name: QueueGroups},
		{name: QueuePackages},
		{name: QueuePackagesPrices},
		{name: QueuePlayers},
		{name: QueuePlayerRanks},
		{name: QueueDelay, skipHeaders: true},
		{name: QueueSteam},
		{name: QueueFailed},
		{name: QueueTest},
		{name: QueueWebsockets, consumer: websocketHandler},
	}

	QueueSteamDefinitions = []QueueDefinition{
		{name: QueueSteam, consumer: steamHandler},
		{name: QueueApps},
		{name: QueuePackages},
		{name: QueuePlayers},
		{name: QueueChanges},
		{name: QueueDelay, skipHeaders: true},
	}

	QueueCronsDefinitions = []QueueDefinition{
		{name: QueueApps},
		{name: QueueAppsYoutube},
		{name: QueueAppsInflux},
		{name: QueueAppsSearch},
		{name: QueueAppPlayers},
		{name: QueueGroups},
		{name: QueuePackages},
		{name: QueuePlayers},
		{name: QueuePlayerRanks},
		{name: QueueSteam},
		{name: QueueDelay, skipHeaders: true},
		{name: QueueWebsockets},
	}

	ChatbotDefinitions = []QueueDefinition{
		{name: QueuePlayers},
		{name: QueueWebsockets},
	}
)

var discordClient *discordgo.Session

func SetDiscordClient(c *discordgo.Session) {
	discordClient = c
}

type QueueDefinition struct {
	name         rabbit.QueueName
	consumer     rabbit.Handler
	skipHeaders  bool
	batchSize    int
	prefetchSize int
}

func Init(definitions []QueueDefinition) {

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

	var consume bool

	for _, queue := range definitions {

		if queue.consumer != nil {
			consume = true
		}

		prefetchSize := 10
		if queue.prefetchSize > 0 {
			prefetchSize = queue.prefetchSize
		}

		q, err := rabbit.NewChannel(producerConnection, queue.name, config.Config.Environment.Get(), prefetchSize, queue.batchSize, queue.consumer, !queue.skipHeaders)
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

				q, err := rabbit.NewChannel(consumerConnection, queue.name, config.Config.Environment.Get(), prefetchSize, queue.batchSize, queue.consumer, !queue.skipHeaders)
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
func sendToFailQueue(message *rabbit.Message) {

	err := message.SendToQueueAndAck(Channels[rabbit.Producer][QueueFailed], nil)
	log.Err(err)
}

func sendToRetryQueue(message *rabbit.Message) {

	sendToRetryQueueWithDelay(message, 0)
}

func sendToRetryQueueWithDelay(message *rabbit.Message, delay time.Duration) {

	var po rabbit.ProduceOptions
	if delay > 0 {
		po = func(p amqp.Publishing) amqp.Publishing {
			p.Headers["delay-until"] = time.Now().Add(delay).Unix()
			return p
		}
	}

	err := message.SendToQueueAndAck(Channels[rabbit.Producer][QueueDelay], po)
	log.Err(err)
}

func sendToLastQueue(message *rabbit.Message) {

	queue := message.LastQueue()

	if queue == "" {
		queue = QueueFailed
	}

	err := message.SendToQueueAndAck(Channels[rabbit.Producer][queue], nil)
	log.Err(err)
}

// Producers
func ProduceApp(payload AppMessage) (err error) {

	if !helpers.IsValidAppID(payload.ID) {
		return mongo.ErrInvalidAppID
	}

	item := memcache.MemcacheAppInQueue(payload.ID)

	if payload.ChangeNumber == 0 && !config.IsLocal() {
		_, err = memcache.Get(item.Key)
		if err == nil {
			return memcache.ErrInQueue
		}
	}

	err = produce(QueueApps, payload)
	if err == nil {
		err = memcache.Set(item.Key, item.Value, item.Expiration)
	}

	return err
}

func ProduceAppsInflux(id int) (err error) {
	return produce(QueueAppsInflux, AppInfluxMessage{ID: id})
}

func ProduceAppsYoutube(id int, name string) (err error) {
	return produce(QueueAppsYoutube, AppYoutubeMessage{ID: id, Name: name})
}

func ProduceAppPlayers(payload AppPlayerMessage) (err error) {

	if len(payload.IDs) == 0 {
		return nil
	}

	return produce(QueueAppPlayers, payload)
}

func ProduceBundle(id int) (err error) {

	item := memcache.MemcacheBundleInQueue(id)

	if !config.IsLocal() {
		_, err = memcache.Get(item.Key)
		if err == nil {
			return memcache.ErrInQueue
		}
	}

	err = produce(QueueBundles, BundleMessage{ID: id})
	if err == nil {
		err = memcache.Set(item.Key, item.Value, item.Expiration)
	}

	return err
}

func ProduceChanges(payload ChangesMessage) (err error) {

	return produce(QueueChanges, payload)
}

func ProduceDLC(appID int, DLCIDs []int) (err error) {

	return produce(QueueAppsDLC, DLCMessage{AppID: appID, DLCIDs: DLCIDs})
}

func ProducePlayerAchievements(playerID int64, appID int) (err error) {

	return produce(QueuePlayersAchievements, PlayerAchievementsMessage{PlayerID: playerID, AppID: appID})
}

func ProduceGroup(payload GroupMessage) (err error) {

	if payload.ID == "" {
		return nil
	}

	if payload.UserAgent != nil && helpers.IsBot(*payload.UserAgent) {
		return ErrIsBot
	}

	payload.ID, err = helpers.IsValidGroupID(payload.ID)
	if err != nil {
		return err
	}

	item := memcache.MemcacheGroupInQueue(payload.ID)

	if !config.IsLocal() {
		_, err = memcache.Get(item.Key)
		if err == nil {
			return memcache.ErrInQueue
		}
	}

	err = produce(QueueGroups, payload)
	if err == nil {
		err = memcache.Set(item.Key, item.Value, item.Expiration)
	}

	return err
}

func ProducePackage(payload PackageMessage) (err error) {

	if !helpers.IsValidPackageID(payload.ID) {
		return mongo.ErrInvalidPackageID
	}

	item := memcache.MemcachePackageInQueue(payload.ID)

	if payload.ChangeNumber == 0 && !config.IsLocal() {
		_, err = memcache.Get(item.Key)
		if err == nil {
			return memcache.ErrInQueue
		}
	}

	err = produce(QueuePackages, payload)
	if err == nil {
		err = memcache.Set(item.Key, item.Value, item.Expiration)
	}

	return err
}

func producePackagePrice(payload PackagePriceMessage) (err error) {
	return produce(QueuePackagesPrices, payload)
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

	item := memcache.MemcachePlayerInQueue(payload.ID)

	if !config.IsLocal() {
		_, err = memcache.Get(item.Key)
		if err == nil {
			return memcache.ErrInQueue
		}
	}

	err = produce(QueuePlayers, payload)
	if err == nil {
		err = memcache.Set(item.Key, item.Value, item.Expiration)
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

func ProduceWebsocket(payload interface{}, pages ...websockets.WebsocketPage) (err error) {

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return produce(QueueWebsockets, WebsocketMessage{
		Pages:   pages,
		Message: b,
	})
}

func produce(q rabbit.QueueName, payload interface{}) error {

	if !config.IsLocal() {
		time.Sleep(time.Second / 1000)
	}

	if val, ok := Channels[rabbit.Producer][q]; ok {
		return val.Produce(payload, nil)
	}

	return errors.New("channel does not exist")
}
