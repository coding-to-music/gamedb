package mongo

import (
	"encoding/json"
	"time"

	"github.com/Jleagle/patreon-go/patreon"
	"github.com/gamedb/website/pkg/log"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PatreonWebhook struct {
	CreatedAt               time.Time `bson:"created_at"`
	RequestBody             string    `bson:"request_body"`
	Event                   string    `bson:"event"`
	Email                   string    `bson:"email"`
	PatronStatus            string    `bson:"patron_status"`
	LifetimeSupportCents    int       `bson:"lifetime_support_cents"`
	PledgeAmountCents       int       `bson:"pledge_amount_cents"`
	PledgeCapAmountCents    int       `bson:"pledge_cap_amount_cents"`
	PledgeRelationshipStart time.Time `bson:"pledge_relationship_start"`
}

func (pw PatreonWebhook) BSON() (ret interface{}) {

	return M{
		"created_at":                pw.CreatedAt,
		"request_body":              pw.RequestBody,
		"event":                     pw.Event,
		"email":                     pw.Email,
		"lifetime_support_cents":    pw.LifetimeSupportCents,
		"patron_status":             pw.PatronStatus,
		"pledge_amount_cents":       pw.PledgeAmountCents,
		"pledge_cap_amount_cents":   pw.PledgeCapAmountCents,
		"pledge_relationship_start": pw.PledgeRelationshipStart,
	}
}

func (pw PatreonWebhook) Raw() (raw patreon.Webhook, err error) {

	err = json.Unmarshal([]byte(pw.RequestBody), &raw)
	return raw, err
}

func GetPatreonWebhooks(offset int64, limit int64, sort bool, filter interface{}, projection M) (webhooks []PatreonWebhook, err error) {

	if filter == nil {
		filter = M{}
	}

	client, ctx, err := getMongo()
	if err != nil {
		return webhooks, err
	}

	ops := options.Find()
	if offset > 0 {
		ops.SetSkip(offset)
	}
	if limit > 0 {
		ops.SetLimit(limit)
	}
	if sort {
		ops.SetSort(M{"created_at": 1})
	} else {
		ops.SetSort(M{"created_at": -1})
	}

	if projection != nil {
		ops.SetProjection(projection)
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionPatreonWebhooks.String())
	cur, err := c.Find(ctx, filter, ops)
	if err != nil {
		return webhooks, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var webhook PatreonWebhook
		err := cur.Decode(&webhook)
		if err != nil {
			log.Err(err)
		}
		webhooks = append(webhooks, webhook)
	}

	return webhooks, cur.Err()
}
