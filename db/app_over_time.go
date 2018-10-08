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

func SaveAppReviewScore(appID int, score float64, pos int, neg int) (err error) {

	kind := new(AppOverTime)
	kind.AppID = appID
	kind.CreatedAt = time.Now()
	kind.Score = score
	kind.ReviewsPositive = pos
	kind.ReviewsNegative = neg

	_, err = SaveKind(kind.GetKey(), kind)
	return err
}

func GetAppReviewScores(appID int64) (scores []AppOverTime, err error) {

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
