package tasks

import (
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/robfig/cron/v3"
)

const ( //                         min hour
	CronTimeUpdateRandomPlayers = "*   *"
	CronTimeSteamClientPlayers  = "*/10 *"
	CronTimeAppPlayers          = "*/10 *"
	CronTimeAutoPlayerRefreshes = "0   */6"
	CronTimeAppsReviews         = "10  0"
	CronTimeAppsYoutube         = "20  0"
	CronTimeQueueAppGroups      = "30  0"
	CronTimeQueuePlayerGroups   = "40  0"
	CronTimePlayerRanks         = "50  0"
	CronTimeScanProductQueues   = "0   1"
	CronTimeSetBadgeCache       = "10  1"
	CronTimeAppsInflux          = "20  1,13"
	CronTimeGenres              = "0   2"
	CronTimeTags                = "0   3"
	CronTimePublishers          = "0   4"
	CronTimeDevelopers          = "0   5"
	CronTimeCategories          = "0   6"
	CronTimeInstagram           = ""
)

var (
	Parser       = cron.NewParser(cron.Minute | cron.Hour)
	TaskRegister = map[string]TaskInterface{}
	tasks        = []TaskInterface{
		AppsAchievementsQueueAll{},
		AppsAchievementsQueueElastic{},
		AppsPlayerCheck{},
		AppsQueueAll{},
		AppsQueueElastic{},
		AppsQueueGroups{},
		AppsQueueInflux{},
		AppsQueuePackages{},
		AppsQueueYoutube{},
		AppsQueueReviews{},
		AutoPlayerRefreshes{},
		BadgesUpdateRandom{},
		DevCodeRun{},
		GroupsQueueElastic{},
		GroupsUpdateTop{},
		InstagramPost{},
		MemcacheClearAll{},
		PlayersQueueAll{},
		PlayersQueueElastic{},
		PlayersQueueRandom{},
		PlayersUpdateRanks{},
		ProductsUpdateKeys{},
		StatsCategories{},
		StatsDevelopers{},
		StatsTags{},
		SteamOnline{},
		TasksGenres{},
		TasksPublishers{},
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

	log.Info("Cron started: " + task.ID())

	// Send start websocket
	wsPayload := queue.AdminPayload{TaskID: task.ID(), Action: "started"}
	err := queue.ProduceWebsocket(wsPayload, websockets.PageAdmin)
	log.Err(err)

	// Do work
	policy := backoff.NewConstantBackOff(time.Second * 30)

	err = backoff.RetryNotify(task.work, backoff.WithMaxRetries(policy, 10), func(err error, t time.Duration) { log.Info(err, task.ID()) })
	if err != nil {

		if val, ok := err.(TaskError); ok && val.Okay {
			log.Info(task.ID(), err)
		} else {
			log.Critical(task.ID(), err)
		}
	} else {

		// Save config row
		err = sql.SetConfig(sql.ConfigID("task-"+task.ID()), strconv.FormatInt(time.Now().Unix(), 10))
		if err != nil {
			log.Err(err)
		}

		// Send end websocket
		wsPayload = queue.AdminPayload{TaskID: task.ID(), Action: "finished", Time: Next(task).Unix()}
		err = queue.ProduceWebsocket(wsPayload, websockets.PageAdmin)
		if err != nil {
			log.Err(err)
		}

		log.Info("Cron finished: " + task.ID())
	}
}

func GetTaskConfig(task TaskInterface) (config sql.Config, err error) {
	return sql.GetConfig(sql.ConfigID("task-" + task.ID()))
}

//
type TaskError struct {
	Err  error
	Okay bool
}

func (te TaskError) Error() string {
	return te.Err.Error()
}
