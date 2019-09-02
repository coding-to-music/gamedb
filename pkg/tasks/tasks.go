package tasks

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
)

func init() {
	for _, v := range tasks {
		TaskRegister[v.ID()] = v
	}
}

var tasks = []TaskInterface{
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

var TaskRegister = map[string]TaskInterface{}

type TaskInterface interface {
	ID() string
	Name() string
	Cron() string
	work()
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

//
func RunTask(task TaskInterface) {

	cronLogInfo("Cron started: " + task.Name())

	task.work()

	// Save config row
	err := sql.SetConfig(sql.ConfigType("task-"+task.ID()), strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	// Send websocket
	page := websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{Message: task.Name() + " complete"})

	//
	cronLogInfo("Cron complete: " + task.Name())
}

//
func GetTaskConfig(task TaskInterface) (config sql.Config, err error) {

	return sql.GetConfig(sql.ConfigType("task-" + task.ID()))
}
