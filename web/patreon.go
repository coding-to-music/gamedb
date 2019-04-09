package web

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

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

	event := r.Header.Get("X-Patreon-Event")
	signature := r.Header.Get("X-Patreon-Signature")

	if event == "" || signature == "" {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("missing event or signature header"))
		log.Err(err)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(err.Error()))
		log.Err(err)
		return
	}

	hash := hmac.New(md5.New, []byte(config.Config.PatreonSecret))
	if _, err := hash.Write(b); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(err.Error()))
		log.Err(err)
		return
	}

	sum := hash.Sum(nil)
	expectedSignature := hex.EncodeToString(sum)

	if expectedSignature != signature {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("invalid webhook signature"))
		log.Err(err)
		return
	}

	pwr, err := unmarshalWebhook(b)
	if err != nil {
		log.Err(err)
		return
	}

	err = saveWebhookToMongo(event, b)
	if err != nil {
		log.Err(err)
		return
	}

	err = saveWebhookEvent(r, pwr)
	if err != nil {
		log.Err(err)
		return
	}

	log.Err(err)
}

func unmarshalWebhook(b []byte) (pwr mongo.PatreonWebhookRaw, err error) {

	err = json.Unmarshal(b, &pwr)
	if err != nil {
		return pwr, err
	}

	for _, v := range pwr.Included {

		typ := mongo.PatreonWebhookRawType{}
		err = json.Unmarshal(v, &typ)
		if err != nil {
			return pwr, err
		}

		switch typ.Type {
		case "campaign":

			include := mongo.PatreonWebhookRawIncludedCampaign{}
			err = json.Unmarshal(v, &include)
			if err != nil {
				return pwr, err
			}
			pwr.Campaign = include

		case "user":

			include := mongo.PatreonWebhookRawIncludedUser{}
			err = json.Unmarshal(v, &include)
			if err != nil {
				return pwr, err
			}
			pwr.User = include

		case "reward":

			include := mongo.PatreonWebhookRawIncludedReward{}
			err = json.Unmarshal(v, &include)
			if err != nil {
				return pwr, err
			}
			pwr.Rewards = append(pwr.Rewards, include)

		case "goal":

			include := mongo.PatreonWebhookRawIncludedGoals{}
			err = json.Unmarshal(v, &include)
			if err != nil {
				return pwr, err
			}
			pwr.Goals = append(pwr.Goals, include)

		default:
			log.Warning("Missing webhook data")
		}
	}

	return pwr, nil
}

func saveWebhookToMongo(event string, pwr mongo.PatreonWebhookRaw, body []byte) (err error) {

	_, err = mongo.InsertDocument(mongo.CollectionPatreonWebhooks, mongo.PatreonWebhook{
		CreatedAt:   time.Now(),
		RequestBody: string(body),
		Event:       event,
		Email:       pwr.User.Attributes.Email,
	})
	return err
}

func saveWebhookEvent(r *http.Request, pwr mongo.PatreonWebhookRaw) (err error) {

	if pwr.User.Attributes.Email != "" {
		player := mongo.Player{}
		err = mongo.FindDocument(mongo.CollectionPlayers, "email", pwr.User.Attributes.Email, mongo.M{"_id": 1}, &player)
		if err != nil {
			return err
		}

		return mongo.CreateEvent(r, player.ID, mongo.EventRefresh)
	}

	return nil
}
