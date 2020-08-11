package youtube

import (
	"context"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	ctx    context.Context
	client *youtube.Service
	lock   sync.Mutex
)

func GetYouTube() (*youtube.Service, context.Context, error) {

	lock.Lock()
	defer lock.Unlock()

	if client == nil {

		var err error
		ctx = context.TODO()
		client, err = youtube.NewService(ctx, option.WithAPIKey(config.Config.YoutubeAPIKey.Get()))
		if err != nil {
			return nil, nil, err
		}
	}

	return client, ctx, nil
}
