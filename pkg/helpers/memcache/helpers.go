package memcache

import (
	"encoding/json"
	"errors"

	"cloud.google.com/go/pubsub"
	"github.com/Jleagle/memcache-go/memcache"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

var ErrInQueue = errors.New("already in queue")

func IsInQueue(item memcache.Item) bool {

	mc := GetClient()

	_, err := mc.Get(item.Key)
	if err == nil {
		return true
	}

	err = mc.Set(&item)
	log.Err(err)

	return false
}

func ListenToPubSubMemcache() {

	mc := GetClient()

	err := helpers.PubSubSubscribe(helpers.PubSubMemcache, func(m *pubsub.Message) {

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

	_, err = helpers.Publish(helpers.PubSubTopicMemcache, keys)
	return err
}
