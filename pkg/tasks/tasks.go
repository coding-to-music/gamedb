package tasks

import (
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type TaskTime string

const ( //                                       min  hour
	CronTimeUpdateLastUpdatedPlayers TaskTime = "*    *"
	CronTimeSteamClientPlayers       TaskTime = "*/10 *"
	CronTimeAppPlayers               TaskTime = "*/10 *"
	CronTimeAppPlayersTop            TaskTime = "*/10 *"
	CronTimeAutoPlayerRefreshes      TaskTime = "0    */6"
	CronTimeAppsReviews              TaskTime = "10   0"
	CronTimeAppsYoutube              TaskTime = "20   0"
	CronTimeQueueAppGroups           TaskTime = "30   0"
	CronTimeQueuePlayerGroups        TaskTime = "40   0"
	CronTimePlayerRanks              TaskTime = "50   0"
	CronTimeScanProductQueues        TaskTime = "0    1"
	CronTimeSetBadgeCache            TaskTime = "10   1"
	CronTimeAppsInflux               TaskTime = "0    */6"
	CronTimeAppsWishlists            TaskTime = "30   1"
	CronTimeAddAppTagsToInflux       TaskTime = "40   1"
	CronTimeGenres                   TaskTime = "0    2"
	CronTimeTags                     TaskTime = "0    3"
	CronTimePublishers               TaskTime = "0    4"
	CronTimeDevelopers               TaskTime = "0    5"
	CronTimeCategories               TaskTime = "0    6"
	CronTimeInstagram                TaskTime = ""
)

type TaskGroup string

const (
	TaskGroupApps     TaskGroup = "apps"
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
		AppsAchievementsQueueAll{},
		AppsAchievementsQueueElastic{},
		AppsAddTagCountsToInflux{},
		AppsPlayerCheck{},
		AppsPlayerCheckTop{},
		AppsQueueAll{},
		AppsQueueElastic{},
		AppsQueueWishlists{},
		AppsQueueGroups{},
		AppsArticlesQueueElastic{},
		AppsQueueInflux{},
		AppsQueuePackages{},
		AppsQueueYoutube{},
		AppsQueueReviews{},
		AutoPlayerRefreshes{},
		BadgesUpdateRandom{},
		GroupsQueueElastic{},
		GroupsQueuePrimaries{},
		GroupsUpdateTop{},
		InstagramPost{},
		MemcacheClearAll{},
		PlayersQueueAll{},
		PlayersQueueElastic{},
		PlayersQueueLastUpdated{},
		PlayersQueueGroups{},
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

	zap.S().Info("Cron started: " + task.ID())

	// Send start websocket
	wsPayload := queue.AdminPayload{TaskID: task.ID(), Action: "started"}
	err := queue.ProduceWebsocket(wsPayload, websockets.PageAdmin)
	if err != nil {
		zap.S().Error(err)
	}

	// Do work
	policy := backoff.NewConstantBackOff(time.Second * 30)

	err = backoff.RetryNotify(task.work, backoff.WithMaxRetries(policy, 10), func(err error, t time.Duration) { zap.S().Info(err, task.ID()) })
	if err != nil {

		if val, ok := err.(TaskError); ok && val.Okay {
			zap.S().Info(task.ID(), err)
		} else {
			zap.S().Fatal(task.ID(), err)
		}
	} else {

		// Save config row
		err = mysql.SetConfig(mysql.ConfigID("task-"+task.ID()), strconv.FormatInt(time.Now().Unix(), 10))
		if err != nil {
			zap.S().Error(err)
		}

		// Send end websocket
		wsPayload = queue.AdminPayload{TaskID: task.ID(), Action: "finished", Time: Next(task).Unix()}
		err = queue.ProduceWebsocket(wsPayload, websockets.PageAdmin)
		if err != nil {
			zap.S().Error(err)
		}

		zap.S().Info("Cron finished: " + task.ID())
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
