package memcache

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/Jleagle/memcache-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	pubsubHelpers "github.com/gamedb/gamedb/pkg/helpers/pubsub"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

var ErrInQueue = errors.New("already in queue")

func ListenToPubSubMemcache() {

	mc := GetClient()

	err := pubsubHelpers.PubSubSubscribe(pubsubHelpers.PubSubMemcache, func(m *pubsub.Message) {

		var ids []string

		err := json.Unmarshal(m.Data, &ids)
		log.Err(err)

		for _, id := range ids {
			err = mc.Delete(id)
			err = helpers.IgnoreErrors(err, memcache.ErrCacheMiss)
			log.Err(err)
		}
	})
	log.Err(err)
}

//
func RemoveKeyFromMemCacheViaPubSub(keys ...string) (err error) {

	_, err = pubsubHelpers.Publish(pubsubHelpers.PubSubTopicMemcache, keys)
	return err
}

//
func bsonMapToString(b bson.M) string {

	if len(b) == 0 {
		return "*"
	}

	var cols []string
	for k := range b {
		cols = append(cols, k)
	}

	sort.Slice(cols, func(i, j int) bool {
		return cols[i] < cols[j]
	})

	return strings.Join(cols, "-")
}
