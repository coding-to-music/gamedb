package db

import (
	"time"

	"cloud.google.com/go/datastore"
)

type AppOverTime struct {
	AppID           int       `datastore:"app_id"`
	CreatedAt       time.Time `datastore:"created_at"`
	Score           float64   `datastore:"score"`
	ReviewsPositive int       `datastore:"reviews_positive"`
	ReviewsNegative int       `datastore:"reviews_negative"`
}

func (p AppOverTime) GetKey() (key *datastore.Key) {
	return datastore.IncompleteKey(KindAppOverTime, nil)
}

func SaveAppOverTime(app App) (err error) {

	aot := new(AppOverTime)
	aot.AppID = app.ID
	aot.CreatedAt = time.Now()
	aot.Score = app.ReviewsScore
	aot.ReviewsPositive = app.ReviewsPositive
	aot.ReviewsNegative = app.ReviewsNegative

	_, err = SaveKind(aot.GetKey(), aot)
	return err
}

func GetAppOverTimes(appID int64) (scores []AppOverTime, err error) {

	client, ctx, err := GetDSClient()
	if err != nil {
		return
	}

	q := datastore.NewQuery(KindAppOverTime)
	q = q.Filter("app_id =", appID)
	q = q.Order("created_at")

	_, err = client.GetAll(ctx, q, &scores)
	return
}
