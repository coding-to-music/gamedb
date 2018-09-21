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
	freeGamesLimit = 100
)

func FreeGamesHandler(w http.ResponseWriter, r *http.Request) {

	var wg sync.WaitGroup

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

	// Wait
	wg.Wait()

	// Template
	t := freeGamesTemplate{}
	t.Fill(w, r, "Free Games")
	t.Types = types

	returnTemplate(w, r, "free_games", t)
	return
}

type freeGamesTemplate struct {
	GlobalTemplate
	Types []freeGameType
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

func FreeGamesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	query.FillFromURL(r.URL.Query())

	//
	var wg sync.WaitGroup

	// Get apps
	var filtered int
	var apps []mysql.App

	wg.Add(1)
	go func() {

		db, err := mysql.GetDB()
		if err != nil {

			logger.Error(err)

		} else {

			db = db.Model(&mysql.App{})
			db = db.Select([]string{"id", "name", "icon", "type", "platforms", "reviews_score"})
			db = db.Where("is_free = ?", "1")
			db = db.Where("name LIKE ?", "%"+query.GetSearch()+"%")

			db = query.Query(db, map[string]string{
				"0": "name",
				"1": "reviews_score",
				"2": "type",
			})

			db = db.Count(&filtered)

			db = db.Limit(freeGamesLimit)

			db = db.Find(&apps)

			logger.Error(db.Error)
		}

		wg.Done()
	}()

	// Get total
	var total int
	wg.Add(1)
	go func() {

		err := memcache.GetSet(memcache.FreeAppsCount, &total, func(total interface{}) (err error) {

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

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(total)
	response.RecordsFiltered = strconv.Itoa(filtered)
	response.Draw = query.Draw

	for _, v := range apps {

		platforms, err := v.GetPlatformImages()
		if err != nil {
			logger.Error(err)
		}

		response.AddRow([]interface{}{
			v.ID,
			v.GetName(),
			v.GetIcon(),
			v.ReviewsScore,
			v.GetType(),
			platforms,
			v.GetInstallLink(),
			v.GetPath(),
		})
	}

	response.Output(w)
}
