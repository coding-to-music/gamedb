package helpers

import (
	"strconv"
	"strings"

	"github.com/Jleagle/google-cloud-storage-go/gcs"
	"github.com/gamedb/website/pkg"
)

var (
	PathBadges      = func(playerID int64) string { return "player-badges/" + strconv.FormatInt(playerID, 10) + ".json" }
	PathFriends     = func(playerID int64) string { return "player-friends/" + strconv.FormatInt(playerID, 10) + ".json" }
	PathRecentGames = func(playerID int64) string { return "player-recent-games/" + strconv.FormatInt(playerID, 10) + ".json" }
)

func IsStorageLocaion(x string) bool {
	return strings.HasSuffix(x, ".json")
}

func Upload(path string, data []byte) (err error) {

	payload := gcs.UploadPayload{}
	payload.Bucket = config.Config.GoogleBucket.Get()
	payload.Path = path
	payload.Transformer = gcs.TransformerSnappyEncode
	payload.Data = data
	payload.Public = false

	return gcs.Upload(payload)
}

func Download(path string) (data []byte, err error) {

	payload := gcs.DownloadPayload{}
	payload.Bucket = config.Config.GoogleBucket.Get()
	payload.Path = path
	payload.Transformer = gcs.TransformerSnappyDecode

	return gcs.Download(payload)
}
