package db

import (
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/memcache-go/memcache"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
)

const (
	EventLogin   = "login"
	EventLogout  = "logout"
	EventRefresh = "refresh"
)

type Event struct {
	CreatedAt time.Time `datastore:"created_at"`
	Type      string    `datastore:"type"`
	PlayerID  int64     `datastore:"player_id"`
	UserAgent string    `datastore:"user_agent,noindex"`
	IP        string    `datastore:"ip,noindex"`
}

func (event Event) GetKey() (key *datastore.Key) {
	return datastore.IncompleteKey(KindEvent, nil)
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

func CountPlayerEvents(playerID int64) (count int, err error) {

	var item = helpers.MemcachePlayerEventsCount(playerID)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		var count int

		client, ctx, err := GetDSClient()
		if err != nil {
			return count, err
		}

		q := datastore.NewQuery(KindEvent).Filter("player_id = ", playerID).Limit(10000)
		count, err = client.Count(ctx, q)
		return count, err
	})

	return count, err
}

func CreateEvent(r *http.Request, playerID int64, eventType string) (err error) {

	login := new(Event)
	login.CreatedAt = time.Now()
	login.PlayerID = playerID
	login.Type = eventType
	login.UserAgent = r.Header.Get("User-Agent")
	login.IP = r.RemoteAddr

	err = SaveKind(login.GetKey(), login)
	if err != nil {
		return err
	}

	// client, ctx, err := GetMongo()
	// if err != nil {
	// 	return err
	// }
	//
	// _, err = client.Database("steam").Collection("events").InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})

	if config.Config.HasMemcache() {
		err = helpers.GetMemcache().Delete(helpers.MemcachePlayerEventsCount(playerID).Key)
		err = helpers.IgnoreErrors(err, memcache.ErrCacheMiss)
	}

	return err
}

func ChunkEvents(kinds []Event) (chunked [][]Event) {

	for i := 0; i < len(kinds); i += 500 {
		end := i + 500

		if end > len(kinds) {
			end = len(kinds)
		}

		chunked = append(chunked, kinds[i:end])
	}

	return chunked
}
