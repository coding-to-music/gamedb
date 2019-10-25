package pages

import (
	"net/http"
	"time"

	"github.com/Jleagle/patreon-go/patreon"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	. "go.mongodb.org/mongo-driver/bson"
)

func PatreonRouter() http.Handler {

	r := chi.NewRouter()
	r.Post("/webhooks", patreonWebhookPostHandler)
	return r
}

func patreonWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	b, event, err := patreon.ValidateRequest(r, config.Config.PatreonSecret.Get())
	if err != nil {
		log.Err(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	pwr, err := patreon.UnmarshalBytes(b)
	if err != nil {
		log.Err(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = saveWebhookToMongo(event, pwr, b)
	if err != nil {
		log.Err(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = saveWebhookEvent(r, mongo.EventEnum(event), pwr)
	if err != nil {
		log.Err(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func saveWebhookToMongo(event string, pwr patreon.Webhook, body []byte) (err error) {

	_, err = mongo.InsertOne(mongo.CollectionPatreonWebhooks, mongo.PatreonWebhook{
		CreatedAt:                   time.Now(),
		RequestBody:                 string(body),
		Event:                       event,
		UserID:                      int(pwr.User.ID),
		UserEmail:                   pwr.User.Attributes.Email,
		DataPatronStatus:            pwr.Data.Attributes.PatronStatus,
		DataLifetimeSupportCents:    pwr.Data.Attributes.LifetimeSupportCents,
		DataPledgeAmountCents:       pwr.Data.Attributes.PledgeAmountCents,
		DataPledgeCapAmountCents:    int(pwr.Data.Attributes.PledgeCapAmountCents),
		DataPledgeRelationshipStart: pwr.Data.Attributes.PledgeRelationshipStart,
	})
	return err
}

func saveWebhookEvent(r *http.Request, event mongo.EventEnum, pwr patreon.Webhook) (err error) {

	if pwr.User.Attributes.Email != "" {
		player := mongo.Player{}
		err = mongo.FindOne(mongo.CollectionPlayers, D{{"email", pwr.User.Attributes.Email}}, nil, M{"_id": 1}, &player)
		if err == mongo.ErrNoDocuments || (err == nil && player.ID == 0) {
			return nil
		}
		if err != nil {
			return err
		}

		return mongo.CreatePlayerEvent(r, player.ID, mongo.EventPatreonWebhook+"-"+event)
	}

	return nil
}
