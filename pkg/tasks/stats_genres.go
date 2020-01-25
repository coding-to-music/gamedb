package tasks

import (
	"strconv"
	"strings"
	"sync"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"go.mongodb.org/mongo-driver/bson"
)

type Genres struct {
	BaseTask
}

func (c Genres) ID() string {
	return "update-genre-stats"
}

func (c Genres) Name() string {
	return "Update genres"
}

func (c Genres) Cron() string {
	return CronTimeGenres
}

func (c Genres) work() (err error) {

	// Get current genres, to delete old ones
	currentGenres, err := sql.GetAllGenres(true)
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
	appsWithGenres, err := mongo.GetNonEmptyArrays("genres", bson.M{"genres": 1, "prices": 1, "reviews_score": 1})
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
					totalPrice: map[steam.ProductCC]int{},
				}
			}

			for _, code := range helpers.GetProdCCs(true) {
				price := app.GetPrice(code.ProductCode)
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

		err := sql.DeleteGenres(genresToDeleteSlice)
		log.Err(err)
	}()

	wg.Wait()

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		return err
	}

	// Update current genres
	var count = 1
	for k, v := range newGenres {

		if limit >= 2 {
			wg.Wait()
		}

		statsLogger("genre", count, len(newGenres), v.name)

		limit++
		wg.Add(1)
		go func(genreID int, v *statsRow) {

			defer func() {
				limit--
				wg.Done()
			}()

			var genre sql.Genre

			gorm = gorm.Unscoped().FirstOrInit(&genre, sql.Genre{ID: genreID})
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
	return memcache.RemoveKeyFromMemCacheViaPubSub(memcache.MemcacheGenreKeyNames.Key)
}
