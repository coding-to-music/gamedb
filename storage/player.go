// https://github.com/GoogleCloudPlatform/google-cloud-go/blob/master/storage/example_test.go
// https://github.com/GoogleCloudPlatform/golang-samples/blob/master/storage/objects/main.go

package storage

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/steam-authority/steam-authority/logger"
)

const (
	FolderImg460x215 = "img460x215"
)

func UploadPlayerGames(playerID int, gameBytes []byte) string {

	playerIDstring := strconv.Itoa(playerID)

	ctx := context.Background()

	// Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		logger.Error(err)
	}

	// Save image to bucket
	fileName := "player-games/" + playerIDstring + ".json"

	wc := client.Bucket(bucket).Object(fileName).NewWriter(ctx)
	if _, err = io.Copy(wc, bytes.NewReader(gameBytes)); err != nil {
		logger.Error(err)
	}
	if err := wc.Close(); err != nil {
		logger.Error(err)
	}

	//
	return "/" + fileName
}

func DownloadPlayerGames(playerID int) (bytes []byte, err error) {

	ctx := context.Background()

	// Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		logger.Error(err)
	}

	rc, err := client.Bucket(bucket).Object("player-games/" + strconv.Itoa(playerID) + ".json").NewReader(ctx)
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
