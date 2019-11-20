package mongo

import (
	"time"

	"github.com/Jleagle/patreon-go/patreon"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
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

func (pw PatreonWebhook) BSON() bson.D {

	return bson.D{
		{"created_at", pw.CreatedAt},
		{"request_body", pw.RequestBody},
		{"event", pw.Event},
		{"user_id", pw.UserID},
		{"user_email", pw.UserEmail},
		{"data_lifetime_support_cents", pw.DataLifetimeSupportCents},
		{"data_patron_status", pw.DataPatronStatus},
		{"data_pledge_amount_cents", pw.DataPledgeAmountCents},
		{"data_pledge_cap_amount_cents", pw.DataPledgeCapAmountCents},
		{"data_pledge_relationship_start", pw.DataPledgeRelationshipStart},
	}
}

func (pw PatreonWebhook) Raw() (raw patreon.Webhook, err error) {

	err = helpers.Unmarshal([]byte(pw.RequestBody), &raw)
	return raw, err
}

func CountPatreonWebhooks(userID int) (count int64, err error) {

	var item = memcache.MemcachePatreonWebhooksCount(userID)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		return CountDocuments(CollectionPatreonWebhooks, bson.D{{"user_id", userID}}, 0)
	})

	return count, err
}

func GetPatreonWebhooks(offset int64, limit int64, sort bson.D, filter bson.D, projection bson.M) (webhooks []PatreonWebhook, err error) {

	cur, ctx, err := Find(CollectionPatreonWebhooks, offset, limit, sort, filter, projection, nil)
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
