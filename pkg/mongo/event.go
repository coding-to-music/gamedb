package mongo

import (
	"net/http"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type EventEnum string

const (
	EventSignup         EventEnum = "signup"
	EventLogin          EventEnum = "login"
	EventForgotPassword EventEnum = "forgot-password"
	EventLogout         EventEnum = "logout"
	EventPatreonWebhook EventEnum = "patreon-webhook"
	EventRefresh        EventEnum = "refresh"

	// Connections
	EventLinkSteam     EventEnum = "link-steam"
	EventUnlinkSteam   EventEnum = "unlink-steam"
	EventLinkPatreon   EventEnum = "link-patreon"
	EventUnlinkPatreon EventEnum = "unlink-patreon"
	EventLinkGoogle    EventEnum = "link-google"
	EventUnlinkGoogle  EventEnum = "unlink-google"
	EventLinkDiscord   EventEnum = "link-discord"
	EventUnlinkDiscord EventEnum = "unlink-discord"
	EventLinkGitHub    EventEnum = "link-github"
	EventUnlinkGitHub  EventEnum = "unlink-github"
)

type Event struct {
	CreatedAt time.Time `bson:"created_at"`
	Type      string    `bson:"type"`
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

// Data array for datatables
func (event Event) OutputForJSON(ip string) (output []interface{}) {

	return []interface{}{
		event.CreatedAt.Unix(),
		event.GetCreatedNice(),
		event.GetType(),
		event.GetIP(""),
		event.UserAgent,
		event.GetUserAgentShort(),
		event.GetIP(ip),
		event.GetIcon(),
	}
}

func (event Event) GetCreatedNice() (t string) {
	return event.CreatedAt.Format(helpers.DateTime)
}

func (event Event) GetUserAgentShort() (t string) {

	if len(event.UserAgent) > 50 {
		return event.UserAgent[0:50] + "&hellip;"
	}

	return event.UserAgent
}

// Defaults to IP on struct
func (event Event) GetIP(ip string) string {

	if ip == "" {
		ip = event.IP
	}

	var ips = strings.Split(ip, ", ")
	if len(ips) > 0 && ips[0] != "" {
		return ips[0]
	}
	return "-"
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
		return strings.Title(event.Type)
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

	defer func() {
		err = cur.Close(ctx)
		zap.S().Error(err)
	}()

	for cur.Next(ctx) {

		var event Event
		err := cur.Decode(&event)
		if err != nil {
			zap.S().Error(err)
		} else {
			events = append(events, event)
		}
	}

	return events, cur.Err()
}

func CreateUserEvent(r *http.Request, userID int, eventType EventEnum) (err error) {

	event := Event{}
	event.CreatedAt = time.Now()
	event.UserID = userID
	event.Type = string(eventType)

	if r != nil {
		event.UserAgent = r.Header.Get("User-Agent")
		event.IP = r.RemoteAddr
	}

	_, err = InsertOne(CollectionEvents, event)
	return err
}
