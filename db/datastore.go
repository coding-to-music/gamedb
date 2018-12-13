// Docs: https://github.com/GoogleCloudPlatform/google-cloud-go/blob/master/datastore/example_test.go

package db

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/log"
	"github.com/spf13/viper"
)

const (
	KindAppOverTime    = "AppOverTime"
	KindChange         = "Change"
	KindDonation       = "Donation"
	KindEvent          = "Event"
	KindNews           = "News"
	KindPlayer         = "Player"
	KindPlayerApp      = "PlayerApp"
	KindPlayerOverTime = "PlayerOverTime"
	KindPlayerRank     = "PlayerRank"
	KindProductPrice   = "ProductPrice"
)

var (
	ErrorTooMany = errors.New("datastore: too many")

	datastoreClient *datastore.Client
)

// Called from main
func InitDS() {
	_, _, err := GetDSClient()
	log.Log(err)
}

type Kind interface {
	GetKey() *datastore.Key
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

	log.Info("Saving "+key.Name, log.LogNameDatastore, log.ServiceGoogle)

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

func BulkSaveKinds(kinds []Kind, kind string, wait bool) (err error) {

	if len(kinds) == 0 {
		return nil
	}

	log.Info("Bulk saving "+strconv.Itoa(len(kinds))+" "+kind+"s", log.LogNameDatastore, log.ServiceGoogle)

	client, ctx, err := GetDSClient()
	if err != nil {
		return err
	}

	var errs []error
	var wg sync.WaitGroup

	chunks := chunkKinds(kinds)
	for _, chunk := range chunks {

		wg.Add(1)
		go func(chunk []Kind, wait bool) {

			defer wg.Done()

			keys := make([]*datastore.Key, 0, len(chunk))
			for _, vv := range chunk {
				keys = append(keys, vv.GetKey())
			}

			switch kind {
			case KindNews:
				_, err = client.PutMulti(ctx, keys, kindsToNews(chunk))
			case KindPlayerApp:
				_, err = client.PutMulti(ctx, keys, kindsToPlayerApps(chunk))
			case KindChange:
				_, err = client.PutMulti(ctx, keys, kindsToChanges(chunk))
			case KindPlayerRank:
				_, err = client.PutMulti(ctx, keys, kindsToPlayerRanks(chunk))
			case KindProductPrice:
				_, err = client.PutMulti(ctx, keys, kindsToProductPrices(chunk))
			default:
				log.Log(errors.New("missing case in BulkSaveKinds"))
			}

			if err != nil {
				if wait {
					errs = append(errs, err)
				} else {
					log.Log(err)
				}
			}

		}(chunk, wait)
	}

	if wait {
		wg.Wait()

		if len(errs) > 0 {
			return errs[0]
		}
	}

	return nil
}

func chunkKinds(kinds []Kind) (chunked [][]Kind) {

	for i := 0; i < len(kinds); i += 500 {
		end := i + 500

		if end > len(kinds) {
			end = len(kinds)
		}

		chunked = append(chunked, kinds[i:end])
	}

	return chunked
}

func BulkDeleteKinds(keys []*datastore.Key, wait bool) (err error) {

	if len(keys) == 0 {
		return nil
	}

	log.Info("Bulk deleting "+strconv.Itoa(len(keys))+" keys", log.LogNameDatastore, log.ServiceGoogle)

	client, ctx, err := GetDSClient()
	if err != nil {
		return err
	}

	var errs []error
	var wg sync.WaitGroup

	chunks := chunkKeys(keys)
	for _, v := range chunks {

		wg.Add(1)
		go func() {

			defer wg.Done()

			err = client.DeleteMulti(ctx, v)
			if err != nil {
				if wait {
					errs = append(errs, err)
				} else {
					log.Log(err)
				}
			}

		}()
	}

	if wait {
		wg.Wait()

		if len(errs) > 0 {
			return errs[0]
		}
	}

	return nil
}

func chunkKeys(keys []*datastore.Key) (chunked [][]*datastore.Key) {

	for i := 0; i < len(keys); i += 500 {
		end := i + 500

		if end > len(keys) {
			end = len(keys)
		}

		chunked = append(chunked, keys[i:end])
	}

	return chunked
}

func kindsToNews(a []Kind) (b []News) {

	for _, v := range a {

		original, ok := v.(News)
		if ok {
			b = append(b, original)
		} else {
			log.Info("kind not a struct")
		}
	}

	return b
}

func kindsToPlayerApps(a []Kind) (b []PlayerApp) {

	for _, v := range a {

		original, ok := v.(PlayerApp)
		if ok {
			b = append(b, original)
		}
	}

	return b
}

func kindsToChanges(a []Kind) (b []Change) {

	for _, v := range a {

		original, ok := v.(Change)
		if ok {
			b = append(b, original)
		}
	}

	return b
}

func kindsToPlayerRanks(a []Kind) (b []PlayerRank) {

	for _, v := range a {

		original, ok := v.(PlayerRank)
		if ok {
			b = append(b, original)
		}
	}

	return b
}

func kindsToProductPrices(a []Kind) (b []ProductPrice) {

	for _, v := range a {

		original, ok := v.(ProductPrice)
		if ok {
			b = append(b, original)
		}
	}

	return b
}
