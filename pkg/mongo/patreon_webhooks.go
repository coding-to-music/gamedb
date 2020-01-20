package mongo

import (
	"time"

	"github.com/Jleagle/patreon-go/patreon"
	"github.com/gamedb/gamedb/pkg/helpers"
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

func (webhook PatreonWebhook) BSON() bson.D {

	return bson.D{
		{"created_at", webhook.CreatedAt},
		{"request_body", webhook.RequestBody},
		{"event", webhook.Event},
		{"user_id", webhook.UserID},
		{"user_email", webhook.UserEmail},
		{"data_lifetime_support_cents", webhook.DataLifetimeSupportCents},
		{"data_patron_status", webhook.DataPatronStatus},
		{"data_pledge_amount_cents", webhook.DataPledgeAmountCents},
		{"data_pledge_cap_amount_cents", webhook.DataPledgeCapAmountCents},
		{"data_pledge_relationship_start", webhook.DataPledgeRelationshipStart},
	}
}

func (webhook PatreonWebhook) Raw() (raw patreon.Webhook, err error) {

	err = helpers.Unmarshal([]byte(webhook.RequestBody), &raw)
	return raw, err
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
		} else {
			webhooks = append(webhooks, webhook)
		}
	}

	return webhooks, cur.Err()
}
