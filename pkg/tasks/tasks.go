package tasks

import (
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type TaskTime string

const ( //                                       min  hour
	CronTimeUpdateLastUpdatedPlayers TaskTime = "*    *"
	CronTimeNewsLatest               TaskTime = "*    *"
	CronTimeSteamClientPlayers       TaskTime = "*/10 *"
	CronTimeAppPlayers               TaskTime = "*/10 *"
	CronTimeAppPlayersTop            TaskTime = "*/10 *"
	CronTimeAppsSameowners           TaskTime = "*/10 *"
	CronTimeAutoPlayerRefreshes      TaskTime = "0    */6"
	CronTimeGameDBStats              TaskTime = "0    */6"
	CronTimeAppsReviews              TaskTime = "0    0"
	CronTimeAppsYoutube              TaskTime = "5    0"
	CronTimeQueueAppGroups           TaskTime = "10   0"
	CronTimeQueuePlayerGroups        TaskTime = "15   0"
	CronTimeScanProductQueues        TaskTime = "20   0"
	CronTimeSetBadgeCache            TaskTime = "25   0"
	CronTimePlayerRanks              TaskTime = "30   0"
	CronTimeStats                    TaskTime = "35   0"
	CronTimeAppsWishlists            TaskTime = "40   0"
	CronTimeAddAppTagsToInflux       TaskTime = "45   0"
	CronTimeAppsInflux               TaskTime = ""
	CronTimeSteamSpy                 TaskTime = ""
	CronTimeInstagram                TaskTime = ""
)

type TaskGroup string

const (
	TaskGroupApps     TaskGroup = "apps"
	TaskGroupBundles  TaskGroup = "bundles"
	TaskGroupGroups   TaskGroup = "groups"
	TaskGroupBadges   TaskGroup = "badges"
	TaskGroupNews     TaskGroup = "news"
	TaskGroupPlayers  TaskGroup = "players"
	TaskGroupPackages TaskGroup = "packages"
	TaskGroupElastic  TaskGroup = "elastic"
)

var (
	Parser       = cron.NewParser(cron.Minute | cron.Hour)
	TaskRegister = map[string]TaskInterface{}
	tasks        = []TaskInterface{
		&AppsAchievementsQueueAll{},
		&AppsAchievementsQueueElastic{},
		&AppsAddTagCountsToInflux{},
		&AppsArticlesQueueElastic{},
		&AppsPlayerCheckTop{},
		&AppsPlayerCheck{},
		&AppsQueueAll{},
		&AppsQueueElastic{},
		&AppsQueueGroups{},
		&AppsQueueInflux{},
		&AppsQueuePackages{},
		&AppsQueueReviews{},
		&AppsQueueSteamSpy{},
		&AppsQueueWishlists{},
		&AppsQueueYoutube{},
		&AppsSameOwners{},
		&ArticlesLatest{},
		&AutoPlayerRefreshes{},
		&BadgesUpdateRandom{},
		&BundlesQueueAll{},
		&BundlesQueueElastic{},
		&GlobalSteamStats{},
		&GroupsQueueElastic{},
		&GroupsQueuePrimaries{},
		&GroupsUpdateTop{},
		&InstagramPost{},
		&MemcacheClearAll{},
		&PlayersQueueAll{},
		&PlayersQueueElastic{},
		&PlayersQueueGroups{},
		&PlayersQueueLastUpdated{},
		&PlayersUpdateRanks{},
		&ProductsUpdateKeys{},
		&StatsTask{},
		&SteamOnline{},
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
	Group() TaskGroup
	Cron() TaskTime
	work() error
}

type BaseTask struct {
}

func Next(task TaskInterface) (t time.Time) {

	sched, err := Parser.Parse(string(task.Cron()))
	if err != nil {
		return t
	}
	return sched.Next(time.Now())
}

func Prev(task TaskInterface) (d time.Time) {

	sched, err := Parser.Parse(string(task.Cron()))
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

	// log.InfoS("Cron started: " + task.ID())

	// Send start websocket
	wsPayload := queue.AdminPayload{TaskID: task.ID(), Action: "started"}
	err := queue.ProduceWebsocket(wsPayload, websockets.PageAdmin)
	if err != nil {
		log.ErrS(err)
	}

	// Do work
	policy := backoff.NewConstantBackOff(time.Second * 30)

	notify := func(err error, t time.Duration) {
		log.Info("Cron retry failed", zap.String("cron id", task.ID()), zap.Error(err))
	}

	err = backoff.RetryNotify(task.work, backoff.WithMaxRetries(policy, 10), notify)
	if err != nil {

		if val, ok := err.(TaskError); ok && val.Okay {
			log.Info("Cron failed", zap.String("cron id", task.ID()), zap.Error(err))
		} else {
			log.Err("Cron failed", zap.String("cron id", task.ID()), zap.Error(err))
		}
	} else {

		// Save config row
		err = mysql.SetConfig(mysql.ConfigID("task-"+task.ID()), strconv.FormatInt(time.Now().Unix(), 10))
		if err != nil {
			log.ErrS(err)
		}

		// Send end websocket
		wsPayload = queue.AdminPayload{TaskID: task.ID(), Action: "finished", Time: Next(task).Unix()}
		err = queue.ProduceWebsocket(wsPayload, websockets.PageAdmin)
		if err != nil {
			log.ErrS(err)
		}

		// log.InfoS("Cron finished: " + task.ID())
	}
}

func GetTaskConfig(task TaskInterface) (config mysql.Config, err error) {
	return mysql.GetConfig(mysql.ConfigID("task-" + task.ID()))
}

//
type TaskError struct {
	Err  error
	Okay bool
}

func (te TaskError) Error() string {
	return te.Err.Error()
}
