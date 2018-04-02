// https://github.com/GoogleCloudPlatform/google-cloud-go/blob/master/storage/example_test.go
// https://github.com/GoogleCloudPlatform/golang-samples/blob/master/storage/objects/main.go

package storage

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/Jleagle/go-helpers/logger"
)

const (
	FolderImg460x215 = "img460x215"
)

var (
	bucket = os.Getenv("STEAM_GOOGLE_BUCKET")
)

func UploadHeaderImage(appID int) (path string) {

	appIDstring := strconv.Itoa(appID)

	ctx := context.Background()

	// Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		logger.Error(err)
	}

	// Get the image
	resp, err := http.Get("http://cdn.akamai.steamstatic.com/steam/apps/" + appIDstring + "/header.jpg")
	if resp.StatusCode != http.StatusOK {
		return "/assets/img/no-app-image-banner.jpg"
	}
	if err != nil {
		logger.Error(err)
	}
	defer resp.Body.Close()

	// Save image to bucket
	fileName := "app-img-460x215/" + appIDstring + ".jpg"
	object := client.Bucket(bucket).Object(fileName)

	wc := object.NewWriter(ctx)
	if _, err = io.Copy(wc, resp.Body); err != nil {
		logger.Error(err)
	}
	if err := wc.Close(); err != nil {
		logger.Error(err)
	}

	// Make public
	if err := object.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		logger.Error(err)
	}

	//
	return "/" + fileName
}

func UploadGamesJson(playerID int, gameBytes []byte) string {

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

func DownloadGamesJson(playerID int) (bytes []byte, err error) {

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
