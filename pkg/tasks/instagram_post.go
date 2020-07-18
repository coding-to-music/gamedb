package tasks

import (
	"errors"
	"math/rand"
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/instagram"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type InstagramPost struct {
	BaseTask
}

func (c InstagramPost) ID() string {
	return "post-to-instagram"
}

func (c InstagramPost) Name() string {
	return "Post an Instagram picture"
}

func (c InstagramPost) Group() string {
	return ""
}

func (c InstagramPost) Cron() string {
	return CronTimeInstagram
}

func (c InstagramPost) work() (err error) {

	filter := bson.D{
		{"type", "game"},
		{"name", bson.M{"$ne": ""}},
		{"reviews_score", bson.M{"$gte": 95}},
		{"tags", bson.M{"$nin": 12095}},
		{"screenshots.0", bson.M{"$exists": true}},
	}
	projection := bson.M{"id": 1, "name": 1, "screenshots": 1, "reviews_score": 1}

	apps, err := mongo.GetRandomApps(1, filter, projection)
	if err != nil {
		return err
	}

	if len(apps) == 0 {
		return errors.New("no apps found for instagram")
	}

	var app = apps[0]

	var url = app.Screenshots[rand.Intn(len(app.Screenshots))].PathFull
	if url == "" {
		return errors.New("empty url")
	}

	text := app.GetName() + " (Score: " + helpers.FloatToString(app.ReviewsScore, 2) + ") https://gamedb.online/games/" + strconv.Itoa(app.ID) +
		" #steamgames #steam #gaming " + helpers.GetHashTag(app.GetName())

	// err = helpers.UpdateBio("https://gamedb.online" + app.GetPath())
	// log.Err(err)

	return instagram.UploadInstagram(url, text)
}
