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
	influxHelpers "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/websockets"
	influx "github.com/influxdata/influxdb1-client"
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
	QueuePlayersAwards       rabbit.QueueName = "GDB_Players.Awards"
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
		{Name: QueueAppPlayersTop},
		{Name: QueueAppPlayers},
		{Name: QueueAppsAchievementsSearch, prefetchSize: 1_000},
		{Name: QueueAppsAchievements},
		{Name: QueueAppsArticlesSearch, prefetchSize: 1_000},
		{Name: QueueAppsDLC},
		{Name: QueueAppsFindGroup},
		{Name: QueueAppsInflux},
		{Name: QueueAppsItems},
		{Name: QueueAppsMorelike},
		{Name: QueueAppsNews},
		{Name: QueueAppsReviews},
		{Name: QueueAppsSameowners},
		{Name: QueueAppsSearch, prefetchSize: 1_000},
		{Name: QueueAppsSteamspy},
		{Name: QueueAppsTwitch},
		{Name: QueueAppsWishlists, prefetchSize: 1_000},
		{Name: QueueAppsYoutube},
		{Name: QueueApps},
		{Name: QueueBundles},
		{Name: QueueChanges},
		{Name: QueueDelay, skipHeaders: true},
		{Name: QueueFailed, skipHeaders: true},
		{Name: QueueGroupsPrimaries, prefetchSize: 1_000},
		{Name: QueueGroupsSearch, prefetchSize: 1_000},
		{Name: QueueGroups},
		{Name: QueuePackagesPrices},
		{Name: QueuePackages},
		{Name: QueuePlayerRanks},
		{Name: QueuePlayersAwards},
		{Name: QueuePlayersAchievements},
		{Name: QueuePlayersAliases},
		{Name: QueuePlayersBadges},
		{Name: QueuePlayersGames},
		{Name: QueuePlayersGroups},
		{Name: QueuePlayersSearch, prefetchSize: 1_000},
		{Name: QueuePlayersWishlist},
		{Name: QueuePlayers},
		{Name: QueueStats},
		{Name: QueueSteam},
		{Name: QueueTest},
		{Name: QueueWebsockets},
	}

	ConsumersDefinitions = []QueueDefinition{
		{Name: QueueAppPlayers, consumer: appPlayersHandler},
		{Name: QueueAppPlayersTop, consumer: appPlayersHandler},
		{Name: QueueApps, consumer: appHandler},
		{Name: QueueAppsAchievements, consumer: appAchievementsHandler},
		{Name: QueueAppsAchievementsSearch, consumer: appsAchievementsSearchHandler, prefetchSize: 1_000},
		{Name: QueueAppsArticlesSearch, consumer: appsArticlesSearchHandler, prefetchSize: 1_000},
		{Name: QueueAppsDLC, consumer: appDLCHandler},
		{Name: QueueAppsFindGroup, consumer: appsFindGroupHandler},
		{Name: QueueAppsInflux, consumer: appInfluxHandler},
		{Name: QueueAppsItems, consumer: appItemsHandler},
		{Name: QueueAppsMorelike, consumer: appMorelikeHandler},
		{Name: QueueAppsNews, consumer: appNewsHandler},
		{Name: QueueAppsReviews, consumer: appReviewsHandler},
		{Name: QueueAppsSameowners, consumer: appSameownersHandler},
		{Name: QueueAppsSearch, consumer: appsSearchHandler, prefetchSize: 1_000},
		{Name: QueueAppsSteamspy, consumer: appSteamspyHandler},
		{Name: QueueAppsTwitch, consumer: appTwitchHandler},
		{Name: QueueAppsWishlists, consumer: appWishlistsHandler, prefetchSize: 1_000},
		{Name: QueueAppsYoutube, consumer: appYoutubeHandler},
		{Name: QueueBundles, consumer: bundleHandler},
		{Name: QueueChanges, consumer: changesHandler},
		{Name: QueueDelay, consumer: delayHandler, skipHeaders: true},
		{Name: QueueFailed, skipHeaders: true},
		{Name: QueueGroups, consumer: groupsHandler},
		{Name: QueueGroupsPrimaries, consumer: groupPrimariesHandler, prefetchSize: 1_000},
		{Name: QueueGroupsSearch, consumer: groupsSearchHandler, prefetchSize: 1_000},
		{Name: QueuePackages, consumer: packageHandler},
		{Name: QueuePackagesPrices, consumer: packagePriceHandler},
		{Name: QueuePlayerRanks, consumer: playerRanksHandler},
		{Name: QueuePlayers, consumer: playerHandler},
		{Name: QueuePlayersAchievements, consumer: playerAchievementsHandler},
		{Name: QueuePlayersAwards, consumer: playerAwardsHandler},
		{Name: QueuePlayersAliases, consumer: playerAliasesHandler},
		{Name: QueuePlayersBadges, consumer: playerBadgesHandler},
		{Name: QueuePlayersGames, consumer: playerGamesHandler},
		{Name: QueuePlayersGroups, consumer: playersGroupsHandler},
		{Name: QueuePlayersSearch, consumer: appsPlayersHandler, prefetchSize: 1_000},
		{Name: QueuePlayersWishlist, consumer: playersWishlistHandler},
		{Name: QueueStats, consumer: statsHandler},
		{Name: QueueSteam},
		{Name: QueueTest, consumer: testHandler},
		{Name: QueueWebsockets},
	}

	FrontendDefinitions = []QueueDefinition{
		{Name: QueueAppPlayersTop},
		{Name: QueueAppPlayers},
		{Name: QueueAppsAchievementsSearch, prefetchSize: 1_000},
		{Name: QueueAppsAchievements},
		{Name: QueueAppsArticlesSearch, prefetchSize: 1_000},
		{Name: QueueAppsInflux},
		{Name: QueueAppsNews},
		{Name: QueueAppsReviews},
		{Name: QueueAppsSearch, prefetchSize: 1_000},
		{Name: QueueAppsSteamspy},
		{Name: QueueAppsWishlists, prefetchSize: 1_000},
		{Name: QueueAppsYoutube},
		{Name: QueueApps},
		{Name: QueueBundles},
		{Name: QueueChanges},
		{Name: QueueDelay, skipHeaders: true},
		{Name: QueueFailed, skipHeaders: true},
		{Name: QueueGroupsPrimaries, prefetchSize: 1_000},
		{Name: QueueGroupsSearch, prefetchSize: 1_000},
		{Name: QueueGroups},
		{Name: QueuePackagesPrices},
		{Name: QueuePackages},
		{Name: QueuePlayerRanks},
		{Name: QueuePlayersGroups},
		{Name: QueuePlayersSearch, prefetchSize: 1_000},
		{Name: QueuePlayersWishlist},
		{Name: QueuePlayers},
		{Name: QueueStats},
		{Name: QueueSteam},
		{Name: QueueTest},
		{Name: QueueWebsockets, consumer: websocketHandler},
	}

	QueueSteamDefinitions = []QueueDefinition{
		{Name: QueueApps},
		{Name: QueueChanges},
		{Name: QueueDelay, skipHeaders: true},
		{Name: QueuePackages},
		{Name: QueuePlayers},
		{Name: QueueSteam, consumer: steamHandler},
	}

	QueueCronsDefinitions = []QueueDefinition{
		{Name: QueueAppPlayersTop},
		{Name: QueueAppPlayers},
		{Name: QueueAppsAchievementsSearch, prefetchSize: 1_000},
		{Name: QueueAppsAchievements},
		{Name: QueueAppsInflux},
		{Name: QueueAppsNews},
		{Name: QueueAppsReviews},
		{Name: QueueAppsSearch, prefetchSize: 1_000},
		{Name: QueueAppsSteamspy},
		{Name: QueueAppsWishlists, prefetchSize: 1_000},
		{Name: QueueAppsYoutube},
		{Name: QueueApps},
		{Name: QueueDelay, skipHeaders: true},
		{Name: QueueGroupsPrimaries, prefetchSize: 1_000},
		{Name: QueueGroupsSearch, prefetchSize: 1_000},
		{Name: QueueGroups},
		{Name: QueuePackages},
		{Name: QueuePlayerRanks},
		{Name: QueuePlayersGroups},
		{Name: QueuePlayersSearch, prefetchSize: 1_000},
		{Name: QueuePlayers},
		{Name: QueueStats},
		{Name: QueueSteam},
		{Name: QueueWebsockets},
	}

	ChatbotDefinitions = []QueueDefinition{
		{Name: QueuePlayers},
		{Name: QueueWebsockets},
	}
)

