package mongo

import (
	"errors"

	"github.com/gamedb/gamedb/pkg/helpers"
	. "go.mongodb.org/mongo-driver/bson"
)

var ErrInvalidAppID = errors.New("invalid app id")

type App struct {
	ID                            int     `bson:"_id"`
	AchievementsCount             int     `bson:"achievements_count"`
	AchievementsAverageCompletion float64 `bson:"achievements_average_completion"`
	PlaytimeTotal                 int64   `bson:"playtime_total"`   // Minutes
	PlaytimeAverage               float64 `bson:"playtime_average"` // Minutes
}

func (a App) BSON() (ret D) {

	return D{
		{"_id", a.ID},
		{"achievements_total", a.AchievementsCount},
		{"achievements_average_completion", a.AchievementsAverageCompletion},
		{"playtime_total", a.PlaytimeTotal},
		{"playtime_average", a.PlaytimeAverage},
	}
}

func (a App) Save() (err error) {

	_, err = ReplaceOne(CollectionApps, D{{"_id", a.ID}}, a)
	return err
}

func GetApp(id int) (app App, err error) {

	if !helpers.IsValidAppID(id) {
		return app, ErrInvalidAppID
	}

	err = FindOne(CollectionApps, D{{"_id", id}}, nil, nil, &app)
	if err != nil {
		return app, err
	}
	if app.ID == 0 {
		return app, ErrNoDocuments
	}

	return app, err
}
