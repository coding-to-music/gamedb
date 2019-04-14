package social

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/ahmdrz/goinsta/v2"
	"github.com/cenkalti/backoff"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/sql"
)

var (
	instagram *goinsta.Instagram
)

func getInstagram() (*goinsta.Instagram, error) {

	if instagram == nil {

		client := goinsta.New(
			config.Config.InstagramUsername.Get(),
			config.Config.InstagramPassword.Get(),
		)

		err := client.Login()
		if err != nil {
			return client, err
		}

		instagram = client
	}

	return instagram, nil
}

func UploadInstagram() {

	log.Info("Running IG")

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err)
		return
	}

	gorm = gorm.Select([]string{"id", "name", "screenshots", "reviews_score"})
	gorm = gorm.Where("JSON_DEPTH(screenshots) = ?", 3)
	gorm = gorm.Where("name != ?", "")
	gorm = gorm.Where("type = ?", "game")
	gorm = gorm.Where("reviews_score >= ?", 90)
	gorm = gorm.Order("RAND()")
	gorm = gorm.Limit(1)

	var apps []sql.App
	gorm = gorm.Find(&apps)
	if gorm.Error != nil {
		log.Err(gorm.Error)
		return
	}

	if len(apps) == 0 {
		log.Err("no apps found for instagram")
		return
	}

	var app = apps[0]

	screenshots, err := app.GetScreenshots()
	if err != nil {
		log.Err(err)
		return
	}

	var url = screenshots[rand.Intn(len(screenshots))].PathFull
	if url == "" {
		UploadInstagram()
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Err(err)
		return
	}

	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	ig, err := getInstagram()
	if err != nil {
		log.Critical(err)
		return
	}

	operation := func() (err error) {

		_, err = ig.UploadPhoto(resp.Body, app.GetName()+" (Score: "+helpers.FloatToString(app.ReviewsScore, 2)+") https://gamedb.online/apps/"+strconv.Itoa(app.ID)+" #steamgames #steam #gaming "+helpers.GetHashTag(app.GetName()), 0, 0)
		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.InitialInterval = time.Second * 10

	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err) })
	if err != nil {
		log.Critical(err)
	}
}
