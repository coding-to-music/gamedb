package storage

// https://github.com/GoogleCloudPlatform/google-cloud-go/blob/master/storage/example_test.go
// https://github.com/GoogleCloudPlatform/golang-samples/blob/master/storage/objects/main.go

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/steam-authority/steam-authority/logger"
)

var (
	bucket = os.Getenv("STEAM_GOOGLE_BUCKET")
	client *storage.Client
)

var (
	PathGames   = func(playerID int) (string) { return "/player-games/" + strconv.Itoa(playerID) + ".json" }
	PathBadges  = func(playerID int) (string) { return "/player-badges/" + strconv.Itoa(playerID) + ".json" }
	PathFriends = func(playerID int) (string) { return "/player-friends/" + strconv.Itoa(playerID) + ".json" }
)

func getClient() (c *storage.Client, ctx context.Context, err error) {

	ctx = context.Background()

	if client == nil {

		client, err = storage.NewClient(ctx)
		if err != nil {
			return client, ctx, nil
		}
	}

	return client, ctx, nil
}

func Upload(path string, data []byte, public bool) (err error) {

	// Get client
	client, ctx, err := getClient()
	if err != nil {
		return err
	}

	//
	object := client.Bucket(bucket).Object(path)

	// Upload bytes
	wc := object.NewWriter(ctx)
	if _, err = io.Copy(wc, bytes.NewReader(data)); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}

	// Make public
	if public {
		if err := object.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
			logger.Error(err)
		}
	}

	return nil
}

func Download(path string) (bytes []byte, err error) {

	// Get client
	client, ctx, err := getClient()
	if err != nil {
		return nil, err
	}

	rc, err := client.Bucket(bucket).Object(path).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	return data, nil
}
