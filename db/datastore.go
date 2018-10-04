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

type Kind interface {
	GetKey() (*datastore.Key)
}

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

func BulkSaveKinds(kinds []Kind) (err error) {

	count := len(kinds)
	if count == 0 {
		return nil
	}

	client, ctx, err := GetDSClient()
	if err != nil {
		return err
	}

	var errs []error
	chunks := chunkKinds(kinds, 0)
	for _, v := range chunks {

		keys := make([]*datastore.Key, 0, len(v))
		for _, vv := range v {
			keys = append(keys, vv.GetKey())
		}

		_, err = client.PutMulti(ctx, keys, v)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

func chunkKinds(kinds []Kind, chunkSize int) (chunked [][]Kind) {

	if chunkSize == 0 {
		chunkSize = 500
	}

	for i := 0; i < len(kinds); i += chunkSize {
		end := i + chunkSize

		if end > len(kinds) {
			end = len(kinds)
		}

		chunked = append(chunked, kinds[i:end])
	}

	return chunked
}

func BulkDeleteKinds(keys map[int64]*datastore.Key, chunkSize int) (err error) {

	if len(keys) == 0 {
		return nil
	}

	if chunkSize == 0 {
		chunkSize = 500
	}

	// Make map a slice
	var keysToDelete []*datastore.Key
	for _, v := range keys {
		keysToDelete = append(keysToDelete, v)
	}

	client, ctx, err := GetDSClient()
	if err != nil {
		return err
	}

	var errs []error
	chunks := chunkKeys(keysToDelete, chunkSize)
	for _, v := range chunks {

		err = client.DeleteMulti(ctx, v)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

func chunkKeys(keys []*datastore.Key, chunkSize int) (chunked [][]*datastore.Key) {

	if chunkSize == 0 {
		chunkSize = 500
	}

	for i := 0; i < len(keys); i += chunkSize {
		end := i + chunkSize

		if end > len(keys) {
			end = len(keys)
		}

		chunked = append(chunked, keys[i:end])
	}

	return chunked
}
