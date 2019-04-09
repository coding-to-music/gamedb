package mongo

import (
	"time"
)

type PatreonWebhook struct {
	CreatedAt   time.Time `bson:"created_at"`
	RequestBody string    `bson:"request_body"`
	Event       string    `bson:"event"`
}

func (pw PatreonWebhook) BSON() (ret interface{}) {

	return M{
		"created_at":   pw.CreatedAt,
		"request_body": pw.RequestBody,
		"event":        pw.Event,
	}
}
