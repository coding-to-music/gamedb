package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
	"github.com/gamedb/website/memcache"
	"github.com/gamedb/website/session"
)

func FreeGamesHandler(w http.ResponseWriter, r *http.Request) {

	var wg sync.WaitGroup

	// Get type
	var types []freeGameType
	wg.Add(1)
	go func() {

		gorm, err := db.GetMySQLClient()
		if err != nil {

			logging.Error(err)

		} else {

			gorm = gorm.Select([]string{"type", "count(type) as count"})
			gorm = gorm.Where("is_free = ?", "1")
			gorm = gorm.Table("apps")
			gorm = gorm.Group("type")
			gorm = gorm.Order("count DESC")
			gorm = gorm.Find(&types)

			logging.Error(gorm.Error)
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
	go func(r *http.Request) {

		gorm, err := db.GetMySQLClient()
		if err != nil {

			logging.Error(err)

		} else {

			gorm = gorm.Model(&db.App{})
			gorm = gorm.Select([]string{"id", "name", "icon", "type", "platforms", "reviews_score"})
			gorm = gorm.Where("is_free = ?", "1")

			search := query.GetSearchString("value")
			if search != "" {
				gorm = gorm.Where("name LIKE ?", "%"+search+"%")
			}

			types := query.GetSearchSlice("types")
			if len(types) > 0 {
				gorm = gorm.Where("type IN (?)", types)
			}

			gorm = query.SetOrderOffsetGorm(gorm, session.GetCountryCode(r), map[string]string{
				"0": "name",
				"1": "reviews_score",
			})

			gorm = gorm.Count(&filtered)

			gorm = gorm.Limit(100)

			gorm = gorm.Find(&apps)

			logging.Error(gorm.Error)
		}

		wg.Done()
	}(r)

	// Get total
	var count int
	wg.Add(1)
	go func() {

		var err error

		count, err = memcache.GetSetInt(memcache.FreeAppsCount, func() (count int, err error) {

			gorm, err := db.GetMySQLClient()
			if err != nil {
				return count, err
			}

			gorm = gorm.Model(&db.App{})
			gorm = gorm.Where("is_free = ?", "1")
			gorm = gorm.Count(&count)

			return count, gorm.Error
		})

		logging.Error(err)

		wg.Done()
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(filtered)
	response.Draw = query.Draw

	for _, v := range apps {

		platforms, err := v.GetPlatformImages()
		logging.Error(err)

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

	response.output(w)
}
