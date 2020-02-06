package tasks

import (
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	pubsubHelpers "github.com/gamedb/gamedb/pkg/helpers/pubsub"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/robfig/cron/v3"
)

const ( //                        min hour
	CronTimeUpdateRandomPlayers = "* *"
	CronTimeSetBadgeCache       = "*/10 *"
	CronTimeSteamClientPlayers  = "*/10 *"
	CronTimeAppPlayers          = "*/10 *"
	CronTimeAutoPlayerRefreshes = "0 */6"
	CronTimeQueueAppGroups      = "0 0"
	CronTimeQueuePlayerGroups   = "0 0"
	CronTimeClearUpcomingCache  = "0 0"
	CronTimePlayerRanks         = "0 0"
	CronTimeGenres              = "0 3"
	CronTimeTags                = "0 4"
	CronTimePublishers          = "0 5"
	CronTimeDevelopers          = "0 6"
	CronTimeCategories          = "0 7"
	CronTimeInstagram           = "0 12"
)

var (
	Parser       = cron.NewParser(cron.Minute | cron.Hour)
	TaskRegister = map[string]TaskInterface{}
	tasks        = []TaskInterface{
		AppPlayers{},
		AppQueueAll{},
		AutoPlayerRefreshes{},
		DevCodeRun{},
		Developers{},
		Genres{},
		MemcacheClear{},
		PackagesQueueAll{},
		PlayerRanks{},
		PlayersQueueAll{},
		Publishers{},
		QueueAppGroups{},
		QueuePlayerGroups{},
		SetBadgeCache{},
		SteamClientPlayers{},
		Tags{},
		UpdateRandomPlayers{},
		// Instagram{},
	}
)

func init() {
	for _, v := range tasks {
		TaskRegister[v.ID()] = v
	}
}

type TaskInterface interface {
	ID() string
	Name() string
	Cron() string
	work() error
}

type BaseTask struct {
}

func Next(task TaskInterface) (t time.Time) {

	sched, err := Parser.Parse(task.Cron())
	if err != nil {
		return t
	}
	return sched.Next(time.Now())
}

func Prev(task TaskInterface) (d time.Time) {

	sched, err := Parser.Parse(task.Cron())
	if err != nil {
		return d
	}
	next := sched.Next(time.Now())
	nextNext := sched.Next(next)
	diff := nextNext.Sub(next)

	return next.Add(-diff)
}

func Bad(task TaskInterface) (b bool) {

	if task.Cron() == "" {
		return false
	}

	config, err := GetTaskConfig(task)
	if err == nil {
		i, err := strconv.ParseInt(config.Value, 10, 64)
		if err == nil {
			return Prev(task).Unix() > i
		}
	}

	return true
}

//
func Run(task TaskInterface) {

	log.Info("Cron started: " + task.Name())

	// Send start websocket
	_, err := pubsubHelpers.Publish(pubsubHelpers.PubSubTopicWebsockets, websockets.AdminPayload{TaskID: task.ID(), Action: "started"})
	log.Err(err)

	// Do work
	policy := backoff.NewConstantBackOff(time.Minute)

	err = backoff.RetryNotify(task.work, backoff.WithMaxRetries(policy, 10), func(err error, t time.Duration) { log.Info(err, task.ID(), err) })
	if err != nil {
		log.Critical(task.ID(), err)
	} else {

		// Save config row
		err = sql.SetConfig(sql.ConfigID("task-"+task.ID()), strconv.FormatInt(time.Now().Unix(), 10))
		log.Err(err)

		// Send end websocket
		_, err = pubsubHelpers.Publish(pubsubHelpers.PubSubTopicWebsockets, websockets.AdminPayload{TaskID: task.ID(), Action: "finished", Time: Next(task).Unix()})
		log.Err(err)
	}

	//
	log.Info("Cron complete: " + task.Name())
}

func GetTaskConfig(task TaskInterface) (config sql.Config, err error) {
	return sql.GetConfig(sql.ConfigID("task-" + task.ID()))
}

//
func statsLogger(tableName string, count int, total int, rowName string) {
	log.Info("Updating " + tableName + " - " + strconv.Itoa(count) + " / " + strconv.Itoa(total) + ": " + rowName)
}
