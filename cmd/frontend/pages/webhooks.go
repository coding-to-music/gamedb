package pages

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Jleagle/patreon-go/patreon"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
	"github.com/nlopes/slack"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

func WebhooksRouter() http.Handler {

	r := chi.NewRouter()
	r.Post("/patreon", patreonWebhookPostHandler)
	r.Post("/github", gitHubWebhookPostHandler)
	return r
}

func patreonWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	b, event, err := patreon.Validate(r, config.C.PatreonSecret)
	if err != nil {
		log.ErrS(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = mongo.InsertOne(mongo.CollectionPatreonWebhooks, mongo.PatreonWebhook{
		CreatedAt:   time.Now(),
		RequestBody: string(b),
		Event:       event,
	})
	if err != nil {
		log.ErrS(err)
	}

	pwr, err := patreon.Unmarshal(b)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("webhook", b))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = savePatreonWebhookEvent(r, mongo.EventEnum(event), pwr)
	if err != nil {
		log.ErrS(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Slack message
	err = slack.PostWebhook(config.C.SlackPatreonWebhook, &slack.WebhookMessage{Text: event})
	log.ErrS(err)
}

func savePatreonWebhookEvent(r *http.Request, event mongo.EventEnum, pwr patreon.Webhook) (err error) {

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

const signaturePrefix = "sha1="
const signatureLength = 45

func gitHubWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	// Get body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), 500)
		return
	}

	defer func() {
		err := r.Body.Close()
		if err != nil {
			log.ErrS(err)
		}
	}()

	zap.L().Named(log.LogNameWebhooksGitHub).Info("Incoming GitHub webhook", zap.ByteString("webhook", body))

	//
	var signature = r.Header.Get("X-Hub-Signature")
	var event = r.Header.Get("X-GitHub-Event")

	if len(signature) != signatureLength || !strings.HasPrefix(signature, signaturePrefix) {
		http.Error(w, "Invalid signature (1)", 400)
		return
	}

	mac := hmac.New(sha1.New, []byte(config.C.GithubWebhookSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signaturePrefix+expectedMAC), []byte(signature)) {
		log.Err("Invalid signature (2)", zap.String("secret", config.C.GithubWebhookSecret))
		http.Error(w, "Invalid signature (2)", 400)
		return
	}

	switch event {
	case "push":

		// Clear cache
		items := []string{
			memcache.MemcacheCommitsPage(1).Key,
			memcache.MemcacheCommitsTotal.Key,
		}

		err := memcache.Delete(items...)
		if err != nil {
			log.ErrS(err)
		}
	}

	http.Error(w, "200", 200)
}
