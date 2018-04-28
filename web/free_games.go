package web

import (
	"net/http"
	"strconv"
	"sync"

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
			return
		}

		db = db.Limit(freeGamesLimit)
		db = db.Offset((page - 1) * freeGamesLimit)
		db = db.Order("reviews_score DESC, name ASC")
		db = db.Select([]string{"id", "name", "icon", "type", "platforms", "reviews_score"})
		db = db.Where("is_free = ?", "1")

		db = db.Find(&apps)
		if db.Error != nil {
			logger.Error(db.Error)
			return
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

			db = db.Model(&mysql.App{}).Where("is_free = ?", "1").Count(total)
			if db.Error != nil {
				return db.Error
			}

			return nil
		})

		if err != nil {
			logger.Error(err)
		}

		wg.Done()
	}()

	// Wait
	wg.Wait()

	// Template
	t := freeGamesTemplate{}
	t.Fill(w, r, "Free Games")
	t.Apps = apps
	t.Total = total
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
	Pagination Pagination
	Total      int
}
