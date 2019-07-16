package mongo

import (
	"net/http"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"go.mongodb.org/mongo-driver/mongo/options"
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
)

type Event struct {
	CreatedAt time.Time `bson:"created_at"`
	Type      string    `bson:"type"`
	UserID    int       `bson:"user_id"`
	UserAgent string    `bson:"user_agent"`
	IP        string    `bson:"ip"`
}

func (event Event) BSON() (ret interface{}) {

	return M{
		"created_at": event.CreatedAt,
		"type":       event.Type,
		"user_id":    event.UserID,
		"user_agent": event.UserAgent,
		"ip":         event.IP,
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

	client, ctx, err := getMongo()
	if err != nil {
		return events, err
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionEvents.String())

	cur, err := c.Find(ctx, M{"user_id": userID}, options.Find().SetLimit(100).SetSkip(offset).SetSort(D{{"created_at", -1}}))
	if err != nil {
		return events, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var event Event
		err := cur.Decode(&event)
		log.Err(err)
		events = append(events, event)
	}

	return events, cur.Err()
}

func CreatePlayerEvent(r *http.Request, steamID int64, eventType EventEnum) (err error) {

	user, err := sql.GetUserByKey("steam_id", steamID, 0)
	if err != nil {
		if err == sql.ErrRecordNotFound {
			return nil
		} else {
			return err
		}
	}

	return CreateUserEvent(r, user.ID, eventType)
}

func CreateUserEvent(r *http.Request, userID int, eventType EventEnum) (err error) {

	event := &Event{}
	event.CreatedAt = time.Now()
	event.UserID = userID
	event.Type = string(eventType)
	event.UserAgent = r.Header.Get("User-Agent")
	event.IP = r.RemoteAddr

	_, err = InsertDocument(CollectionEvents, event)
	if err != nil {
		return err
	}

	// Clear cache
	err = helpers.RemoveKeyFromMemCacheViaPubSub(helpers.MemcacheUserEventsCount(userID).Key)
	log.Err(err)

	return err
}

func CountEvents(userID int) (count int64, err error) {

	var item = helpers.MemcacheUserEventsCount(userID)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		return CountDocuments(CollectionEvents, M{"user_id": userID}, 0)
	})

	return count, err
}
