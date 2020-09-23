package mongo

import (
	"net/http"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/oauth"
	"go.mongodb.org/mongo-driver/bson"
)

type EventEnum string

var (
	EventSignup         EventEnum = "signup"
	EventLogin          EventEnum = "login"
	EventForgotPassword EventEnum = "forgot-password"
	EventLogout         EventEnum = "logout"
	EventPatreonWebhook EventEnum = "patreon-webhook"
	EventRefresh        EventEnum = "refresh"
	EventLink                     = func(provider oauth.ProviderEnum) EventEnum { return EventEnum("link-" + provider) }
	EventUnlink                   = func(provider oauth.ProviderEnum) EventEnum { return EventEnum("unlink-" + provider) }
)

type Event struct {
	CreatedAt time.Time `bson:"created_at"`
	Type      EventEnum `bson:"type"`
	UserID    int       `bson:"user_id"`
	UserAgent string    `bson:"user_agent"`
	IP        string    `bson:"ip"`
}

func (event Event) BSON() bson.D {

	return bson.D{
		{"created_at", event.CreatedAt},
		{"type", event.Type},
		{"user_id", event.UserID},
		{"user_agent", event.UserAgent},
		{"ip", event.IP},
	}
}

func (event Event) GetCreatedNice() (t string) {
	return event.CreatedAt.Format(helpers.DateTime)
}

func (event Event) GetType() string {

	switch EventEnum(event.Type) {
	case EventLogin:
		return "User Login"
	case EventLogout:
		return "User Logout"
	case EventRefresh:
		return "Profile Update"
	default:
		return strings.Title(string(event.Type))
	}
}

func (event Event) GetIcon() string {

	switch EventEnum(event.Type) {
	case EventLogin:
		return "fa-sign-in-alt"
	case EventLogout:
		return "fa-sign-out-alt"
	case EventRefresh:
		return "fa-sync-alt"
	default:
		return "fa-star"
	}
}

func GetEvents(userID int, offset int64) (events []Event, err error) {

	var sort = bson.D{{"created_at", -1}}
	var filter = bson.D{{"user_id", userID}}

	cur, ctx, err := Find(CollectionEvents, offset, 100, sort, filter, nil, nil)
	if err != nil {
		return events, err
	}

	defer close(cur, ctx)

	for cur.Next(ctx) {

		var event Event
		err := cur.Decode(&event)
		if err != nil {
			log.ErrS(err)
		} else {
			events = append(events, event)
		}
	}

	return events, cur.Err()
}

func NewEvent(r *http.Request, userID int, eventType EventEnum) (err error) {

	event := Event{}
	event.CreatedAt = time.Now()
	event.UserID = userID
	event.Type = eventType

	if r != nil {
		event.UserAgent = r.Header.Get("User-Agent")
		event.IP = r.RemoteAddr
	}

	_, err = InsertOne(CollectionEvents, event)
	return err
}
