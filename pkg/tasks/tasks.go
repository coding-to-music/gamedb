package tasks

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/robfig/cron/v3"
)

const ( //                         min hour
	CronTimeSetBadgeCache       = "*/6 *"
	CronTimeSteamClientPlayers  = "*/10 *"
	CronTimeWishlists           = "0 0"
	CronTimeClearUpcomingCache  = "0 0"
	CronTimePlayerRanks         = "2 0"
	CronTimeGenres              = "0 1"
	CronTimeTags                = "0 2"
	CronTimePublishers          = "0 3"
	CronTimeDevelopers          = "0 4"
	CronTimeCategories          = "0 5"
	CronTimeInstagram           = "0 12"
	CronTimeAppPlayers          = "0 */5"
	CronTimeAutoPlayerRefreshes = "0 */6"
)

var (
	Parser       = cron.NewParser(cron.Minute | cron.Hour)
	TaskRegister = map[string]TaskInterface{}
	Tasks        = []TaskInterface{
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

func init() {
	for _, v := range Tasks {
		TaskRegister[v.ID()] = v
	}
}

type TaskInterface interface {
	ID() string
	Name() string
	Cron() string
	work()
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

	// Send websocket
	page := websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{TaskID: task.ID(), Action: "started"})

	// Do work
	task.work()

	// Save config row
	err := sql.SetConfig(sql.ConfigID("task-"+task.ID()), strconv.FormatInt(time.Now().Unix(), 10))
	log.Err(err)

	// Send websocket
	page = websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{
		TaskID: task.ID(),
		Action: "finished",
		Time:   Next(task).Unix(),
	})

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
