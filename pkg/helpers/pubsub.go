package helpers

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/pubsub"
	"github.com/gamedb/gamedb/pkg/config"
)

type PubSubItem string

const (
	PubSubWebsockets PubSubItem = "gamedb-websockets"
)

var pubSubClient *pubsub.Client

func GetPubSub() (client *pubsub.Client, ctx context.Context, err error) {

	ctx = context.Background()

	if pubSubClient == nil {
		client, err = pubsub.NewClient(ctx, config.Config.GoogleProject.Get())
	}

	return client, ctx, err
}

func Publish(topic PubSubItem, message interface{}) (res *pubsub.PublishResult, err error) {

	b, err := json.Marshal(message)
	if err != nil {
		return res, err
	}

	client, ctx, err := GetPubSub()
	if err != nil {
		return res, err
	}

	t := client.Topic(string(topic))
	res = t.Publish(ctx, &pubsub.Message{Data: b})

	return res, err
}

func Subscribe(topic PubSubItem, callback func(m *pubsub.Message)) (err error) {

	client, ctx, err := GetPubSub()
	if err != nil {
		return err
	}

	sub := client.Subscription(string(topic))
	return sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		callback(m)
		m.Ack()
	})
}
