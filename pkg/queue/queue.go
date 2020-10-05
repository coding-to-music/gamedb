package queue

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamid"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type QueueMessageInterface interface {
	Queue() rabbit.QueueName
}

const (
	// Apps
	QueueApps                   rabbit.QueueName = "GDB_Apps"
	QueueAppsAchievements       rabbit.QueueName = "GDB_Apps.Achievements"
	QueueAppsItems              rabbit.QueueName = "GDB_Apps.Items"
	QueueAppsArticlesSearch     rabbit.QueueName = "GDB_Apps.Articles.Search"
	QueueAppsAchievementsSearch rabbit.QueueName = "GDB_Apps.Achievements.Search"
	QueueAppsYoutube            rabbit.QueueName = "GDB_Apps.Youtube"
	QueueAppsWishlists          rabbit.QueueName = "GDB_Apps.Wishlists"
	QueueAppsInflux             rabbit.QueueName = "GDB_Apps.Influx"
	QueueAppsDLC                rabbit.QueueName = "GDB_Apps.DLC"
	QueueAppsSameowners         rabbit.QueueName = "GDB_Apps.Sameowners"
	QueueAppsNews               rabbit.QueueName = "GDB_Apps.News"
	QueueAppsFindGroup          rabbit.QueueName = "GDB_Apps.FindGroup"
	QueueAppsReviews            rabbit.QueueName = "GDB_Apps.Reviews"
	QueueAppsTwitch             rabbit.QueueName = "GDB_Apps.Twitch"
	QueueAppsMorelike           rabbit.QueueName = "GDB_Apps.Morelike"
	QueueAppsSteamspy           rabbit.QueueName = "GDB_Apps.Steamspy"
	QueueAppsSearch             rabbit.QueueName = "GDB_Apps.Search"

	// Packages
	QueuePackages       rabbit.QueueName = "GDB_Packages"
	QueuePackagesPrices rabbit.QueueName = "GDB_Packages.Prices"

	// Players
	QueuePlayers             rabbit.QueueName = "GDB_Players"
	QueuePlayersAchievements rabbit.QueueName = "GDB_Players.Achievements"
	QueuePlayersBadges       rabbit.QueueName = "GDB_Players.Badges"
	QueuePlayersSearch       rabbit.QueueName = "GDB_Players.Search"
	QueuePlayersGames        rabbit.QueueName = "GDB_Players.Games"
	QueuePlayersAliases      rabbit.QueueName = "GDB_Players.Aliases"
	QueuePlayersGroups       rabbit.QueueName = "GDB_Players.Groups"
	QueuePlayersWishlist     rabbit.QueueName = "GDB_Players.Wishlist"

	// Group
	QueueGroups          rabbit.QueueName = "GDB_Groups"
	QueueGroupsSearch    rabbit.QueueName = "GDB_Groups.Search"
	QueueGroupsPrimaries rabbit.QueueName = "GDB_Groups.Primaries"

	// App players
	QueueAppPlayers    rabbit.QueueName = "GDB_App_Players"
	QueueAppPlayersTop rabbit.QueueName = "GDB_App_Players_Top"

	// Other
	QueueBundles     rabbit.QueueName = "GDB_Bundles"
	QueueChanges     rabbit.QueueName = "GDB_Changes"
	QueueDelay       rabbit.QueueName = "GDB_Delay"
	QueueFailed      rabbit.QueueName = "GDB_Failed"
	QueuePlayerRanks rabbit.QueueName = "GDB_Player_Ranks"
	QueueStats       rabbit.QueueName = "GDB_Stats"
	QueueSteam       rabbit.QueueName = "GDB_Steam"
	QueueTest        rabbit.QueueName = "GDB_Test"
	QueueWebsockets  rabbit.QueueName = "GDB_Websockets"
)

