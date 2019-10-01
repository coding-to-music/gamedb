package tasks

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/robfig/cron/v3"
)

func init() {
	for _, v := range tasks {
		TaskRegister[v.ID()] = BaseTask{v}
	}
}

var (
	Parser = cron.NewParser(cron.Minute | cron.Hour)
	tasks  = []TaskInterface{
		AppPlayers{},
		AppQueueAll{},
		AutoPlayerRefreshes{},
		ClearUpcomingCache{},
		DevCodeRun{},
		Developers{},
		Genres{},
		Instagram{},
		MemcacheClear{},
		SetBadgeCache{},
		PackagesQueueAll{},
		PlayerRanks{},
		PlayersQueueAll{},
		Publishers{},
		SteamClientPlayers{},
		Tags{},
		Wishlists{},
	}
)

var TaskRegister = map[string]TaskInterface{}

type TaskInterface interface {
	ID() string
	Name() string
	Cron() string
	work()

	// Base
	Next() (t time.Time)
	Prev() (t time.Time)
	GetTaskConfig() (config sql.Config, err error)
	Run()
}

type BaseTask struct {
	TaskInterface
}

func (task BaseTask) Next() (t time.Time) {

	sched, err := Parser.Parse(task.Cron())
	if err != nil {
		return t
	}
	return sched.Next(time.Now())
}

func (task BaseTask) Prev() (d time.Time) {

	sched, err := Parser.Parse(task.Cron())
	if err != nil {
		return d
	}
	next := sched.Next(time.Now())
	nextNext := sched.Next(next)
	diff := nextNext.Sub(next)

	return next.Add(-diff)
}

func (task BaseTask) Run() {

	cronLogInfo("Cron started: " + task.Name())

	// Send websocket
	page := websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{TaskID: task.ID(), Action: "started"})

	// Do work
	task.work()

	// Save config row
	err := sql.SetConfig(sql.ConfigType("task-"+task.ID()), strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	// Send websocket
	page = websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{
		TaskID: task.ID(),
		Action: "finished",
		Time:   BaseTask{task}.Next().Unix(),
	})

	//
	cronLogInfo("Cron complete: " + task.Name())
}

//
func (task BaseTask) GetTaskConfig() (config sql.Config, err error) {

	return sql.GetConfig(sql.ConfigType("task-" + task.ID()))
}

// Logging
func cronLogErr(interfaces ...interface{}) {
	log.Err(append(interfaces, log.LogNameCron, log.LogNameGameDB)...)
}

func cronLogInfo(interfaces ...interface{}) {
	log.Info(append(interfaces, log.LogNameCron, log.LogNameGameDB)...)
}

func statsLogger(tableName string, count int, total int, rowName string) {
	cronLogInfo("Updating " + tableName + " - " + strconv.Itoa(count) + " / " + strconv.Itoa(total) + ": " + rowName)
}
