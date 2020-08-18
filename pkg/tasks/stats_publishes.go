package tasks

import (
	"strconv"
	"strings"
	"sync"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type TasksPublishers struct {
	BaseTask
}

func (c TasksPublishers) ID() string {
	return "update-publisher-stats"
}

func (c TasksPublishers) Name() string {
	return "Update publishers"
}

func (c TasksPublishers) Group() TaskGroup {
	return ""
}

func (c TasksPublishers) Cron() TaskTime {
	return CronTimePublishers
}

func (c TasksPublishers) work() (err error) {

	// Get current publishers, to delete old ones
	allPublishers, err := mysql.GetAllPublishers()
	if err != nil {
		return err
	}

	publishersToDelete := map[int]bool{}
	for _, publisherRow := range allPublishers {
		publishersToDelete[publisherRow.ID] = true
	}

	var publisherNameMap = map[int]string{}
	for _, v := range allPublishers {
		publisherNameMap[v.ID] = strings.TrimSpace(v.GetName())
	}

	// Get apps from mysql
	appsWithPublishers, err := mongo.GetNonEmptyArrays(0, 0, "publishers", bson.M{"publishers": 1, "prices": 1, "reviews_score": 1})
	if err != nil {
		return err
	}

	zap.S().Info("Found " + strconv.Itoa(len(appsWithPublishers)) + " apps with publishers")

	newPublishers := make(map[int]*statsRow)
	for _, app := range appsWithPublishers {

		if len(app.Publishers) == 0 {
			// appPublishers = []string{""}
		}

		// For each publisher in an app
		for _, appPublisherID := range app.Publishers {

			delete(publishersToDelete, appPublisherID)

			var publisherName string
			if val, ok := publisherNameMap[appPublisherID]; ok {
				publisherName = val
			} else {
				// publisherName = "Unknown"
				continue
			}

			if _, ok := newPublishers[appPublisherID]; ok {
				newPublishers[appPublisherID].count++
				newPublishers[appPublisherID].totalScore += app.ReviewsScore
			} else {
				newPublishers[appPublisherID] = &statsRow{
					name:       publisherName,
					count:      1,
					totalPrice: map[steamapi.ProductCC]int{},
					totalScore: app.ReviewsScore,
				}
			}

			for _, code := range i18n.GetProdCCs(true) {
				price := app.Prices.Get(code.ProductCode)
				if price.Exists {
					newPublishers[appPublisherID].totalPrice[code.ProductCode] += price.Final
				}
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old publishers
	limit++
	wg.Add(1)
	go func() {

		defer func() {
			limit--
			wg.Done()
		}()

		var pubsToDeleteSlice []int
		for publisherID := range publishersToDelete {
			pubsToDeleteSlice = append(pubsToDeleteSlice, publisherID)
		}

		err := mysql.DeletePublishers(pubsToDeleteSlice)
		if err != nil {
			zap.S().Error(err)
		}
	}()

	wg.Wait()

	// Update current publishers
	var count = 1
	for k, v := range newPublishers {

		if limit >= 2 {
			wg.Wait()
		}

		limit++
		wg.Add(1)
		go func(publisherID int, v *statsRow) {

			defer func() {
				limit--
				wg.Done()
			}()

			var publisher mysql.Publisher

			db, err := mysql.GetMySQLClient()
			if err != nil {
				zap.S().Error(err)
				return
			}

			db = db.Unscoped().FirstOrInit(&publisher, mysql.Publisher{ID: publisherID})
			if db.Error != nil {
				zap.S().Error(db.Error)
			}

			publisher.Name = v.name
			publisher.Apps = v.count
			publisher.MeanPrice = v.getMeanPrice()
			publisher.MeanScore = v.getMeanScore()
			publisher.DeletedAt = nil

			db = db.Unscoped().Save(&publisher)
			if db.Error != nil {
				zap.S().Error(db.Error)
			}

		}(k, v)

		count++
	}

	wg.Wait()

	// Clear cache
	return memcache.Delete(
		memcache.MemcachePublisherKeyNames.Key,
	)
}
