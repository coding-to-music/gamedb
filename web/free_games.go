package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
)

func freeGamesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", freeGamesHandler)
	r.Get("/ajax", freeGamesAjaxHandler)
	return r
}

func freeGamesHandler(w http.ResponseWriter, r *http.Request) {

	var wg sync.WaitGroup

	// Get type
	var types []freeGameType
	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm, err := db.GetMySQLClient()
		if err != nil {

			log.Log(err)
			return
		}

		gorm = gorm.Select([]string{"type", "count(type) as count"})
		gorm = gorm.Where("is_free = ?", "1")
		gorm = gorm.Table("apps")
		gorm = gorm.Group("type")
		gorm = gorm.Order("count DESC")
		gorm = gorm.Find(&types)

		log.Log(gorm.Error)

	}()

	// Wait
	wg.Wait()

	// Template
	t := freeGamesTemplate{}
	t.Fill(w, r, "Free Games")
	t.Types = types

	err := returnTemplate(w, r, "free_games", t)
	log.Log(err)
}

type freeGamesTemplate struct {
	GlobalTemplate
	Types []freeGameType
}

type freeGameType struct {
	Type  string `column:"type"`
	Count int    `column:"count"`
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

func freeGamesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	log.Log(err)

	//
	var wg sync.WaitGroup

	// Get apps
	var filtered int
	var apps []db.App

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		gorm, err := db.GetMySQLClient()
		if err != nil {

			log.Log(err)
			return
		}

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

		log.Log(gorm.Error)

	}(r)

	// Get total
	var count int
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		count, err = helpers.GetMemcache().GetSetInt(helpers.MemcacheFreeAppsCount, func() (count int, err error) {

			gorm, err := db.GetMySQLClient()
			if err != nil {
				return count, err
			}

			gorm = gorm.Model(&db.App{})
			gorm = gorm.Where("is_free = ?", "1")
			gorm = gorm.Count(&count)

			return count, gorm.Error
		})

		log.Log(err)

	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(filtered)
	response.Draw = query.Draw

	for _, v := range apps {

		platforms, err := v.GetPlatformImages()
		log.Log(err)

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
