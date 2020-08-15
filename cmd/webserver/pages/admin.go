package pages

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/patreon-go/patreon"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/middleware"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

func AdminRouter() http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.MiddlewareAuthCheck())
	r.Use(middleware.MiddlewareAdminCheck(Error404Handler))

	r.Get("/", adminHandler)
	r.Get("/tasks", adminTasksHandler)
	r.Get("/users", adminUsersHandler)
	r.Get("/users.json", adminUsersAjaxHandler)
	r.Get("/consumers", adminConsumersHandler)
	r.Get("/consumers.json", adminConsumersAjaxHandler)
	r.Get("/patreon", adminPatreonHandler)
	r.Get("/patreon.json", adminPatreonAjaxHandler)
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

func adminUsersHandler(w http.ResponseWriter, r *http.Request) {

	t := adminUsersTemplate{}
	t.fill(w, r, "Admin", "Admin")

	returnTemplate(w, r, "admin/users", t)
}

type adminUsersTemplate struct {
	globalTemplate
}

func adminUsersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	var query = datatable.NewDataTableQuery(r, false)
	var wg sync.WaitGroup
	var users []mysql.User

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		db, err := mysql.GetMySQLClient()
		if err != nil {
			zap.S().Error(err)
			return
		}

		db = db.Model(&mysql.User{})
		db = db.Select([]string{"created_at", "email", "email_verified", "steam_id", "level", "logged_in_at"})
		db = db.Limit(100)
		db = db.Offset(query.GetOffset())

		sortCols := map[string]string{
			"1": "created_at",
			"2": "created_at",
			"4": "level",
		}
		db = query.SetOrderOffsetGorm(db, sortCols)

		db = db.Find(&users)

		zap.S().Error(db.Error, r)
	}(r)

	// Get total
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		db, err := mysql.GetMySQLClient()
		if err != nil {
			zap.S().Error(err)
			return
		}

		db = db.Table("users").Count(&count)
		if db.Error != nil {
			zap.S().Error(db.Error, r)
			return
		}
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, count, count, nil)
	for _, user := range users {
		response.AddRow([]interface{}{
			user.CreatedAt.Format(helpers.DateSQL),  // 0
			user.Email,                              // 1
			user.EmailVerified,                      // 2
			user.SteamID.String,                     // 3 - Must be string
			user.Level,                              // 4
			user.LoggedInAt.Format(helpers.DateSQL), // 5
		})
	}

	returnJSON(w, r, response)
}

func adminConsumersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	var query = datatable.NewDataTableQuery(r, false)
	var wg sync.WaitGroup
	var consumers []mysql.Consumer

	wg.Add(1)
	go func() {

		defer wg.Done()

		db, err := mysql.GetMySQLClient()
		if err != nil {
			zap.S().Error(err)
			return
		}

		db = db.Model(&mysql.Consumer{})
		db = db.Select([]string{"expires", "owner", "environment", "version", "commits", "ip"})
		db = db.Limit(100)
		db = db.Offset(query.GetOffset())

		sortCols := map[string]string{
			"0": "expires",
			"4": "commits",
		}
		db = query.SetOrderOffsetGorm(db, sortCols)

		db = db.Find(&consumers)

		zap.S().Error(db.Error, r)
	}()

	// Get total
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		db, err := mysql.GetMySQLClient()
		if err != nil {
			zap.S().Error(err)
			return
		}

		db = db.Table("consumers").Count(&count)
		if db.Error != nil {
			zap.S().Error(db.Error, r)
			return
		}
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, count, count, nil)
	for _, consumer := range consumers {

		expires := consumer.Expires.Format(helpers.DateSQL)
		inDate := consumer.Expires.Add(mysql.ConsumerSessionLength).After(time.Now())

		response.AddRow([]interface{}{
			expires,              // 0
			consumer.Owner,       // 1
			consumer.Environment, // 2
			consumer.Version,     // 3
			consumer.Commits,     // 4
			consumer.IP,          // 5
			inDate,               // 6
		})
	}

	returnJSON(w, r, response)
}

func adminConsumersHandler(w http.ResponseWriter, r *http.Request) {

	t := adminConsumersTemplate{}
	t.fill(w, r, "Admin", "Admin")

	returnTemplate(w, r, "admin/consumers", t)
}

type adminConsumersTemplate struct {
	globalTemplate
}

func adminPatreonHandler(w http.ResponseWriter, r *http.Request) {

	t := adminPatreonTemplate{}
	t.fill(w, r, "Admin", "Admin")

	returnTemplate(w, r, "admin/patreon", t)
}

func adminPatreonAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	var wg sync.WaitGroup

	// Get webhooks
	var webhooks []mongo.PatreonWebhook
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		webhooks, err = mongo.GetPatreonWebhooks(query.GetOffset64(), 100, bson.D{{"created_at", -1}}, nil, nil)
		if err != nil {
			zap.S().Error(err)
		}
	}()

	// Get count
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionPatreonWebhooks, nil, 0)
		if err != nil {
			zap.S().Error(err)
		}
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, count, count, nil)
	for _, app := range webhooks {

		wh, err := patreon.Unmarshal([]byte(app.RequestBody))
		zap.S().Error(err)

		response.AddRow([]interface{}{
			app.CreatedAt.Format(helpers.DateSQL), // 0
			app.Event,                             // 1
			wh,                                    // 2
		})
	}

	returnJSON(w, r, response)
}

type adminPatreonTemplate struct {
	globalTemplate
}

func adminTasksHandler(w http.ResponseWriter, r *http.Request) {

	task := r.URL.Query().Get("run")
	if task != "" {

		c := r.URL.Query().Get("run")

		if val, ok := tasks.TaskRegister[c]; ok {
			go tasks.Run(val)
		}

		session.SetFlash(r, session.SessionGood, "Done")
		session.Save(w, r)

		http.Redirect(w, r, "/admin/tasks", http.StatusFound)
		return
	}

	//
	t := adminTasksTemplate{}
	t.fill(w, r, "Admin", "Admin")
	t.hideAds = true

	var grouped = map[tasks.TaskGroup][]adminTaskTemplate{}

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
	zap.S().Error(err)

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
			zap.S().Error(err)
		}

		middleware.DownMessage = r.PostFormValue("down-message")

		mcItem := r.PostFormValue("del-mc-item")
		if mcItem != "" {
			err := memcache.Delete(mcItem)
			zap.S().Error(err)
		}

		session.SetFlash(r, session.SessionGood, "Done")
		session.Save(w, r)

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
			zap.S().Error(err)
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

			zap.S().Info("Queueing apps")

			val = strings.TrimSpace(val)

			ts, err := strconv.ParseInt(val, 10, 64)
			if err == nil {

				apps, err := steam.GetSteam().GetAppList(100000, 0, ts, "")
				err = steam.AllowSteamCodes(err)
				zap.S().Error(err)
				if err == nil {

					zap.S().Info("Found " + strconv.Itoa(len(apps.Apps)) + " apps")

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
					zap.S().Error(err)
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
					zap.S().Error(err)
				}
			}
		}

		if val := r.PostForm.Get("test-id"); val != "" {

			val = strings.TrimSpace(val)
			count, err := strconv.Atoi(val)
			zap.S().Error(err)

			for i := 1; i <= count; i++ {

				err = queue.ProduceTest(i)
				err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
				zap.S().Error(err)
			}
		}

		if val := r.PostForm.Get("group-id"); val != "" {

			vals := strings.Split(val, ",")

			for _, val := range vals {

				val = strings.TrimSpace(val)

				err := queue.ProduceGroup(queue.GroupMessage{ID: val, UserAgent: &ua})
				err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
				zap.S().Error(err)
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
					zap.S().Error(err)

					for _, playerID := range resp.Members.SteamID64 {

						err = queue.ProducePlayer(queue.PlayerMessage{ID: int64(playerID), SkipExistingPlayer: true})
						err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
						zap.S().Error(err)

					}

					if resp.NextPageLink == "" {
						break
					}

					page++
				}
			}
		}

		err = queue.ProduceSteam(queue.SteamMessage{AppIDs: appIDs, PackageIDs: packageIDs})
		zap.S().Error(err)

		session.SetFlash(r, session.SessionGood, "Done")
		session.Save(w, r)

		http.Redirect(w, r, "/admin/tasks", http.StatusFound)
		return
	}

	t := globalTemplate{}
	t.fill(w, r, "Admin", "Admin")

	returnTemplate(w, r, "admin/queues", t)
}

func adminBinLogsHandler(w http.ResponseWriter, r *http.Request) {

	db, err := mysql.GetMySQLClient()
	if err != nil {
		zap.S().Error(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Can't connect to mysql"})
		return
	}

	deleteLog := r.URL.Query().Get("delete")
	if deleteLog != "" {

		db = db.Exec("PURGE BINARY LOGS TO '" + deleteLog + "'")
		if db.Error != nil {
			zap.S().Error(db.Error, r)
		}

		session.SetFlash(r, session.SessionGood, "Done")
		session.Save(w, r)

		http.Redirect(w, r, "/admin/sql-bin-logs", http.StatusFound)
		return
	}

	t := adminBinLogsTemplate{}
	t.fill(w, r, "Admin", "Admin")

	db = db.Raw("show binary logs").Scan(&t.BinLogs)
	if db.Error != nil {
		zap.S().Error(db.Error, r)
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

	for _, v := range websockets.Pages {
		t.Total += v.CountConnections()
	}

	returnTemplate(w, r, "admin/websockets", t)
}

type adminWebsocketsTemplate struct {
	globalTemplate
	Websockets map[websockets.WebsocketPage]*websockets.Page
	Total      int
}
