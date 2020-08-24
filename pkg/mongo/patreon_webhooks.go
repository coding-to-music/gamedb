package mongo

import (
	"time"

	"github.com/Jleagle/patreon-go/patreon"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

type PatreonWebhook struct {
	CreatedAt   time.Time `bson:"created_at"`
	RequestBody string    `bson:"request_body"`
	Event       string    `bson:"event"`
}

func (webhook PatreonWebhook) BSON() bson.D {

	return bson.D{
		{"created_at", webhook.CreatedAt},
		{"request_body", webhook.RequestBody},
		{"event", webhook.Event},
	}
}

func (webhook PatreonWebhook) Unmarshal() (wh patreon.Webhook, err error) {

	err = helpers.Unmarshal([]byte(webhook.RequestBody), &wh)
	return wh, err
}

func GetPatreonWebhooks(offset int64, limit int64, sort bson.D, filter bson.D, projection bson.M) (webhooks []PatreonWebhook, err error) {

	cur, ctx, err := Find(CollectionPatreonWebhooks, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return webhooks, err
	}

	defer func() {
		err = cur.Close(ctx)
		if err != nil {
			log.ErrS(err)
		}
	}()

	for cur.Next(ctx) {

		var webhook PatreonWebhook
		err := cur.Decode(&webhook)
		if err != nil {
			log.ErrS(err)
		} else {
			webhooks = append(webhooks, webhook)
		}
	}

	return webhooks, cur.Err()
}
