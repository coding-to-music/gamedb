package pubsub

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/pubsub"
)

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

	return res, nil
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
