package tasks

import (
	"strconv"
	"strings"
	"sync"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"go.mongodb.org/mongo-driver/bson"
)

type TasksGenres struct {
	BaseTask
}

func (c TasksGenres) ID() string {
	return "update-genre-stats"
}

func (c TasksGenres) Name() string {
	return "Update genres"
}

func (c TasksGenres) Group() string {
	return ""
}

func (c TasksGenres) Cron() string {
	return CronTimeGenres
}

func (c TasksGenres) work() (err error) {

	// Get current genres, to delete old ones
	currentGenres, err := mysql.GetAllGenres(true)
	if err != nil {
		return err
	}

	genresToDelete := map[int]bool{}
	for _, v := range currentGenres {
		genresToDelete[v.ID] = true
	}

	var genreNameMap = map[int]string{}
	for _, v := range currentGenres {
		genreNameMap[v.ID] = strings.TrimSpace(v.GetName())
	}

	// Get apps from mysql
	appsWithGenres, err := mongo.GetNonEmptyArrays(0, 0, "genres", bson.M{"genres": 1, "prices": 1, "reviews_score": 1})
	if err != nil {
		return err
	}

	log.Info("Found " + strconv.Itoa(len(appsWithGenres)) + " apps with genres")

	newGenres := make(map[int]*statsRow)
	for _, app := range appsWithGenres {

		if len(app.Genres) == 0 {
			// appGenreIDs = []db.AppGenre{{ID: 0, Name: ""}}
		}

		// For each genre in an app
		for _, genreID := range app.Genres {

			delete(genresToDelete, genreID)

			var genreName string
			if val, ok := genreNameMap[genreID]; ok {
				genreName = val
			} else {
				// genreName = "Unknown"
				continue
			}

			if _, ok := newGenres[genreID]; ok {
				newGenres[genreID].count++
				newGenres[genreID].totalScore += app.ReviewsScore
			} else {
				newGenres[genreID] = &statsRow{
					name:       genreName,
					count:      1,
					totalScore: app.ReviewsScore,
					totalPrice: map[steamapi.ProductCC]int{},
				}
			}

			for _, code := range i18n.GetProdCCs(true) {
				price := app.Prices.Get(code.ProductCode)
				if price.Exists {
					newGenres[genreID].totalPrice[code.ProductCode] += price.Final
				}
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old genres
	limit++
	wg.Add(1)
	go func() {

		defer func() {
			limit--
			wg.Done()
		}()

		var genresToDeleteSlice []int
		for genreID := range genresToDelete {
			genresToDeleteSlice = append(genresToDeleteSlice, genreID)
		}

		err := mysql.DeleteGenres(genresToDeleteSlice)
		log.Err(err)
	}()

	wg.Wait()

	gorm, err := mysql.GetMySQLClient()
	if err != nil {
		return err
	}

	// Update current genres
	var count = 1
	for k, v := range newGenres {

		if limit >= 2 {
			wg.Wait()
		}

		limit++
		wg.Add(1)
		go func(genreID int, v *statsRow) {

			defer func() {
				limit--
				wg.Done()
			}()

			var genre mysql.Genre

			gorm = gorm.Unscoped().FirstOrInit(&genre, mysql.Genre{ID: genreID})
			log.Err(gorm.Error)

			genre.Name = v.name
			genre.Apps = v.count
			genre.MeanPrice = v.getMeanPrice()
			genre.MeanScore = v.getMeanScore()
			genre.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&genre)
			log.Err(gorm.Error)

		}(k, v)

		count++
	}
	wg.Wait()

	// Clear cache
	return memcache.Delete(
		memcache.MemcacheGenreKeyNames.Key,
	)
}
