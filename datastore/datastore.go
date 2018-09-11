// Docs: https://github.com/GoogleCloudPlatform/google-cloud-go/blob/master/datastore/example_test.go

package datastore

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
	KindPrice          = "Price"
	KindRank           = "Rank"
	KindAppReviewScore = "AppReviewScore"
)

var (
	ErrNoSuchEntity  = datastore.ErrNoSuchEntity
	ErrorTooMany     = errors.New("datastore: too many")
)

var (
	client *datastore.Client
)

func getClient() (ret *datastore.Client, ctx context.Context, err error) {

	ctx = context.Background()

	if client == nil {
		client, err = datastore.NewClient(ctx, viper.GetString("GOOGLE_PROJECT"))
		if err != nil {
			return client, ctx, err
		}
	}

	return client, ctx, nil
}

func SaveKind(key *datastore.Key, data interface{}) (newKey *datastore.Key, err error) {

	client, ctx, err := getClient()
	if err != nil {
		return nil, err
	}

	newKey, err = client.Put(ctx, key, data)
	if err != nil {
		return newKey, err
	}

	return newKey, nil
}
