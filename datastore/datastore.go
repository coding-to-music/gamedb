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
	KindPlayerApp      = "PlayerApp"
	KindPrice          = "Price"
	KindRank           = "Rank"
	KindAppReviewScore = "AppReviewScore"
)

var (
	ErrNoSuchEntity = datastore.ErrNoSuchEntity
	ErrorTooMany    = errors.New("datastore: too many")
)

var (
	client *datastore.Client
)

type kind interface {
	GetKey() *datastore.Key
}

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

func bulkAddKind(kinds []kind) (err error) {

	if len(kinds) == 0 {
		return nil
	}

	client, ctx, err := getClient()
	if err != nil {
		return err
	}

	chunks := chunkKinds(kinds, 500)

	for _, chunk := range chunks {

		keys := make([]*datastore.Key, 0, len(chunk))
		for _, v := range chunk {
			keys = append(keys, v.GetKey())
		}

		_, err = client.PutMulti(ctx, keys, chunk)
		if err != nil {
			return err
		}
	}

	return nil
}

func chunkKinds(games []kind, chunkSize int) (divided [][]kind) {

	for i := 0; i < len(games); i += chunkSize {
		end := i + chunkSize

		if end > len(games) {
			end = len(games)
		}

		divided = append(divided, games[i:end])
	}

	return divided
}
