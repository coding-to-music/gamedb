package mongo

import (
	"github.com/gamedb/gamedb/pkg/helpers"
)

type App struct {
	ID                            int     `bson:"_id"`
	AchievementsTotal             int     `bson:"achievements_total"`
	AchievementsAverageCompletion float64 `bson:"achievements_average_completion"`
	PlaytimeTotal                 int64   `bson:"playtime_total"`   // Minutes
	PlaytimeAverage               float64 `bson:"playtime_average"` // Minutes
}

func (a App) BSON() (ret interface{}) {

	return M{
		"_id":                             a.ID,
		"achievements_total":              a.AchievementsTotal,
		"achievements_average_completion": a.AchievementsAverageCompletion,
		"playtime_total":                  a.PlaytimeTotal,
		"playtime_average":                a.PlaytimeAverage,
	}
}

func (a *App) UpdatePlaytime() (err error) {

	players, err := GetAppPlayTimes(a.ID)
	if err != nil {
		return err
	}

	var minutes int64
	for _, v := range players {
		minutes += int64(v.AppTime)
	}

	a.PlaytimeTotal = minutes
	a.PlaytimeAverage = float64(minutes) / float64(len(players))

	return err
}

func (a *App) UpdateAchievements() (err error) {



	players, err := GetAppPlayTimes(a.ID)
	if err != nil {
		return err
	}

	var minutes int64
	for _, v := range players {
		minutes += int64(v.AppTime)
	}

	a.PlaytimeTotal = minutes
	a.PlaytimeAverage = float64(minutes) / float64(len(players))

	return err
}

func (a App) Save() (err error) {

	_, err = ReplaceDocument(CollectionApps, M{"_id": a.ID}, a)
	return err
}

func GetApp(id int) (app App, err error) {

	if !helpers.IsValidAppID(id) {
		return app, ErrInvalidGroupID
	}

	err = FindDocumentByKey(CollectionApps, "_id", id, nil, &app)

	return app, err
}
