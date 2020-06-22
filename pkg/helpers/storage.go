package helpers

import (
	"context"
	"io/ioutil"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gamedb/gamedb/pkg/config"
	"golang.org/x/oauth2/google"
)

const BucketChatBot = "chat-bot-attachments"

var (
	storageClient  *storage.Client
	storageContext context.Context
	storageLock    sync.Mutex
)

func GetStorageClient() (c *storage.Client, ctx context.Context, err error) {

	storageLock.Lock()
	defer storageLock.Unlock()

	if storageClient == nil {
		storageContext = context.Background()
		storageClient, err = storage.NewClient(storageContext)
	}

	return storageClient, storageContext, err
}

var (
	signedOptions     *storage.SignedURLOptions
	signedOptionsLock sync.Mutex
)

func GetSignedURLOptions() (*storage.SignedURLOptions, error) {

	signedOptionsLock.Lock()
	defer signedOptionsLock.Unlock()

	if signedOptions == nil {

		var path = "/root/"
		if config.IsLocal() {
			path = config.Config.InfraPath.Get() + "/"
		}

		jsonKey, err := ioutil.ReadFile(path + "google-auth.json")
		if err != nil {
			return nil, err
		}

		conf, err := google.JWTConfigFromJSON(jsonKey)
		if err != nil {
			return nil, err
		}

		signedOptions = &storage.SignedURLOptions{
			Scheme:         storage.SigningSchemeV4,
			Method:         "GET",
			GoogleAccessID: conf.Email,
			PrivateKey:     conf.PrivateKey,
			Expires:        time.Now().Add(time.Hour * 24),
		}
	}

	return signedOptions, nil
}