var discordClient *discordgo.Session

func SetDiscordClient(c *discordgo.Session) {
	discordClient = c
}

type QueueDefinition struct {
	Name         rabbit.QueueName
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
			Dial: amqp.DefaultDial(time.Second * 5),
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
			QueueName:     queue.Name,
			ConsumerName:  config.C.Environment + "-" + strconv.Itoa(k),
			PrefetchCount: prefetchSize,
			Handler:       queue.consumer,
			UpdateHeaders: !queue.skipHeaders,
			AutoDelete:    false,
		}

		q, err := rabbit.NewChannel(chanConfig)
		if err != nil {
			log.ErrS(string(queue.Name), err)
		} else {
			ProducerChannels[queue.Name] = q
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
						QueueName:     queue.Name,
						ConsumerName:  config.C.Environment + "-" + strconv.Itoa(k),
						PrefetchCount: prefetchSize,
						Handler:       queue.consumer,
						UpdateHeaders: !queue.skipHeaders,
						AutoDelete:    false,
					}

					q, err := rabbit.NewChannel(chanConfig)
					if err != nil {
						log.ErrS(string(queue.Name), err)
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

	item := memcache.ItemAppInQueue(payload.ID)

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

	item := memcache.ItemBundleInQueue(id)

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

func ProducePlayerAchievements(playerID int64, appID int, force bool, oldCount, oldCount100, oldCountApps int) (err error) {

	return produce(QueuePlayersAchievements, PlayerAchievementsMessage{
		PlayerID:     playerID,
		AppID:        appID,
		Force:        force,
		OldCount:     oldCount,
		OldCount100:  oldCount100,
		OldCountApps: oldCountApps,
	})
}

func ProduceGroup(payload GroupMessage) (err error) {

	if payload.UserAgent != nil && helpers.IsBot(*payload.UserAgent) {
		return ErrIsBot
	}

	item := memcache.ItemGroupInQueue(payload.ID)

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

	item := memcache.ItemPackageInQueue(payload.ID)

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

func ProducePlayer(payload PlayerMessage, event string) (err error) {

	if payload.UserAgent != nil && helpers.IsBot(*payload.UserAgent) {
		return ErrIsBot
	}

	payload.ID, err = helpers.IsValidPlayerID(payload.ID)
	if err != nil {
		return steamid.ErrInvalidPlayerID
	}

	item := memcache.ItemPlayerInQueue(payload.ID)

	_, err = memcache.Get(item.Key)
	if err == nil {
		return memcache.ErrInQueue
	}

	err = produce(QueuePlayers, payload)
	if err == nil {

		go func() {
			if err := memcache.Set(item.Key, item.Value, item.Expiration); err != nil {
				log.ErrS(err)
			}
		}()

		go func() {

			fields := map[string]interface{}{
				"produce": 1,
			}

			if payload.UserAgent != nil {
				fields["user-agent"] = *payload.UserAgent
			}

			point := influx.Point{
				Measurement: string(influxHelpers.InfluxMeasurementPlayerUpdates),
				Tags: map[string]string{
					"trigger": event,
				},
				Fields:    fields,
				Time:      time.Now(),
				Precision: "ms",
			}

			if _, err := influxHelpers.InfluxWrite(influxHelpers.InfluxRetentionPolicy14Day, point); err != nil {
				log.ErrS(err)
			}
		}()
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

func ProduceAppNews(appID int) (err error) {

	m := AppNewsMessage{AppID: appID}
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
		time.Sleep(time.Second / 1_000)
	}

	if val, ok := ProducerChannels[q]; ok {
		return val.Produce(payload, nil)
	}

	return errors.New("channel not in register")
}
