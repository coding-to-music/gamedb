// Docs: https://github.com/GoogleCloudPlatform/google-cloud-go/blob/master/datastore/example_test.go

package datastore

import (
	"context"
	"os"

	"cloud.google.com/go/datastore"
)

const (
	KindArticle      = "Article"
	KindChange       = "Change"
	KindConfig       = "Config"
	KindDonation     = "Donation"
	KindGroup        = "Group"
	KindLogin        = "Login"
	KindPlayer       = "Player"
	KindPriceApp     = "AppPrice"
	KindPricePackage = "PackagePrice"
	KindRank         = "Rank"
)

const (
	ErrorNotFound = "datastore: no such entity"
)

func getDSClient() (client *datastore.Client, ctx context.Context, err error) {

	ctx = context.Background()
	client, err = datastore.NewClient(ctx, os.Getenv("STEAM_GOOGLE_PROJECT"))
	if err != nil {
		return client, ctx, err
	}

	return client, ctx, nil
}

func SaveKind(key *datastore.Key, data interface{}) (newKey *datastore.Key, err error) {

	client, ctx, err := getDSClient()
	if err != nil {
		return nil, err
	}

	newKey, err = client.Put(ctx, key, data)
	if err != nil {
		return newKey, err
	}

	return newKey, nil
}
