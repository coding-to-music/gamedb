// Docs: https://github.com/GoogleCloudPlatform/google-cloud-go/blob/master/datastore/example_test.go

package db

import (
	"context"
	"errors"

	"cloud.google.com/go/datastore"
	"github.com/spf13/viper"
)

const (
	KindArticle        = "Article"
	KindChange         = "Change"
	KindDonation       = "Donation"
	KindEvent          = "Event"
	KindPlayer         = "Player"
	KindPlayerApp      = "PlayerApp"
	KindPlayerRankTime = "PlayerRankTime"
	KindPrice          = "Price"
	KindRank           = "Rank"
	KindAppReviewScore = "AppReviewScore"
)

var (
	ErrNoSuchEntity = datastore.ErrNoSuchEntity
	ErrorTooMany    = errors.New("datastore: too many")
)

var (
	datastoreClient *datastore.Client
)

func GetDSClient() (ret *datastore.Client, ctx context.Context, err error) {

	ctx = context.Background()

	if datastoreClient == nil {
		datastoreClient, err = datastore.NewClient(ctx, viper.GetString("GOOGLE_PROJECT"))
		if err != nil {
			return datastoreClient, ctx, err
		}
	}

	return datastoreClient, ctx, nil
}

func SaveKind(key *datastore.Key, data interface{}) (newKey *datastore.Key, err error) {

	client, ctx, err := GetDSClient()
	if err != nil {
		return nil, err
	}

	newKey, err = client.Put(ctx, key, data)
	if err != nil {
		return newKey, err
	}

	return newKey, nil
}
