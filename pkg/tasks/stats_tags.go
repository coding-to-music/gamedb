package tasks

import (
	"strconv"
	"strings"
	"sync"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	steamHelper "github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"go.mongodb.org/mongo-driver/bson"
)

type Tags struct {
	BaseTask
}

func (c Tags) ID() string {
	return "update-tags-stats"
}

func (c Tags) Name() string {
	return "Update tags"
}

func (c Tags) Cron() string {
	return CronTimeTags
}

func (c Tags) work() (err error) {

	// Get current tags, to delete old ones
	tags, err := sql.GetAllTags()
	if err != nil {
		return err
	}

	tagsToDelete := map[int]int{}
	for _, tag := range tags {
		tagsToDelete[tag.ID] = tag.ID
	}

	// Get tags from Steam
	tagsResp, b, err := steamHelper.GetSteam().GetTags()
	err = steamHelper.AllowSteamCodes(err, b, nil)
	if err != nil {
		return err
	}

	steamTagMap := tagsResp.GetMap()

	appsWithTags, err := mongo.GetNonEmptyArrays("tags", bson.M{"tags": 1, "prices": 1, "reviews_score": 1})
	if err != nil {
		return err
	}

	log.Info("Found " + strconv.Itoa(len(appsWithTags)) + " apps with tags")

	newTags := make(map[int]*statsRow)
	for _, app := range appsWithTags {

		// For each tag in an app
		for _, tagID := range app.Tags {

			delete(tagsToDelete, tagID)

			if _, ok := newTags[tagID]; ok {
				newTags[tagID].count++
				newTags[tagID].totalScore += app.ReviewsScore
			} else {
				newTags[tagID] = &statsRow{
					name:       strings.TrimSpace(steamTagMap[tagID]),
					count:      1,
					totalPrice: map[steamapi.ProductCC]int{},
					totalScore: app.ReviewsScore,
				}
			}

			for _, code := range helpers.GetProdCCs(true) {
				price := app.Prices.Get(code.ProductCode)
				if price.Exists {
					newTags[tagID].totalPrice[code.ProductCode] += price.Final
				}
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old tags
	limit++
	wg.Add(1)
	go func() {

		defer func() {
			limit--
			wg.Done()
		}()

		var tagsToDeleteSlice []int
		for _, v := range tagsToDelete {
			tagsToDeleteSlice = append(tagsToDeleteSlice, v)
		}

		err := sql.DeleteTags(tagsToDeleteSlice)
		log.Err(err)
	}()

	wg.Wait()

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		return err
	}

	// Update current tags
	var count = 1
	for k, v := range newTags {

		if limit >= 2 {
			wg.Wait()
		}

		statsLogger("tag", count, len(newTags), v.name)

		limit++
		wg.Add(1)
		go func(tagID int, v *statsRow) {

			defer func() {
				limit--
				wg.Done()
			}()

			var tag sql.Tag

			gorm = gorm.Unscoped().FirstOrInit(&tag, sql.Tag{ID: tagID})
			log.Err(gorm.Error)

			tag.Name = v.name
			tag.Apps = v.count
			tag.MeanPrice = v.getMeanPrice()
			tag.MeanScore = v.getMeanScore()
			tag.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&tag)
			log.Err(gorm.Error)

		}(k, v)

		count++
	}
	wg.Wait()

	// Clear cache
	return memcache.RemoveKeyFromMemCacheViaPubSub(memcache.MemcacheTagKeyNames.Key)
}
