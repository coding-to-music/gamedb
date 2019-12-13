package pubsub

import (
	"context"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
)

// Topics
type PubSubTopic string

const (
	PubSubTopicWebsockets PubSubTopic = "gamedb-websockets"
	PubSubTopicMemcache   PubSubTopic = "gamedb-memcache"
)

// Subscriptions
type PubSubSubscription string

var (
	PubSubWebsockets = PubSubSubscription("gamedb-websockets-" + config.Config.Environment.Get())
	PubSubMemcache   = PubSubSubscription("gamedb-memcache-" + config.Config.Environment.Get())
)

//
var pubSubClient *pubsub.Client
var pubSubClientLock sync.Mutex

func GetPubSub() (client *pubsub.Client, ctx context.Context, err error) {

	pubSubClientLock.Lock()
	defer pubSubClientLock.Unlock()

	ctx = context.Background()

	if pubSubClient == nil {
		log.Info("Connecting to PubSub")
		pubSubClient, err = pubsub.NewClient(ctx, config.Config.GoogleProject.Get())
	}

	return pubSubClient, ctx, err
}
