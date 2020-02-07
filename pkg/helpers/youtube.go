package helpers

import (
	"context"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	YoutubeContext = context.TODO()
	YoutubeService *youtube.Service
)

func init() {

	var err error

	YoutubeService, err = youtube.NewService(YoutubeContext, option.WithAPIKey(config.Config.YoutubeAPIKey.Get()))
	if err != nil {
		log.Critical(err)
	}
}
