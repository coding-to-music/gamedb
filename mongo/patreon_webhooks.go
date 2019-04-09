package mongo

import (
	"encoding/json"
	"time"

	"github.com/gamedb/website/log"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PatreonWebhook struct {
	CreatedAt               time.Time `bson:"created_at"`
	RequestBody             string    `bson:"request_body"`
	Event                   string    `bson:"event"`
	Email                   string    `bson:"email"`
	PatronStatus            string    `bson:"patron_status"`
	LifetimeSupportCents    int       `bson:"lifetime_support_cents"`
	PledgeAmountCents       int       `bson:"pledge_amount_cents"`
	PledgeCapAmountCents    int       `bson:"pledge_cap_amount_cents"`
	PledgeRelationshipStart time.Time `bson:"pledge_relationship_start"`
}

func (pw PatreonWebhook) BSON() (ret interface{}) {

	return M{
		"created_at":                pw.CreatedAt,
		"request_body":              pw.RequestBody,
		"event":                     pw.Event,
		"email":                     pw.Email,
		"lifetime_support_cents":    pw.LifetimeSupportCents,
		"patron_status":             pw.PatronStatus,
		"pledge_amount_cents":       pw.PledgeAmountCents,
		"pledge_cap_amount_cents":   pw.PledgeCapAmountCents,
		"pledge_relationship_start": pw.PledgeRelationshipStart,
	}
}

func (pw PatreonWebhook) Raw() (raw PatreonWebhookRaw, err error) {

	err = json.Unmarshal([]byte(pw.RequestBody), &raw)
	return raw, err
}

func GetPatreonWebhooks(offset int64, limit int64, sort bool, filter interface{}, projection M) (webhooks []PatreonWebhook, err error) {

	if filter == nil {
		filter = M{}
	}

	client, ctx, err := getMongo()
	if err != nil {
		return webhooks, err
	}

	ops := options.Find()
	if offset > 0 {
		ops.SetSkip(offset)
	}
	if limit > 0 {
		ops.SetLimit(limit)
	}
	if sort {
		ops.SetSort(M{"created_at": 1})
	} else {
		ops.SetSort(M{"created_at": -1})
	}

	if projection != nil {
		ops.SetProjection(projection)
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionPatreonWebhooks.String())
	cur, err := c.Find(ctx, filter, ops)
	if err != nil {
		return webhooks, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var webhook PatreonWebhook
		err := cur.Decode(&webhook)
		if err != nil {
			log.Err(err)
		}
		webhooks = append(webhooks, webhook)
	}

	return webhooks, cur.Err()
}

type PatreonWebhookRawType struct {
	Type string `json:"type"`
}

type PatreonWebhookRaw struct {
	Data     PatreonWebhookRawData             `json:"data"`
	Included []json.RawMessage                 `json:"included"`
	Links    map[string]string                 `json:"links"`
	Campaign PatreonWebhookRawIncludedCampaign ``
	User     PatreonWebhookRawIncludedUser     ``
	Goals    []PatreonWebhookRawIncludedGoals  ``
	Rewards  []PatreonWebhookRawIncludedReward ``
}

type PatreonWebhookRawData struct {
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
}

type PatreonWebhookRawIncludedGoals struct {
	Attributes struct {
		AmountCents         int         `json:"amount_cents"`
		CompletedPercentage int         `json:"completed_percentage"`
		CreatedAt           time.Time   `json:"created_at"`
		Description         string      `json:"description"`
		ReachedAt           interface{} `json:"reached_at"`
		Title               string      `json:"title"`
	} `json:"attributes"`
	ID   string `json:"id"`
	Type string `json:"type"`
}

type PatreonWebhookRawIncludedReward struct {
	Attributes struct {
		Amount           int         `json:"amount"`
		AmountCents      int         `json:"amount_cents"`
		CreatedAt        time.Time   `json:"created_at"`
		Description      string      `json:"description"`
		DiscordRoleIds   []string    `json:"discord_role_ids"`
		EditedAt         time.Time   `json:"edited_at"`
		ImageURL         interface{} `json:"image_url"`
		PatronCount      int         `json:"patron_count"`
		PostCount        int         `json:"post_count"`
		Published        bool        `json:"published"`
		PublishedAt      time.Time   `json:"published_at"`
		Remaining        interface{} `json:"remaining"`
		RequiresShipping bool        `json:"requires_shipping"`
		Title            string      `json:"title"`
		UnpublishedAt    interface{} `json:"unpublished_at"`
		URL              string      `json:"url"`
		UserLimit        interface{} `json:"user_limit"`
	} `json:"attributes"`
	ID            string `json:"id"`
	Relationships struct {
		Campaign struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
			Links struct {
				Related string `json:"related"`
			} `json:"links"`
		} `json:"campaign"`
	} `json:"relationships"`
	Type string `json:"type"`
}

type PatreonWebhookRawIncludedUser struct {
	Attributes struct {
		About              string      `json:"about"`
		CanSeeNsfw         bool        `json:"can_see_nsfw"`
		Created            time.Time   `json:"created"`
		DefaultCountryCode interface{} `json:"default_country_code"`
		DiscordID          string      `json:"discord_id"`
		Email              string      `json:"email"`
		Facebook           interface{} `json:"facebook"`
		FacebookID         interface{} `json:"facebook_id"`
		FirstName          string      `json:"first_name"`
		FullName           string      `json:"full_name"`
		Gender             int         `json:"gender"`
		HasPassword        bool        `json:"has_password"`
		ImageURL           string      `json:"image_url"`
		IsDeleted          bool        `json:"is_deleted"`
		IsEmailVerified    bool        `json:"is_email_verified"`
		IsNuked            bool        `json:"is_nuked"`
		IsSuspended        bool        `json:"is_suspended"`
		LastName           string      `json:"last_name"`
		SocialConnections  struct {
			Deviantart interface{} `json:"deviantart"`
			Discord    struct {
				Scopes []string    `json:"scopes"`
				URL    interface{} `json:"url"`
				UserID string      `json:"user_id"`
			} `json:"discord"`
			Facebook  interface{} `json:"facebook"`
			Instagram struct {
				Scopes []string `json:"scopes"`
				URL    string   `json:"url"`
				UserID string   `json:"user_id"`
			} `json:"instagram"`
			Reddit struct {
				Scopes []string `json:"scopes"`
				URL    string   `json:"url"`
				UserID string   `json:"user_id"`
			} `json:"reddit"`
			Spotify interface{} `json:"spotify"`
			Twitch  interface{} `json:"twitch"`
			Twitter struct {
				URL    string `json:"url"`
				UserID string `json:"user_id"`
			} `json:"twitter"`
			Youtube interface{} `json:"youtube"`
		} `json:"social_connections"`
		ThumbURL string      `json:"thumb_url"`
		Twitch   interface{} `json:"twitch"`
		Twitter  interface{} `json:"twitter"`
		URL      string      `json:"url"`
		Vanity   string      `json:"vanity"`
		Youtube  interface{} `json:"youtube"`
	} `json:"attributes"`
	ID            string `json:"id"`
	Relationships struct {
		Campaign struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
			Links struct {
				Related string `json:"related"`
			} `json:"links"`
		} `json:"campaign"`
	} `json:"relationships"`
	Type string `json:"type"`
}

type PatreonWebhookRawIncludedCampaign struct {
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
	} `json:"relationships"`
	Type string `json:"type"`
}
