package crons

import (
	"math/rand"
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
)

type Instagram struct {
}

func (c Instagram) ID() CronEnum {
	return CronInstagram
}

func (c Instagram) Name() string {
	return "Post an Instagram picture"
}

func (c Instagram) Config() sql.ConfigType {
	return sql.ConfInstagram
}

func (c Instagram) Work() {

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
		c.Work()
		return
	}

	err = helpers.UploadInstagram(url, app.GetName()+" (Score: "+helpers.FloatToString(app.ReviewsScore, 2)+") https://gamedb.online/apps/"+strconv.Itoa(app.ID)+" #steamgames #steam #gaming "+helpers.GetHashTag(app.GetName()))
	if err != nil {
		log.Critical(err, url)
	}
}
