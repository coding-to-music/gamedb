package pages

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/session-go/session"
	"github.com/gamedb/gamedb/cmd/webserver/helpers/middleware"
	sessionHelpers "github.com/gamedb/gamedb/cmd/webserver/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/go-chi/chi"
)

func AdminRouter() http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.MiddlewareAuthCheck())
	r.Use(middleware.MiddlewareAdminCheck(Error404Handler))

	r.Get("/", adminHandler)
	r.Post("/", adminHandler)
	r.Get("/{option}", adminHandler)
	r.Post("/{option}", adminHandler)
	return r
}

func adminHandler(w http.ResponseWriter, r *http.Request) {

	option := chi.URLParam(r, "option")

	err := r.ParseForm()
	if err != nil {
		log.Err(err, r)
	}

	switch option {
	case "run-cron":
		go adminRunCron(r)
	case "delete-bin-logs":
		go adminDeleteBinLogs(r)
	case "disable-consumers":
		go adminDisableConsumers()
	case "queues":
		go adminQueues(r)
	case "settings":
		go adminSettings(r)
	}

	// Redirect away after action
	if option != "" {

		err := session.SetFlash(r, sessionHelpers.SessionGood, option+" run")
		log.Err(err, r)

		http.Redirect(w, r, "/admin?"+option, http.StatusFound)
		return
	}

	// Get configs for times
	configs, err := sql.GetAllConfigs()
	log.Err(err, r)

	// Template
	t := adminTemplate{}
	t.fill(w, r, "Admin", "Game DB admin")
	t.hideAds = true
	t.Configs = configs
	t.Websockets = websockets.Pages

	for _, v := range tasks.TaskRegister {
		t.Tasks = append(t.Tasks, adminTaskTemplate{
			Task: v,
			Bad:  tasks.Bad(v),
			Next: tasks.Next(v),
			Prev: tasks.Prev(v),
		})
	}

	//
	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err, r)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Can't connect to mysql"})
		return
	}

	gorm.Raw("show binary logs").Scan(&t.BinLogs)

	var total uint64
	for k, v := range t.BinLogs {
		total = total + v.Bytes
		t.BinLogs[k].Total = total
	}

	gorm = gorm.Raw("SELECT * FROM information_schema.processlist where command != 'sleep'").Scan(&t.Queries)
	log.Err(gorm.Error, r)

	returnTemplate(w, r, "admin", t)
}

type adminTemplate struct {
	GlobalTemplate
	Errors     []string
	Configs    map[string]sql.Config
	Queries    []adminQuery
	BinLogs    []adminBinLog
	Websockets map[websockets.WebsocketPage]*websockets.Page
	Tasks      []adminTaskTemplate
}

type adminTaskTemplate struct {
	Task tasks.TaskInterface
	Bad  bool
	Next time.Time
	Prev time.Time
}

type adminQuery struct {
	ID       int    `gorm:"column:ID"`
	User     string `gorm:"column:USER"`
	Host     string `gorm:"column:HOST"`
	Database string `gorm:"column:DB"`
	Command  string `gorm:"column:COMMAND"`
	Seconds  int64  `gorm:"column:TIME"`
	State    string `gorm:"column:STATE"`
	Info     string `gorm:"column:INFO"`
}

type adminBinLog struct {
	Name      string `gorm:"column:Log_name"`
	Bytes     uint64 `gorm:"column:File_size"`
	Encrypted string `gorm:"column:Encrypted"`
	Total     uint64
}

func (at adminTemplate) GetMCConfigKey() string {
	return "wipe-memcache" + "-" + config.Config.Environment.Get()
}

func adminDisableConsumers() {

}

func adminRunCron(r *http.Request) {

	c := r.URL.Query().Get("cron")

	tasks.Run(tasks.TaskRegister[c])
}

func adminQueues(r *http.Request) {

	ua := r.UserAgent()

	//
	var appIDs []int
	if val := r.PostForm.Get("app-id"); val != "" {

		vals := strings.Split(val, ",")

		for _, val := range vals {

			val = strings.TrimSpace(val)

			appID, err := strconv.Atoi(val)
			if err == nil {
				appIDs = append(appIDs, appID)
			}
		}
	}

	if val := r.PostForm.Get("apps-ts"); val != "" {

		log.Info("Queueing apps")

		ts, err := strconv.ParseInt(val, 10, 64)
		if err == nil {

			apps, b, err := steam.GetSteam().GetAppList(100000, 0, ts, "")
			err = steam.AllowSteamCodes(err, b, nil)
			log.Err(err, r)
			if err == nil {

				log.Info("Found " + strconv.Itoa(len(apps.Apps)) + " apps")

				for _, app := range apps.Apps {
					appIDs = append(appIDs, app.AppID)
				}
			}
		}
	}

	var packageIDs []int
	if val := r.PostForm.Get("package-id"); val != "" {

		vals := strings.Split(val, ",")

		for _, val := range vals {

			val = strings.TrimSpace(val)

			packageID, err := strconv.Atoi(val)
			if err == nil {
				packageIDs = append(packageIDs, packageID)
			}
		}
	}

	if val := r.PostForm.Get("player-id"); val != "" {

		vals := strings.Split(val, ",")

		for _, val := range vals {

			val = strings.TrimSpace(val)

			playerID, err := strconv.ParseInt(val, 10, 64)
			if err == nil {
				err = queue.ProducePlayer(queue.PlayerMessage{ID: playerID, UserAgent: &ua})
				err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
				log.Err(err, r)
			}
		}
	}

	if val := r.PostForm.Get("bundle-id"); val != "" {

		vals := strings.Split(val, ",")

		for _, val := range vals {

			val = strings.TrimSpace(val)

			bundleID, err := strconv.Atoi(val)
			if err == nil {

				err = queue.ProduceBundle(bundleID)
				err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
				log.Err(err, r)
			}
		}
	}

	if val := r.PostForm.Get("test-id"); val != "" {

		val = strings.TrimSpace(val)
		count, err := strconv.Atoi(val)
		log.Err(err, r)

		for i := 1; i <= count; i++ {

			err = queue.ProduceTest(i)
			err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
			log.Err(err, r)
		}
	}

	if val := r.PostForm.Get("group-id"); val != "" {

		vals := strings.Split(val, ",")

		for _, val := range vals {

			err := queue.ProduceGroup(queue.GroupMessage{ID: val, UserAgent: &ua})
			err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
			log.Err(err, r)
		}
	}

	err := queue.ProduceSteam(queue.SteamMessage{AppIDs: appIDs, PackageIDs: packageIDs})
	err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
	log.Err(err, r)
}

func adminDeleteBinLogs(r *http.Request) {

	name := r.URL.Query().Get("name")
	if name != "" {

		gorm, err := sql.GetMySQLClient()
		if err != nil {
			log.Err(err, r)
			return
		}

		gorm.Exec("PURGE BINARY LOGS TO '" + name + "'")
	}
}

func adminSettings(r *http.Request) {

	middleware.DownMessage = r.PostFormValue("down-message")
}
