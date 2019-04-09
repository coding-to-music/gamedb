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
		log.Err("invalid webhook headers")
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Err(err)
		return
	}

	valid, err := verifyPatreonWebhook(b, config.Config.PatreonSecret, signature)
	if err != nil {
		log.Err(err)
		return
	}

	if !valid {
		log.Err("invalid webhook signature")
		return
	}

	wh := PatreonWebhook{}
	err = json.Unmarshal(b, &wh)
	if err != nil {
		log.Err(err)
		return
	}

	_, err = mongo.InsertDocument(mongo.CollectionPatreonWebhooks, mongo.PatreonWebhook{
		CreatedAt:   time.Now(),
		RequestBody: string(b),
		Event:       event,
	})

	log.Err(err)
}

func verifyPatreonWebhook(message []byte, secret string, signature string) (bool, error) {

	hash := hmac.New(md5.New, []byte(secret))
	if _, err := hash.Write(message); err != nil {
		return false, err
	}

	sum := hash.Sum(nil)
	expectedSignature := hex.EncodeToString(sum)

	return expectedSignature == signature, nil
}

type PatreonWebhook struct {
	Data struct {
		Attributes struct {
			FullName                string      `json:"full_name"`
			IsFollower              bool        `json:"is_follower"`
			LastChargeDate          time.Time   `json:"last_charge_date"`
			LastChargeStatus        string      `json:"last_charge_status"`
			LifetimeSupportCents    int         `json:"lifetime_support_cents"`
			PatronStatus            string      `json:"patron_status"`
			PledgeAmountCents       int         `json:"pledge_amount_cents"`
			PledgeCapAmountCents    interface{} `json:"pledge_cap_amount_cents"`
			PledgeRelationshipStart time.Time   `json:"pledge_relationship_start"`
		} `json:"attributes"`
		ID            interface{} `json:"id"`
		Relationships struct {
			Address struct {
				Data interface{} `json:"data"`
			} `json:"address"`
			Campaign struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
				Links struct {
					Related string `json:"related"`
				} `json:"links"`
			} `json:"campaign"`
			User struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
				Links struct {
					Related string `json:"related"`
				} `json:"links"`
			} `json:"user"`
		} `json:"relationships"`
		Type string `json:"type"`
	} `json:"data"`
	Included []struct {
		Attributes struct {
			AvatarPhotoURL                string      `json:"avatar_photo_url"`
			CoverPhotoURL                 string      `json:"cover_photo_url"`
			CreatedAt                     time.Time   `json:"created_at"`
			CreationCount                 int         `json:"creation_count"`
			CreationName                  string      `json:"creation_name"`
			DiscordServerID               string      `json:"discord_server_id"`
			DisplayPatronGoals            bool        `json:"display_patron_goals"`
			EarningsVisibility            string      `json:"earnings_visibility"`
			ImageSmallURL                 string      `json:"image_small_url"`
			ImageURL                      string      `json:"image_url"`
			IsChargeUpfront               bool        `json:"is_charge_upfront"`
			IsChargedImmediately          bool        `json:"is_charged_immediately"`
			IsMonthly                     bool        `json:"is_monthly"`
			IsNsfw                        bool        `json:"is_nsfw"`
			IsPlural                      bool        `json:"is_plural"`
			MainVideoEmbed                interface{} `json:"main_video_embed"`
			MainVideoURL                  interface{} `json:"main_video_url"`
			Name                          string      `json:"name"`
			OneLiner                      interface{} `json:"one_liner"`
			OutstandingPaymentAmountCents int         `json:"outstanding_payment_amount_cents"`
			PatronCount                   int         `json:"patron_count"`
			PayPerName                    string      `json:"pay_per_name"`
			PledgeSum                     int         `json:"pledge_sum"`
			PledgeURL                     string      `json:"pledge_url"`
			PublishedAt                   time.Time   `json:"published_at"`
			Summary                       string      `json:"summary"`
			ThanksEmbed                   interface{} `json:"thanks_embed"`
			ThanksMsg                     interface{} `json:"thanks_msg"`
			ThanksVideoURL                interface{} `json:"thanks_video_url"`
			URL                           string      `json:"url"`
		} `json:"attributes"`
		ID            string `json:"id"`
		Relationships struct {
			Creator struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
				Links struct {
					Related string `json:"related"`
				} `json:"links"`
			} `json:"creator"`
			Goals struct {
				Data []struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
			} `json:"goals"`
			Rewards struct {
				Data []struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
			} `json:"rewards"`
		} `json:"relationships,omitempty"`
		Type string `json:"type"`
	} `json:"included"`
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
}
