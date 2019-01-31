package social

import (
	"math/rand"
	"net/http"
	"strconv"

	"github.com/ahmdrz/goinsta"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/robfig/cron"
)

var (
	instagram *goinsta.Instagram
)

func RunInstagram() {

	c := cron.New()
	err := c.AddFunc("0 0 12 * * *", uploadInstagram)

	if err != nil {
		log.Critical(err)
		return
	}

	c.Start()
}

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

func uploadInstagram() {

	gorm, err := db.GetMySQLClient()
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

	var apps []db.App
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
		uploadInstagram()
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
		log.Err(err)
		return
	}

	_, err = ig.UploadPhoto(resp.Body, app.GetName()+" (Score: "+helpers.FloatToString(app.ReviewsScore, 2)+", ID: "+strconv.Itoa(app.ID)+") #steamgames", 0, 0)
	if err != nil {
		log.Err(err)
		return
	}
}
