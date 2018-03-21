// https://github.com/GoogleCloudPlatform/google-cloud-go/blob/master/storage/example_test.go
// https://github.com/GoogleCloudPlatform/golang-samples/blob/master/storage/objects/main.go

package storage

import (
	"context"
	"os"

	"cloud.google.com/go/storage"
	"github.com/Jleagle/go-helpers/logger"
)

func xx() {
	ctx := context.Background()

	// Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		logger.Error(err)
	}

	// Creates a Bucket instance.
	bucket := client.Bucket("steam-191600.appspot.com")

	// Creates the new bucket.
	if err := bucket.Create(ctx, os.Getenv("STEAM_GOOGLE_PROJECT"), nil); err != nil {
		logger.Error(err)
	}
}
