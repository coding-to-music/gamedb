package mongo

import (
	"strings"
	"time"

	"github.com/Jleagle/patreon-go/patreon"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

type WebhookService string

func (s WebhookService) ToString() string {
	switch s {
	case WebhookServiceGithub:
		return "GitHub"
	case WebhookServiceSendgrid:
		return "SendGrid"
	default:
		return strings.Title(string(s))
	}
}

const (
	WebhookServicePatreon  WebhookService = "patreon"
	WebhookServiceGithub   WebhookService = "github"
	WebhookServiceTwitter  WebhookService = "twitter"
	WebhookServiceSendgrid WebhookService = "sendgrid"
	WebhookServiceMailjet  WebhookService = "mailjet"
)

type Webhook struct {
	CreatedAt   time.Time      `bson:"created_at"`
	Service     WebhookService `bson:"service"`
	Event       string         `bson:"event"`
	RequestBody string         `bson:"request_body"`
}

func (webhook Webhook) BSON() bson.D {

	return bson.D{
		{"created_at", webhook.CreatedAt},
		{"service", webhook.Service},
		{"event", webhook.Event},
		{"request_body", webhook.RequestBody},
	}
}

func (webhook Webhook) UnmarshalPatreon() (wh patreon.Webhook, err error) {

	err = helpers.Unmarshal([]byte(webhook.RequestBody), &wh)
	return wh, err
}

func (webhook Webhook) GetHash() string {

	hash := helpers.MD5([]byte(webhook.RequestBody))
	if len(hash) > 7 {
		hash = hash[0:7]
	}
	return hash
}

func SaveWebhook(service WebhookService, event string, body string) error {

	row := Webhook{
		CreatedAt:   time.Now(),
		RequestBody: body,
		Event:       event,
		Service:     service,
	}

	_, err := InsertOne(CollectionWebhooks, row)
	return err
}

func GetWebhooks(offset int64, limit int64, sort bson.D, filter bson.D, projection bson.M) (webhooks []Webhook, err error) {

	cur, ctx, err := Find(CollectionWebhooks, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return webhooks, err
	}

	defer closeCursor(cur, ctx)

	for cur.Next(ctx) {

		var webhook Webhook
		err := cur.Decode(&webhook)
		if err != nil {
			log.ErrS(err)
		} else {
			webhooks = append(webhooks, webhook)
		}
	}

	return webhooks, cur.Err()
}