var (
	ProducerChannels = map[rabbit.QueueName]*rabbit.Channel{}

	AllProducerDefinitions = []QueueDefinition{
		{name: QueueAppPlayersTop},
		{name: QueueAppPlayers},
		{name: QueueAppsAchievementsSearch, prefetchSize: 1_000},
		{name: QueueAppsAchievements},
		{name: QueueAppsArticlesSearch, prefetchSize: 1_000},
		{name: QueueAppsDLC},
		{name: QueueAppsFindGroup},
		{name: QueueAppsInflux},
		{name: QueueAppsItems},
		{name: QueueAppsMorelike},
		{name: QueueAppsNews},
		{name: QueueAppsReviews},
		{name: QueueAppsSameowners},
		{name: QueueAppsSearch, prefetchSize: 1_000},
		{name: QueueAppsSteamspy},
		{name: QueueAppsTwitch},
		{name: QueueAppsWishlists, prefetchSize: 1_000},
		{name: QueueAppsYoutube},
		{name: QueueApps},
		{name: QueueBundles},
		{name: QueueChanges},
		{name: QueueDelay, skipHeaders: true},
		{name: QueueFailed, skipHeaders: true},
		{name: QueueGroupsPrimaries, prefetchSize: 1_000},
		{name: QueueGroupsSearch, prefetchSize: 1_000},
		{name: QueueGroups},
		{name: QueuePackagesPrices},
		{name: QueuePackages},
		{name: QueuePlayerRanks},
		{name: QueuePlayersAchievements},
		{name: QueuePlayersAliases},
		{name: QueuePlayersBadges},
		{name: QueuePlayersGames},
		{name: QueuePlayersGroups},
		{name: QueuePlayersSearch, prefetchSize: 1_000},
		{name: QueuePlayersWishlist},
		{name: QueuePlayers},
		{name: QueueStats},
		{name: QueueSteam},
		{name: QueueTest},
		{name: QueueWebsockets},
	}

	ConsumersDefinitions = []QueueDefinition{
		{name: QueueAppPlayers, consumer: appPlayersHandler},
		{name: QueueAppPlayersTop, consumer: appPlayersHandler},
		{name: QueueApps, consumer: appHandler},
		{name: QueueAppsAchievements, consumer: appAchievementsHandler},
		{name: QueueAppsAchievementsSearch, consumer: appsAchievementsSearchHandler, prefetchSize: 1_000},
		{name: QueueAppsArticlesSearch, consumer: appsArticlesSearchHandler, prefetchSize: 1_000},
		{name: QueueAppsDLC, consumer: appDLCHandler},
		{name: QueueAppsFindGroup, consumer: appsFindGroupHandler},
		{name: QueueAppsInflux, consumer: appInfluxHandler},
		{name: QueueAppsItems, consumer: appItemsHandler},
		{name: QueueAppsMorelike, consumer: appMorelikeHandler},
		{name: QueueAppsNews, consumer: appNewsHandler},
		{name: QueueAppsReviews, consumer: appReviewsHandler},
		{name: QueueAppsSameowners, consumer: appSameownersHandler},
		{name: QueueAppsSearch, consumer: appsSearchHandler, prefetchSize: 1_000},
		{name: QueueAppsSteamspy, consumer: appSteamspyHandler},
		{name: QueueAppsTwitch, consumer: appTwitchHandler},
		{name: QueueAppsWishlists, consumer: appWishlistsHandler, prefetchSize: 1_000},
		{name: QueueAppsYoutube, consumer: appYoutubeHandler},
		{name: QueueBundles, consumer: bundleHandler},
		{name: QueueChanges, consumer: changesHandler},
		{name: QueueDelay, consumer: delayHandler, skipHeaders: true},
		{name: QueueFailed, skipHeaders: true},
		{name: QueueGroups, consumer: groupsHandler},
		{name: QueueGroupsPrimaries, consumer: groupPrimariesHandler, prefetchSize: 1_000},
		{name: QueueGroupsSearch, consumer: groupsSearchHandler, prefetchSize: 1_000},
		{name: QueuePackages, consumer: packageHandler},
		{name: QueuePackagesPrices, consumer: packagePriceHandler},
		{name: QueuePlayerRanks, consumer: playerRanksHandler},
		{name: QueuePlayers, consumer: playerHandler},
		{name: QueuePlayersAchievements, consumer: playerAchievementsHandler},
		{name: QueuePlayersAliases, consumer: playerAliasesHandler},
		{name: QueuePlayersBadges, consumer: playerBadgesHandler},
		{name: QueuePlayersGames, consumer: playerGamesHandler},
		{name: QueuePlayersGroups, consumer: playersGroupsHandler},
		{name: QueuePlayersSearch, consumer: appsPlayersHandler, prefetchSize: 1_000},
		{name: QueuePlayersWishlist, consumer: playersWishlistHandler},
		{name: QueueStats, consumer: statsHandler},
		{name: QueueSteam},
		{name: QueueTest, consumer: testHandler},
		{name: QueueWebsockets},
	}

	FrontendDefinitions = []QueueDefinition{
		{name: QueueAppPlayersTop},
		{name: QueueAppPlayers},
		{name: QueueAppsAchievementsSearch, prefetchSize: 1_000},
		{name: QueueAppsAchievements},
		{name: QueueAppsArticlesSearch, prefetchSize: 1_000},
		{name: QueueAppsInflux},
		{name: QueueAppsReviews},
		{name: QueueAppsSearch, prefetchSize: 1_000},
		{name: QueueAppsSteamspy},
		{name: QueueAppsWishlists, prefetchSize: 1_000},
		{name: QueueAppsYoutube},
		{name: QueueApps},
		{name: QueueBundles},
		{name: QueueChanges},
		{name: QueueDelay, skipHeaders: true},
		{name: QueueFailed, skipHeaders: true},
		{name: QueueGroupsPrimaries, prefetchSize: 1_000},
		{name: QueueGroupsSearch, prefetchSize: 1_000},
		{name: QueueGroups},
		{name: QueuePackagesPrices},
		{name: QueuePackages},
		{name: QueuePlayerRanks},
		{name: QueuePlayersGroups},
		{name: QueuePlayersSearch, prefetchSize: 1_000},
		{name: QueuePlayersWishlist},
		{name: QueuePlayers},
		{name: QueueStats},
		{name: QueueSteam},
		{name: QueueTest},
		{name: QueueWebsockets, consumer: websocketHandler},
	}

	QueueSteamDefinitions = []QueueDefinition{
		{name: QueueApps},
		{name: QueueChanges},
		{name: QueueDelay, skipHeaders: true},
		{name: QueuePackages},
		{name: QueuePlayers},
		{name: QueueSteam, consumer: steamHandler},
	}

	QueueCronsDefinitions = []QueueDefinition{
		{name: QueueAppPlayersTop},
		{name: QueueAppPlayers},
		{name: QueueAppsAchievementsSearch, prefetchSize: 1_000},
		{name: QueueAppsAchievements},
		{name: QueueAppsInflux},
		{name: QueueAppsReviews},
		{name: QueueAppsSearch, prefetchSize: 1_000},
		{name: QueueAppsSteamspy},
		{name: QueueAppsWishlists, prefetchSize: 1_000},
		{name: QueueAppsYoutube},
		{name: QueueApps},
		{name: QueueDelay, skipHeaders: true},
		{name: QueueGroupsPrimaries, prefetchSize: 1_000},
		{name: QueueGroupsSearch, prefetchSize: 1_000},
		{name: QueueGroups},
		{name: QueuePackages},
		{name: QueuePlayerRanks},
		{name: QueuePlayersGroups},
		{name: QueuePlayersSearch, prefetchSize: 1_000},
		{name: QueuePlayers},
		{name: QueueStats},
		{name: QueueSteam},
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
	prefetchSize int
}

func Init(definitions []QueueDefinition) {

	heartbeat := time.Minute
	if config.IsLocal() {
		heartbeat = time.Hour
	}

	// Producers
	c := rabbit.ConnectionConfig{
		Address:  config.RabbitDSN(),
		ConnType: rabbit.Producer,
		Config: amqp.Config{
			Heartbeat: heartbeat,
			Properties: map[string]interface{}{
				"connection_name": config.C.Environment + "-" + string(rabbit.Consumer) + "-" + config.GetSteamKeyTag(),
			},
		},
		LogInfo: func(i ...interface{}) {
			// zap.S().Named(log.LogNameRabbit).Info(i...)
		},
		LogError: func(i ...interface{}) {
			zap.S().Named(log.LogNameRabbit).Error(i...)
		},
	}

	producerConnection, err := rabbit.NewConnection(c)
	if err != nil {
		log.InfoS(err)
		return
	}

	var consume bool

	for k, queue := range definitions {

		if queue.consumer != nil {
			consume = true
		}

		prefetchSize := 50
		if queue.prefetchSize > 0 {
			prefetchSize = queue.prefetchSize
		}

		chanConfig := rabbit.ChannelConfig{
			Connection:    producerConnection,
			QueueName:     queue.name,
			ConsumerName:  config.C.Environment + "-" + strconv.Itoa(k),
			PrefetchCount: prefetchSize,
			Handler:       queue.consumer,
			UpdateHeaders: !queue.skipHeaders,
			AutoDelete:    false,
		}

		q, err := rabbit.NewChannel(chanConfig)
		if err != nil {
			log.ErrS(string(queue.name), err)
		} else {
			ProducerChannels[queue.name] = q
		}
	}

	// Consumers
	if consume {

		c = rabbit.ConnectionConfig{
			Address:  config.RabbitDSN(),
			ConnType: rabbit.Consumer,
			Config: amqp.Config{
				Heartbeat: heartbeat,
				Properties: map[string]interface{}{
					"connection_name": config.C.Environment + "-" + string(rabbit.Consumer) + "-" + config.GetSteamKeyTag(),
				},
			},
			LogInfo: func(i ...interface{}) {
				// log.InfoS(i...)
			},
			LogError: func(i ...interface{}) {
				log.ErrS(i...)
			},
		}

		consumerConnection, err := rabbit.NewConnection(c)
		if err != nil {
			log.InfoS(err)
			return
		}

		for _, queue := range definitions {
			if queue.consumer != nil {

				prefetchSize := 50
				if queue.prefetchSize > 0 {
					prefetchSize = queue.prefetchSize
				}

				for k := range make([]int, 2) {

					chanConfig := rabbit.ChannelConfig{
						Connection:    consumerConnection,
						QueueName:     queue.name,
						ConsumerName:  config.C.Environment + "-" + strconv.Itoa(k),
						PrefetchCount: prefetchSize,
						Handler:       queue.consumer,
						UpdateHeaders: !queue.skipHeaders,
						AutoDelete:    false,
					}

					q, err := rabbit.NewChannel(chanConfig)
					if err != nil {
						log.ErrS(string(queue.name), err)
						continue
					}

					go q.Consume()
				}
			}
		}
	}
}

// Message helpers
func sendToFailQueue(message *rabbit.Message) {

	err := message.SendToQueueAndAck(ProducerChannels[QueueFailed], nil)
	if err != nil {
		log.ErrS(err)
	}
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

	err := message.SendToQueueAndAck(ProducerChannels[QueueDelay], po)
	if err != nil {
		log.ErrS(err)
	}
}

func sendToLastQueue(message *rabbit.Message) {

	queue := message.LastQueue()

	if queue == "" {
		queue = QueueFailed
	}

	err := message.SendToQueueAndAck(ProducerChannels[queue], nil)
	if err != nil {
		log.ErrS(err)
	}
}

// Producers
func ProduceApp(payload AppMessage) (err error) {

	if !helpers.IsValidAppID(payload.ID) {
		return mongo.ErrInvalidAppID
	}

	item := memcache.MemcacheAppInQueue(payload.ID)

	if payload.ChangeNumber == 0 {
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

func ProduceAppsInflux(appIDs []int) (err error) {
	m := AppInfluxMessage{AppIDs: appIDs}
	return produce(m.Queue(), m)
}

func ProduceAppsReviews(id int) (err error) {
	m := AppReviewsMessage{AppID: id}
	return produce(m.Queue(), m)
}

func ProduceAppsYoutube(id int, name string) (err error) {
	return produce(QueueAppsYoutube, AppYoutubeMessage{ID: id, Name: name})
}

func ProduceAppsWishlists(id int) (err error) {
	return produce(QueueAppsWishlists, AppWishlistsMessage{AppID: id})
}

func ProduceAppPlayers(appIDs []int) (err error) {

	if len(appIDs) == 0 {
		return nil
	}

	return produce(QueueAppPlayers, AppPlayerMessage{IDs: appIDs})
}

func ProduceAppPlayersTop(appIDs []int) (err error) {

	if len(appIDs) == 0 {
		return nil
	}

	return produce(QueueAppPlayersTop, AppPlayerMessage{IDs: appIDs})
}

func ProduceBundle(id int) (err error) {

	item := memcache.MemcacheBundleInQueue(id)

	_, err = memcache.Get(item.Key)
	if err == nil {
		return memcache.ErrInQueue
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

func ProducePlayerAchievements(playerID int64, appID int, force bool) (err error) {

	return produce(QueuePlayersAchievements, PlayerAchievementsMessage{PlayerID: playerID, AppID: appID, Force: force})
}

func ProduceGroup(payload GroupMessage) (err error) {

	if payload.UserAgent != nil && helpers.IsBot(*payload.UserAgent) {
		return ErrIsBot
	}

	item := memcache.MemcacheGroupInQueue(payload.ID)

	_, err = memcache.Get(item.Key)
	if err == nil {
		return memcache.ErrInQueue
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

	if payload.ChangeNumber == 0 {
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
		return steamid.ErrInvalidPlayerID
	}

	item := memcache.MemcachePlayerInQueue(payload.ID)

	_, err = memcache.Get(item.Key)
	if err == nil {
		return memcache.ErrInQueue
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

func ProduceGroupSearch(group *mongo.Group, groupID string, groupType string) (err error) {

	return produce(QueueGroupsSearch, GroupSearchMessage{Group: group, GroupID: groupID, GroupType: groupType})
}

func ProduceGroupPrimaries(groupID string, groupType string, prims int) (err error) {

	m := GroupPrimariesMessage{GroupID: groupID, GroupType: groupType, CurrentPrimaries: prims}
	return produce(m.Queue(), m)
}

func ProduceAchievementSearch(achievement mongo.AppAchievement, appName string, appOwners int64) (err error) {

	return produce(QueueAppsAchievementsSearch, AppsAchievementsSearchMessage{
		AppAchievement: achievement,
		AppName:        appName,
		AppOwners:      appOwners,
	})
}

func ProduceArticlesSearch(payload AppsArticlesSearchMessage) (err error) {

	return produce(QueueAppsArticlesSearch, payload)
}

//goland:noinspection GoUnusedExportedFunction
func ProducePlayerAlias(id int64, removed bool) (err error) {

	return produce(QueuePlayersAliases, PlayersAliasesMessage{PlayerID: id, PlayerRemoved: removed})
}

func ProduceAppAchievement(appID int, appName string, appOwners int64) (err error) {

	return produce(QueueAppsAchievements, AppAchievementsMessage{AppID: appID, AppName: appName, AppOwners: appOwners})
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

func ProduceStats(typex mongo.StatsType, ID int, appsCount int64) (err error) {

	m := StatsMessage{
		Type:      typex,
		StatID:    ID,
		AppsCount: appsCount,
	}

	return produce(m.Queue(), m)
}

func ProducePlayerGroup(player mongo.Player, skipGroupUpdate bool, force bool) (err error) {

	return produce(QueuePlayersGroups, PlayersGroupsMessage{
		Player:                    player,
		SkipGroupUpdate:           skipGroupUpdate,
		ForceResavingPlayerGroups: force,
	})
}

func ProduceAppSearch(app *mongo.App, appID int) (err error) {

	m := AppsSearchMessage{App: app, AppID: appID}
	return produce(m.Queue(), m)
}

func ProduceAppSteamSpy(appID int) (err error) {

	m := AppSteamspyMessage{AppID: appID}
	return produce(m.Queue(), m)
}

func ProducePlayerSearch(player *mongo.Player, playerID int64) (err error) {

	return produce(QueuePlayersSearch, PlayersSearchMessage{Player: player, PlayerID: playerID})
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
		time.Sleep(time.Second / 2_000)
	}

	if val, ok := ProducerChannels[q]; ok {
		return val.Produce(payload, nil)
	}

	return errors.New("channel not in register")
}
