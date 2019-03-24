package mongo

import (
	"net/http"
	"strings"
	"time"

	"github.com/Jleagle/memcache-go/memcache"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	EventLogin   = "login"
	EventLogout  = "logout"
	EventRefresh = "refresh"
)

type Event struct {
	CreatedAt time.Time
	Type      string
	PlayerID  int64
	UserAgent string
	IP        string
}

func (event Event) Key() interface{} {
	return nil
}

func (event Event) BSON() (ret interface{}) {

	return bson.M{
		"created_at": event.CreatedAt,
		"type":       event.Type,
		"player_id":  event.PlayerID,
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

	switch event.Type {
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

	switch event.Type {
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

func GetEvents(playerID int64, offset int64) (events []Event, err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return events, err
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionEvents)

	cur, err := c.Find(ctx, bson.M{"player_id": playerID}, options.Find().SetLimit(100).SetSkip(offset).SetSort(bson.M{"created_at": -1}))
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

func CreateEvent(r *http.Request, playerID int64, eventType string) (err error) {

	event := new(Event)
	event.CreatedAt = time.Now()
	event.PlayerID = playerID
	event.Type = eventType
	event.UserAgent = r.Header.Get("User-Agent")
	event.IP = r.RemoteAddr

	_, err = InsertDocument(CollectionEvents, event)
	if err != nil {
		return err
	}

	if config.Config.HasMemcache() {
		err = helpers.GetMemcache().Delete(helpers.MemcachePlayerEventsCount(playerID).Key)
		err = helpers.IgnoreErrors(err, memcache.ErrCacheMiss)
	}

	return err
}

func CountEvents(playerID int64) (count int64, err error) {

	var item = helpers.MemcachePlayerEventsCount(playerID)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		return CountDocuments(CollectionEvents, bson.M{"player_id": playerID})
	})

	return count, err
}
