package pages

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/session-go/session"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/middleware"
	sessionHelpers "github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/go-chi/chi"
)

func AdminRouter() http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.MiddlewareAuthCheck())
	r.Use(middleware.MiddlewareAdminCheck(Error404Handler))

	r.Get("/", adminHandler)
	r.Get("/tasks", adminTasksHandler)
	r.Get("/queues", adminQueuesHandler)
	r.Post("/queues", adminQueuesHandler)
	r.Get("/settings", adminSettingsHandler)
	r.Post("/settings", adminSettingsHandler)
	r.Get("/sql-bin-logs", adminBinLogsHandler)
	r.Get("/websockets", adminWebsocketsHandler)
	return r
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin/tasks", http.StatusFound)
}

func adminTasksHandler(w http.ResponseWriter, r *http.Request) {

	task := r.URL.Query().Get("run")
	if task != "" {

		c := r.URL.Query().Get("run")

		if val, ok := tasks.TaskRegister[c]; ok {
			go tasks.Run(val)
		}

		err := session.SetFlash(r, sessionHelpers.SessionGood, "Done")
		log.Err(err, r)

		http.Redirect(w, r, "/admin/tasks", http.StatusFound)
		return
	}

	//
	t := adminTasksTemplate{}
	t.fill(w, r, "Admin", "Admin")
	t.hideAds = true

	var grouped = map[string][]adminTaskTemplate{}

	for _, v := range tasks.TaskRegister {
		grouped[v.Group()] = append(grouped[v.Group()], adminTaskTemplate{
			Task: v,
			Bad:  tasks.Bad(v),
			Next: tasks.Next(v),
			Prev: tasks.Prev(v),
		})
	}

	t.Tasks = []adminTaskListTemplate{
		{Tasks: grouped[tasks.TaskGroupApps], Title: "Apps"},
		{Tasks: grouped[tasks.TaskGroupPackages], Title: "Packages"},
		{Tasks: grouped[tasks.TaskGroupGroups], Title: "Groups"},
		{Tasks: grouped[tasks.TaskGroupPlayers], Title: "Players"},
		{Tasks: grouped[tasks.TaskGroupBadges], Title: "Badges"},
		{Tasks: grouped[tasks.TaskGroupNews], Title: "News"},
		{Tasks: grouped[tasks.TaskGroupElastic], Title: "Elastic"},
		{Tasks: grouped[""], Title: "Other"},
	}

	// Get configs for times
	configs, err := mysql.GetAllConfigs()
	log.Err(err, r)

	t.Configs = configs

	returnTemplate(w, r, "admin/tasks", t)
}

type adminTasksTemplate struct {
	globalTemplate
	Tasks   []adminTaskListTemplate
	Configs map[string]mysql.Config
}

type adminTaskListTemplate struct {
	Title string
	Tasks []adminTaskTemplate
}

type adminTaskTemplate struct {
	Task tasks.TaskInterface
	Bad  bool
	Next time.Time
	Prev time.Time
}

func adminSettingsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {

		err := r.ParseForm()
		if err != nil {
			log.Err(err, r)
		}

		middleware.DownMessage = r.PostFormValue("down-message")

		mcItem := r.PostFormValue("del-mc-item")
		if mcItem != "" {
			err := memcache.Delete(mcItem)
			log.Err(err, r)
		}

		err = session.SetFlash(r, sessionHelpers.SessionGood, "Done")
		log.Err(err, r)

		http.Redirect(w, r, "/admin/settings", http.StatusFound)
		return
	}

	t := adminSettingsTemplate{}
	t.fill(w, r, "Admin", "Admin")
	t.DownMessage = middleware.DownMessage

	returnTemplate(w, r, "admin/settings", t)
}

type adminSettingsTemplate struct {
	globalTemplate
	DownMessage string
}

func adminQueuesHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {

		err := r.ParseForm()
		if err != nil {
			log.Err(err, r)
		}

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

			val = strings.TrimSpace(val)

			ts, err := strconv.ParseInt(val, 10, 64)
			if err == nil {

				apps, err := steam.GetSteam().GetAppList(100000, 0, ts, "")
				err = steam.AllowSteamCodes(err)
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

				val = strings.TrimSpace(val)

				err := queue.ProduceGroup(queue.GroupMessage{ID: val, UserAgent: &ua})
				err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
				log.Err(err, r)
			}
		}

		if val := r.PostForm.Get("group-members"); val != "" {

			vals := strings.Split(val, ",")
			for _, val := range vals {

				val = strings.TrimSpace(val)

				page := 1
				for {
					resp, err := steam.GetSteam().GetGroup(val, "", page)
					err = steam.AllowSteamCodes(err)

					for _, playerID := range resp.Members.SteamID64 {

						err = queue.ProducePlayer(queue.PlayerMessage{ID: int64(playerID)})
						err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
						log.Err(err, r)

					}

					if resp.NextPageLink == "" {
						break
					}

					page++
				}
			}
		}

		err = queue.ProduceSteam(queue.SteamMessage{AppIDs: appIDs, PackageIDs: packageIDs})
		log.Err(err, r)

		err = session.SetFlash(r, sessionHelpers.SessionGood, "Done")
		log.Err(err, r)

		http.Redirect(w, r, "/admin/tasks", http.StatusFound)
		return
	}

	t := globalTemplate{}
	t.fill(w, r, "Admin", "Admin")

	returnTemplate(w, r, "admin/queues", t)
}

func adminBinLogsHandler(w http.ResponseWriter, r *http.Request) {

	g, err := mysql.GetMySQLClient()
	if err != nil {
		log.Err(err, r)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Can't connect to mysql"})
		return
	}

	deleteLog := r.URL.Query().Get("delete")
	if deleteLog != "" {

		g = g.Exec("PURGE BINARY LOGS TO '" + deleteLog + "'")
		if g.Error != nil {
			log.Err(g.Error, r)
		}

		err := session.SetFlash(r, sessionHelpers.SessionGood, "Done")
		log.Err(err, r)

		http.Redirect(w, r, "/admin/sql-bin-logs", http.StatusFound)
		return
	}

	t := adminBinLogsTemplate{}
	t.fill(w, r, "Admin", "Admin")

	g = g.Raw("show binary logs").Scan(&t.BinLogs)
	if g.Error != nil {
		log.Err(g.Error, r)
	}

	returnTemplate(w, r, "admin/binlogs", t)
}

type adminBinLogsTemplate struct {
	globalTemplate
	BinLogs []adminBinLogTemplate
}

type adminBinLogTemplate struct {
	Name      string `gorm:"column:Log_name"`
	Bytes     uint64 `gorm:"column:File_size"`
	Encrypted string `gorm:"column:Encrypted"`
	Total     uint64
}

func adminWebsocketsHandler(w http.ResponseWriter, r *http.Request) {

	t := adminWebsocketsTemplate{}
	t.fill(w, r, "Admin", "Admin")
	t.Websockets = websockets.Pages

	returnTemplate(w, r, "admin/websockets", t)
}

type adminWebsocketsTemplate struct {
	globalTemplate
	Websockets map[websockets.WebsocketPage]*websockets.Page
}
