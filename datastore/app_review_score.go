package datastore

import (
	"time"

	"cloud.google.com/go/datastore"
)

type AppReviewScore struct {
	CreatedAt       *time.Time `datastore:"created_at"`
	ID              int64      `datastore:"id"`
	AppID           string     `datastore:"app_id"`
	ReviewsScore    float64    `datastore:"reviews_score"`
	Reviews         int64      `datastore:"reviews"`
	ReviewsPositive int64      `datastore:"reviews_positive"`
	ReviewsNegative int64      `datastore:"reviews_negative"`
}

func (p AppReviewScore) GetKey() (key *datastore.Key) {
	return datastore.IncompleteKey(KindAppReviewScore, nil)
}
