package helpers

import (
	"context"
	"encoding/json"
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

func Publish(topic PubSubTopic, message interface{}) (res *pubsub.PublishResult, err error) {

	client, ctx, err := GetPubSub()
	if err != nil {
		return res, err
	}

	b, err := json.Marshal(message)
	if err != nil {
		return res, err
	}

	t := client.Topic(string(topic))
	res = t.Publish(ctx, &pubsub.Message{Data: b})

	return res, err
}

func PubSubSubscribe(subscription PubSubSubscription, callback func(m *pubsub.Message)) (err error) {

	client, ctx, err := GetPubSub()
	if err != nil {
		return err
	}

	sub := client.Subscription(string(subscription))
	return sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		callback(m)
		m.Ack()
	})
}
