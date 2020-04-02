package memcache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/gamedb/gamedb/pkg/helpers"
	pubsubHelpers "github.com/gamedb/gamedb/pkg/helpers/pubsub"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

var ErrInQueue = errors.New("already in queue")

func ListenToPubSubMemcache() {

	err := pubsubHelpers.PubSubSubscribe(pubsubHelpers.PubSubMemcache, func(m *pubsub.Message) {

		var ids []string

		err := json.Unmarshal(m.Data, &ids)
		log.Err(err)

		for _, id := range ids {
			err = Delete(id)
			err = helpers.IgnoreErrors(err, ErrNotFound)
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
func ProjectionToString(m bson.M) string {

	if len(m) == 0 {
		return "*"
	}

	var cols []string
	for k := range m {
		cols = append(cols, k)
	}

	sort.Slice(cols, func(i, j int) bool {
		return cols[i] < cols[j]
	})

	return strings.Join(cols, "-")
}

func FilterToString(d bson.D) string {

	if d == nil || len(d) == 0 {
		return "[]"
	}

	b, err := json.Marshal(d)
	if err != nil {
		log.Err(err)
		return "[]"
	}

	h := md5.Sum(b)

	return hex.EncodeToString(h[:])
}
