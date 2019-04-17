package mongo

import (
	"net/http"
	"strings"
	"time"

	"github.com/Jleagle/memcache-go/memcache"
	"github.com/gamedb/website/pkg"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	EventLogin   = "login"
	EventLogout  = "logout"
	EventRefresh = "refresh"
	EventPatreon = "patreon"
)

type Event struct {
	CreatedAt time.Time `bson:"created_at"`
	Type      string    `bson:"type"`
	PlayerID  int64     `bson:"player_id"`
	UserAgent string    `bson:"user_agent"`
	IP        string    `bson:"ip"`
}

func (event Event) BSON() (ret interface{}) {

	return pkg.M{
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
	return event.CreatedAt.Format(pkg.DateTime)
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

	client, ctx, err := pkg.getMongo()
	if err != nil {
		return events, err
	}

	c := client.Database(pkg.MongoDatabase, options.Database()).Collection(pkg.CollectionEvents.String())

	cur, err := c.Find(ctx, pkg.M{"player_id": playerID}, options.Find().SetLimit(100).SetSkip(offset).SetSort(pkg.M{"created_at": -1}))
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

	event := &Event{}
	event.CreatedAt = time.Now()
	event.PlayerID = playerID
	event.Type = eventType
	event.UserAgent = r.Header.Get("User-Agent")
	event.IP = r.RemoteAddr

	_, err = pkg.InsertDocument(pkg.CollectionEvents, event)
	if err != nil {
		return err
	}

	if config.Config.HasMemcache() {
		err = pkg.GetMemcache().Delete(pkg.MemcachePlayerEventsCount(playerID).Key)
		err = helpers.IgnoreErrors(err, memcache.ErrCacheMiss)
	}

	return err
}

func CountEvents(playerID int64) (count int64, err error) {

	var item = pkg.MemcachePlayerEventsCount(playerID)

	err = pkg.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		return pkg.CountDocuments(pkg.CollectionEvents, pkg.M{"player_id": playerID})
	})

	return count, err
}
