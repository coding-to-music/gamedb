package helpers

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/pubsub"
	"github.com/gamedb/gamedb/pkg/config"
)

type PubSubTopic string

const (
	PubSubTopicWebsockets PubSubTopic = "gamedb-websockets"
	PubSubTopicMemcache   PubSubTopic = "gamedb-memcache"
)

type PubSubSubscription string

var (
	PubSubWebsockets = PubSubSubscription("gamedb-websockets-" + config.Config.Environment.Get())
	PubSubMemcache   = PubSubSubscription("gamedb-memcache-" + config.Config.Environment.Get())
)

var pubSubClient *pubsub.Client

func GetPubSub() (client *pubsub.Client, ctx context.Context, err error) {

	ctx = context.Background()

	if pubSubClient == nil {
		client, err = pubsub.NewClient(ctx, config.Config.GoogleProject.Get())
	}

	return client, ctx, err
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
