package web

import (
	"net/http"
	"time"

	"github.com/Jleagle/patreon-go/patreon"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/mongo"
	"github.com/go-chi/chi"
)

func patreonRouter() http.Handler {
	r := chi.NewRouter()
	r.Post("/webhooks", webhookHandler)
	return r
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {

	b, event, err := patreon.ValidateRequest(r, config.Config.PatreonSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Err(err)
		return
	}

	pwr, err := patreon.UnmarshalBytes(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Err(err)
		return
	}

	err = saveWebhookToMongo(event, pwr, b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Err(err)
		return
	}

	err = saveWebhookEvent(r, event, pwr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Err(err)
		return
	}
}

func saveWebhookToMongo(event string, pwr patreon.Webhook, body []byte) (err error) {

	_, err = mongo.InsertDocument(mongo.CollectionPatreonWebhooks, mongo.PatreonWebhook{
		CreatedAt:               time.Now(),
		RequestBody:             string(body),
		Event:                   event,
		Email:                   pwr.User.Attributes.Email,
		PatronStatus:            pwr.Data.Attributes.PatronStatus,
		LifetimeSupportCents:    pwr.Data.Attributes.LifetimeSupportCents,
		PledgeAmountCents:       pwr.Data.Attributes.PledgeAmountCents,
		PledgeCapAmountCents:    int(pwr.Data.Attributes.PledgeCapAmountCents),
		PledgeRelationshipStart: pwr.Data.Attributes.PledgeRelationshipStart,
	})
	return err
}

func saveWebhookEvent(r *http.Request, event string, pwr patreon.Webhook) (err error) {

	if pwr.User.Attributes.Email != "" {
		player := mongo.Player{}
		err = mongo.FindDocument(mongo.CollectionPlayers, "email", pwr.User.Attributes.Email, mongo.M{"_id": 1}, &player)
		if err == mongo.ErrNoDocuments {
			return nil
		}
		if err != nil {
			return err
		}

		return mongo.CreateEvent(r, player.ID, mongo.EventPatreon+"-"+event)
	}

	return nil
}
