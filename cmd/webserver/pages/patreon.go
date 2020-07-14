package pages

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Jleagle/patreon-go/patreon"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func WebhooksRouter() http.Handler {

	r := chi.NewRouter()
	r.Post("/patreon", patreonWebhookPostHandler)
	r.Post("/github", gitHubWebhookPostHandler)
	return r
}

func patreonWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	b, event, err := patreon.ValidateRequest(r, config.Config.PatreonSecret.Get())
	if err != nil {
		log.Err(err, r)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	pwr, err := patreon.UnmarshalBytes(b)
	if err != nil {
		log.Err(err, r)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = saveWebhookToMongo(event, pwr, b)
	if err != nil {
		log.Err(err, r)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = saveWebhookEvent(r, mongo.EventEnum(event), pwr)
	if err != nil {
		log.Err(err, r)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func saveWebhookToMongo(event string, pwr patreon.Webhook, body []byte) (err error) {

	userId, err := strconv.Atoi(pwr.User.ID)
	log.Err(err)

	_, err = mongo.InsertOne(mongo.CollectionPatreonWebhooks, mongo.PatreonWebhook{
		CreatedAt:                   time.Now(),
		RequestBody:                 string(body),
		Event:                       event,
		UserID:                      userId,
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

	email := pwr.User.Attributes.Email
	if email == "" {
		return nil
	}

	player := mongo.Player{}
	err = mongo.FindOne(mongo.CollectionPlayers, bson.D{{Key: "email", Value: email}}, nil, bson.M{"_id": 1}, &player)
	if err == mongo.ErrNoDocuments {
		return nil
	}
	if err != nil {
		return err
	}

	user, err := mysql.GetUserByKey("steam_id", player.ID, 0)
	if err == mysql.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		return err
	}

	return mongo.CreateUserEvent(r, user.ID, mongo.EventPatreonWebhook+"-"+event)
}

func gitHubWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

}
