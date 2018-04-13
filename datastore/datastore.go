// Docs: https://github.com/GoogleCloudPlatform/google-cloud-go/blob/master/datastore/example_test.go

package datastore

import (
	"context"
	"errors"
	"os"

	"cloud.google.com/go/datastore"
)

const (
	KindArticle  = "Article"
	KindChange   = "Change"
	KindConfig   = "Config"
	KindDonation = "Donation"
	KindLogin    = "Login"
	KindPlayer   = "Player"
	KindPrice    = "Price"
	KindRank     = "Rank"
)

var (
	ErrNoSuchEntity = datastore.ErrNoSuchEntity
	ErrorTooMany    = errors.New("datastore: too many")
)

var (
	client *datastore.Client
)

func getClient() (ret *datastore.Client, ctx context.Context, err error) {

	ctx = context.Background()

	if client == nil {
		client, err = datastore.NewClient(ctx, os.Getenv("STEAM_GOOGLE_PROJECT"))
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
