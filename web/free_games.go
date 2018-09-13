package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/memcache"
	"github.com/steam-authority/steam-authority/mysql"
)

const (
	freeGamesLimit = 500
)

func FreeGamesHandler(w http.ResponseWriter, r *http.Request) {

	// Get page number
	page, err := strconv.Atoi(r.URL.Query().Get("p"))
	if err != nil {
		page = 1
	}

	var wg sync.WaitGroup

	// Get apps
	var apps []mysql.App
	wg.Add(1)
	go func() {

		db, err := mysql.GetDB()
		if err != nil {

			logger.Error(err)

		} else {

			db = db.Select([]string{"id", "name", "icon", "type", "platforms", "reviews_score"})
			db = db.Where("is_free = ?", "1")
			db = db.Order("reviews_score DESC, name ASC")
			db = db.Limit(freeGamesLimit)
			db = db.Offset((page - 1) * freeGamesLimit)
			db = db.Find(&apps)

			logger.Error(db.Error)
		}

		wg.Done()
	}()

	// Get type
	var types []freeGameType
	wg.Add(1)
	go func() {

		db, err := mysql.GetDB()
		if err != nil {

			logger.Error(err)

		} else {

			db = db.Select([]string{"type", "count(type) as count"})
			db = db.Where("is_free = ?", "1")
			db = db.Table("apps")
			db = db.Group("type")
			db = db.Order("count DESC")
			db = db.Find(&types)

			logger.Error(db.Error)
		}

		wg.Done()
	}()

	// Get total
	var total int
	wg.Add(1)
	go func() {

		err = memcache.GetSet(memcache.FreeAppsCount, &total, func(total interface{}) (err error) {

			db, err := mysql.GetDB()
			if err != nil {
				return err
			}

			db = db.Model(&mysql.App{})
			db = db.Where("is_free = ?", "1")
			db = db.Count(total)

			if db.Error != nil {
				return db.Error
			}

			return nil
		})

		logger.Error(err)

		wg.Done()
	}()

	// Wait
	wg.Wait()

	// Template
	t := freeGamesTemplate{}
	t.Fill(w, r, "Free Games")
	t.Apps = apps
	t.Total = total
	t.Types = types
	t.Pagination = Pagination{
		path:  "/free-games?p=",
		page:  page,
		limit: freeGamesLimit,
		total: total,
	}

	returnTemplate(w, r, "free_games", t)
	return
}

type freeGamesTemplate struct {
	GlobalTemplate
	Apps       []mysql.App
	Types      []freeGameType
	Pagination Pagination
	Total      int
}

type freeGameType struct {
	Type  string `column:type"`
	Count int    `column:count"`
}

func (f freeGameType) GetType() string {

	if f.Type == "" {
		return "unknown"
	}
	return f.Type
}

func (f freeGameType) GetTypeNice() string {

	app := mysql.App{}
	app.Type = f.Type
	return app.GetType()
}

func (f freeGameType) GetCount() string {

	return humanize.Comma(int64(f.Count))
}
