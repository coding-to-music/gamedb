package tasks

import (
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
)

type Instagram struct {
	BaseTask
}

func (c Instagram) ID() string {
	return "post-to-instagram"
}

func (c Instagram) Name() string {
	return "Post an Instagram picture"
}

func (c Instagram) Cron() string {
	return "0 12"
}

func (c Instagram) work() {

	operation := func() (err error) {

		gorm, err := sql.GetMySQLClient()
		if err != nil {
			return err
		}

		gorm = gorm.Select([]string{"id", "name", "screenshots", "reviews_score"})
		gorm = gorm.Where("JSON_DEPTH(screenshots) = ?", 3)
		gorm = gorm.Where("name != ?", "")
		gorm = gorm.Where("type = ?", "game")
		gorm = gorm.Where("reviews_score >= ?", 90)
		gorm = gorm.Order("RAND()")
		gorm = gorm.Limit(1)

		var apps []sql.App
		gorm = gorm.First(&apps)
		if gorm.Error != nil {
			return gorm.Error
		}

		if len(apps) == 0 {
			return errors.New("no apps found for instagram")
		}

		var app = apps[0]

		screenshots, err := app.GetScreenshots()
		if err != nil {
			return err
		}

		var url = screenshots[rand.Intn(len(screenshots))].PathFull
		if url == "" {
			return errors.New("empty url")
		}

		text := app.GetName() + " (Score: " + helpers.FloatToString(app.ReviewsScore, 2) + ") https://gamedb.online/apps/" + strconv.Itoa(app.ID) +
			" #steamgames #steam #gaming " + helpers.GetHashTag(app.GetName())

		err = helpers.UpdateBio("https://gamedb.online" + app.GetPath())
		log.Err(err)

		return helpers.UploadInstagram(url, text)
	}

	policy := backoff.NewExponentialBackOff()
	policy.InitialInterval = time.Second * 10

	err := backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err) })
	log.Critical(err)
}
