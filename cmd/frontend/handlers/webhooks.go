package handlers

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Jleagle/patreon-go/patreon"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/oauth"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"golang.org/x/text/currency"
)

func WebhooksRouter() http.Handler {

	r := chi.NewRouter()
	r.Post("/github", gitHubWebhookPostHandler)
	r.Post("/mailjet", mailjetWebhookPostHandler)
	r.Post("/patreon", patreonWebhookPostHandler)
	r.Post("/sendgrid", sendgridWebhookPostHandler)
	r.Post("/twitter", twitterZapierWebhookPostHandler)

	return r
}

func mailjetWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	// Get body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer helpers.Close(r.Body)

	// Save webhook
	err = mongo.SaveWebhook(mongo.WebhookServiceMailjet, "", string(body))
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return
	_, err = w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.ErrS(err)
	}
}

func sendgridWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	// Validate
	if config.C.SendGridSecret == "" {
		log.Err("Missing sendgrid environment variables")
	}

	if r.Header.Get("X-Twilio-Email-Event-Webhook-Signature") != config.C.SendGridSecret {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Get body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer helpers.Close(r.Body)

	// Save webhook
	err = mongo.SaveWebhook(mongo.WebhookServiceSendgrid, "", string(body))
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return
	_, err = w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.ErrS(err)
	}
}

func twitterZapierWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	// Validate secret
	if config.C.TwitterWebhookSecret == "" {
		log.Err("Missing Twitter environment variables")
	}

	var secret = r.URL.Query().Get("secret")
	if secret != config.C.TwitterWebhookSecret {
		log.Err("invalid secret", zap.String("secret", secret))
		http.Error(w, "invalid secret", http.StatusBadRequest)
		return
	}

	// Get body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer helpers.Close(r.Body)

	// Save webhook
	err = mongo.SaveWebhook(mongo.WebhookServiceTwitter, "", string(body))
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle
	webhooks := twitterWebhook{}
	err = json.Unmarshal(body, &webhooks)

	// Delete cache
	err = memcache.Delete(memcache.ItemHomeTweets.Key)
	if err != nil {
		log.Err(err.Error())
	}

	// Forward to Discord
	if webhooks.Link != "" {

		if config.C.DiscordChatBotToken == "" {
			log.Err("Missing discord environment variable")
		}

		discordSession, err := discordgo.New("Bot " + config.C.DiscordChatBotToken)
		if err != nil {
			log.ErrS(err)
			return
		}

		_, err = discordSession.ChannelMessageSend(announcementsChannelID, webhooks.Link)
		if err != nil {
			log.Err(err.Error())
		}

		err = discordSession.Close()
		if err != nil {
			log.ErrS(err)
		}
	}

	// Return
	_, err = w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.ErrS(err)
	}
}

type twitterWebhook struct {
	Text      string `json:"text"`
	Link      string `json:"link"`
	CreatedAt string `json:"created_at"`
}

type patreonTier int

func (t patreonTier) toUserLevel() mysql.UserLevel {
	switch t {
	case patreonTier1:
		return mysql.UserLevel1
	case patreonTier2:
		return mysql.UserLevel2
	case patreonTier3:
		return mysql.UserLevel3
	}
	return mysql.UserLevelFree
}

const (
	patreonTier1 patreonTier = 2431311
	patreonTier2 patreonTier = 2431320
	patreonTier3 patreonTier = 2431347
)

func patreonWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	// Validate
	if config.C.PatreonSecret == "" {
		log.Err("Missing patreon environment variable")
	}

	b, event, err := patreon.Validate(r, config.C.PatreonSecret)
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save webhook
	err = mongo.SaveWebhook(mongo.WebhookServicePatreon, event, string(b))
	if err != nil {
		log.ErrS(err)
	}

	// Handle
	pwr, err := patreon.Unmarshal(b)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(b)))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Find user
	user, err := mysql.GetUserByProviderID(oauth.ProviderPatreon, strconv.Itoa(int(pwr.User.ID)))
	if err == mysql.ErrRecordNotFound {

		user, err = mysql.GetUserByEmail(pwr.User.Attributes.Email)
		if err == mysql.ErrRecordNotFound {
			user = mysql.User{} // Continue
		} else if err != nil {
			log.ErrS(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	} else if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user.Email == "" {
		user.Email = pwr.User.Attributes.Email
	}

	// Update donation bits
	amount := pwr.Data.Attributes.LifetimeSupportCents - user.DonatedPatreon
	if amount > 0 {

		// Get player ID
		var playerID int64

		steam, err := mysql.GetUserProviderByUserID(oauth.ProviderSteam, user.ID)
		if err != nil && err != mysql.ErrRecordNotFound {
			log.ErrS(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if err == nil {
			playerID, err = strconv.ParseInt(steam.ID, 10, 64)
			if err != nil {
				log.ErrS(err)
			}
		}

		// Save donation
		db, err := mysql.GetMySQLClient()
		if err != nil {
			log.ErrS(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		donation := mysql.Donation{
			UserID:           user.ID,
			PlayerID:         playerID,
			Email:            user.Email,
			AmountUSD:        amount,
			OriginalCurrency: currency.USD.String(),
			OriginalAmount:   amount,
			Source:           mysql.DonationSourcePatreon,
			Anon:             false, // todo
			PatreonRef:       pwr.Data.ID,
		}

		db = db.Create(&donation)
		if db.Error != nil {
			log.ErrS(db.Error)
			http.Error(w, db.Error.Error(), http.StatusInternalServerError)
			return
		}

		// Update user
		db, err = mysql.GetMySQLClient()
		if err != nil {
			log.ErrS(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update user
		if user.ID > 0 {

			var level = mysql.UserLevelFree

			for _, v := range pwr.Data.Relationships.CurrentlyEntitledTiers.Data {
				switch i := patreonTier(v.ID); i {
				case patreonTier1, patreonTier2, patreonTier3:
					if i.toUserLevel() > level {
						level = i.toUserLevel()
					}
				}
			}

			update := map[string]interface{}{
				"donated_patreon": pwr.Data.Attributes.LifetimeSupportCents,
				"level":           level,
			}

			db = db.Model(&user).Updates(update)
			if db.Error != nil {
				log.ErrS(db.Error)
				http.Error(w, db.Error.Error(), http.StatusInternalServerError)
				return
			}

			// Create event
			err = mongo.NewEvent(r, user.ID, mongo.EventPatreonWebhook+"-"+mongo.EventEnum(event))
			if err != nil {
				log.Err(err.Error(), zap.String("body", string(b)))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	// Return
	_, err = w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.ErrS(err)
	}
}

func gitHubWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	// Get body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer helpers.Close(r.Body)

	// Validate
	var signature = r.Header.Get("X-Hub-Signature")
	var event = r.Header.Get("X-GitHub-Event")

	if len(signature) != 45 || !strings.HasPrefix(signature, "sha1=") {
		http.Error(w, "Invalid signature (1)", 400)
		return
	}

	if config.C.GithubWebhookSecret == "" {
		log.Err("Missing github environment variables")
	}

	mac := hmac.New(sha1.New, []byte(config.C.GithubWebhookSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte("sha1="+expectedMAC), []byte(signature)) {
		log.Err("Invalid signature (2)", zap.String("secret", config.C.GithubWebhookSecret))
		http.Error(w, "Invalid signature (2)", 400)
		return
	}

	// Save webhook
	err = mongo.SaveWebhook(mongo.WebhookServiceGithub, event, string(body))
	if err != nil {
		log.ErrS(err)
	}

	// Handle
	switch event {
	case "push":

		// Clear cache
		err := memcache.Delete(memcache.ItemCommitsPage(1).Key)
		if err != nil {
			log.ErrS(err)
		}
	}

	// Return
	_, err = w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.ErrS(err)
	}
}
