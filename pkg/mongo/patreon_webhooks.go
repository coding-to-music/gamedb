package mongo

import (
	"time"

	"github.com/Jleagle/patreon-go/patreon"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PatreonWebhook struct {
	CreatedAt                   time.Time `bson:"created_at"`
	RequestBody                 string    `bson:"request_body"`
	Event                       string    `bson:"event"`
	UserID                      int       `json:"user_id"`
	UserEmail                   string    `bson:"user_email"`
	DataPatronStatus            string    `bson:"patron_status"`
	DataLifetimeSupportCents    int       `bson:"lifetime_support_cents"`
	DataPledgeAmountCents       int       `bson:"pledge_amount_cents"`
	DataPledgeCapAmountCents    int       `bson:"pledge_cap_amount_cents"`
	DataPledgeRelationshipStart time.Time `bson:"pledge_relationship_start"`
}

func (pw PatreonWebhook) BSON() (ret interface{}) {

	return M{
		"created_at":                     pw.CreatedAt,
		"request_body":                   pw.RequestBody,
		"event":                          pw.Event,
		"user_id":                        pw.UserID,
		"user_email":                     pw.UserEmail,
		"data_lifetime_support_cents":    pw.DataLifetimeSupportCents,
		"data_patron_status":             pw.DataPatronStatus,
		"data_pledge_amount_cents":       pw.DataPledgeAmountCents,
		"data_pledge_cap_amount_cents":   pw.DataPledgeCapAmountCents,
		"data_pledge_relationship_start": pw.DataPledgeRelationshipStart,
	}
}

func (pw PatreonWebhook) Raw() (raw patreon.Webhook, err error) {

	err = helpers.Unmarshal([]byte(pw.RequestBody), &raw)
	return raw, err
}

func CountPatreonWebhooks(userID int) (count int64, err error) {

	var item = helpers.MemcachePatreonWebhooksCount(userID)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		return CountDocuments(CollectionPatreonWebhooks, M{"user_id": userID})
	})

	return count, err
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
