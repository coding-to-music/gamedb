package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/memcache"
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

		gorm, err := db.GetMySQLClient()
		if err != nil {

			logger.Error(err)

		} else {

			gorm = gorm.Select([]string{"type", "count(type) as count"})
			gorm = gorm.Where("is_free = ?", "1")
			gorm = gorm.Table("apps")
			gorm = gorm.Group("type")
			gorm = gorm.Order("count DESC")
			gorm = gorm.Find(&types)

			logger.Error(gorm.Error)
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

	app := db.App{}
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
	var apps []db.App

	wg.Add(1)
	go func() {

		gorm, err := db.GetMySQLClient()
		if err != nil {

			logger.Error(err)

		} else {

			gorm = gorm.Model(&db.App{})
			gorm = gorm.Select([]string{"id", "name", "icon", "type", "platforms", "reviews_score"})
			gorm = gorm.Where("is_free = ?", "1")
			gorm = gorm.Where("name LIKE ?", "%"+query.GetSearch()+"%")

			gorm = query.SetOrderOffsetGorm(gorm, map[string]string{
				"0": "name",
				"1": "reviews_score",
				"2": "type",
			})

			gorm = gorm.Count(&filtered)

			gorm = gorm.Limit(freeGamesLimit)

			gorm = gorm.Find(&apps)

			logger.Error(gorm.Error)
		}

		wg.Done()
	}()

	// Get total
	var total int
	wg.Add(1)
	go func() {

		err := memcache.GetSet(memcache.FreeAppsCount, &total, func(total interface{}) (err error) {

			gorm, err := db.GetMySQLClient()
			if err != nil {
				return err
			}

			gorm = gorm.Model(&db.App{})
			gorm = gorm.Where("is_free = ?", "1")
			gorm = gorm.Count(total)

			if gorm.Error != nil {
				return gorm.Error
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
