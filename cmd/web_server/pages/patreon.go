package pages

import (
	"net/http"
	"time"

	"github.com/Jleagle/patreon-go/patreon"
	"github.com/gamedb/website/pkg"
	"github.com/go-chi/chi"
)

func patreonRouter() http.Handler {
	r := chi.NewRouter()
	r.Post("/webhooks", patreonWebhookPostHandler)
	return r
}

func patreonWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, 0)

	b, event, err := patreon.ValidateRequest(r, config.Config.PatreonSecret.Get())
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

	_, err = pkg.InsertDocument(pkg.CollectionPatreonWebhooks, pkg.PatreonWebhook{
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
		player := pkg.Player{}
		err = pkg.FindDocument(pkg.CollectionPlayers, "email", pwr.User.Attributes.Email, pkg.M{"_id": 1}, &player)
		if err == pkg.ErrNoDocuments {
			return nil
		}
		if err != nil {
			return err
		}

		return pkg.CreateEvent(r, player.ID, pkg.EventPatreon+"-"+event)
	}

	return nil
}
