// https://github.com/GoogleCloudPlatform/google-cloud-go/blob/master/storage/example_test.go
// https://github.com/GoogleCloudPlatform/golang-samples/blob/master/storage/objects/main.go

package storage

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/Jleagle/go-helpers/logger"
)

const (
	Img460x215 = "img460x215"
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
	object := client.Bucket("steam-191600.appspot.com").Object(fileName)

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
