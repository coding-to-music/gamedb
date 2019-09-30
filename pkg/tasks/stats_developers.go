package tasks

import (
	"strconv"
	"strings"
	"sync"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/sql"
)

type Developers struct {
	BaseTask
}

func (c Developers) ID() string {
	return "update-developer-stats"
}

func (c Developers) Name() string {
	return "Update developers"
}

func (c Developers) Cron() string {
	return "0 4"
}

func (c Developers) work() {

	// Get current developers, to delete old ones
	allDevelopers, err := sql.GetAllDevelopers([]string{"id", "name"})
	if err != nil {
		cronLogErr(err)
		return
	}

	developersToDelete := map[int]bool{}
	for _, v := range allDevelopers {
		developersToDelete[v.ID] = true
	}

	var developersNameMap = map[int]string{}
	for _, v := range allDevelopers {
		developersNameMap[v.ID] = strings.TrimSpace(v.GetName())
	}

	// Get apps from mysql
	appsWithDevelopers, err := sql.GetAppsWithColumnDepth("developers", 2, []string{"developers", "prices", "reviews_score"})
	cronLogErr(err)

	cronLogInfo("Found " + strconv.Itoa(len(appsWithDevelopers)) + " apps with developers")

	newDevelopers := make(map[int]*statsRow)
	for _, app := range appsWithDevelopers {

		appDevelopers, err := app.GetDeveloperIDs()
		if err != nil {
			cronLogErr(err)
			continue
		}

		if len(appDevelopers) == 0 {
			// appDevelopers = []string{""}
		}

		// For each developer in an app
		for _, appDeveloperID := range appDevelopers {

			delete(developersToDelete, appDeveloperID)

			var developersName string
			if val, ok := developersNameMap[appDeveloperID]; ok {
				developersName = val
			} else {
				continue
			}

			if _, ok := newDevelopers[appDeveloperID]; ok {
				newDevelopers[appDeveloperID].count++
				newDevelopers[appDeveloperID].totalScore += app.ReviewsScore
			} else {
				newDevelopers[appDeveloperID] = &statsRow{
					name:       developersName,
					count:      1,
					totalPrice: map[steam.ProductCC]int{},
					totalScore: app.ReviewsScore,
				}
			}

			for _, code := range helpers.GetProdCCs(true) {
				price := app.GetPrice(code.ProductCode)
				if price.Exists {
					newDevelopers[appDeveloperID].totalPrice[code.ProductCode] += price.Final
				}
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old developers
	limit++
	wg.Add(1)
	go func() {

		defer func() {
			limit--
			wg.Done()
		}()

		var devsToDeleteSlice []int
		for k := range developersToDelete {
			devsToDeleteSlice = append(devsToDeleteSlice, k)
		}

		err := sql.DeleteDevelopers(devsToDeleteSlice)
		cronLogErr(err)

	}()

	wg.Wait()

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		cronLogErr(err)
		return
	}

	// Update current developers
	var count = 1
	for k, v := range newDevelopers {

		if limit >= 2 {
			wg.Wait()
		}

		statsLogger("developer", count, len(newDevelopers), v.name)

		limit++
		wg.Add(1)
		go func(developerInt int, v *statsRow) {

			defer func() {
				limit--
				wg.Done()
			}()

			var developer sql.Developer

			gorm = gorm.Unscoped().FirstOrInit(&developer, sql.Developer{ID: developerInt})
			cronLogErr(gorm.Error)

			developer.Name = v.name
			developer.Apps = v.count
			developer.MeanPrice = v.getMeanPrice()
			developer.MeanScore = v.getMeanScore()
			developer.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&developer)
			cronLogErr(gorm.Error)

		}(k, v)

		count++
	}
	wg.Wait()

	//
	err = helpers.RemoveKeyFromMemCacheViaPubSub(helpers.MemcacheDeveloperKeyNames.Key)
	cronLogErr(err)
}
