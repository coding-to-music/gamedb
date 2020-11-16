package mongo

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/oauth"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
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

func (event EventEnum) ToString() string {

	if strings.HasPrefix(string(event), "link") || strings.HasPrefix(string(event), "unlink") {
		parts := strings.Split(string(event), "-")
		if len(parts) == 2 {
			parts[0] = strings.Title(parts[0])

			provider := oauth.New(oauth.ProviderEnum(parts[1]))
			if provider != nil {
				parts[1] = provider.GetName()
			} else {
				log.ErrS("invalid provider", zap.String("provider", parts[1]))
			}
		}
		return strings.Join(parts, " ")
	}

	switch event {
	case EventLogin:
		return "User Login"
	case EventLogout:
		return "User Logout"
	case EventRefresh:
		return "Profile Update"
	case EventForgotPassword:
		return "Forgot Password"
	default:
		return strings.Title(string(event))
	}
}

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

func GetEvents(filter bson.D, offset int64) (events []Event, err error) {

	var sort = bson.D{{"created_at", -1}}

	cur, ctx, err := Find(CollectionEvents, offset, 100, sort, filter, nil, nil)
	if err != nil {
		return events, err
	}

	defer closeCursor(cur, ctx)

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

func GetEventCounts(userID int) (counts []StringCount, err error) {

	var item = memcache.MemcacheUserEvents(userID)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &counts, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return counts, err
		}

		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.M{"user_id": userID}}},
			{{Key: "$group", Value: bson.M{"_id": "$type", "count": bson.M{"$sum": 1}}}},
		}

		cur, err := client.Database(config.C.MongoDatabase, options.Database()).Collection(CollectionEvents.String()).Aggregate(ctx, pipeline, options.Aggregate())
		if err != nil {
			return counts, err
		}

		defer closeCursor(cur, ctx)

		for cur.Next(ctx) {

			var count StringCount
			err := cur.Decode(&count)
			if err != nil {
				log.ErrS(err, count.ID)
			}
			counts = append(counts, count)
		}

		sort.Slice(counts, func(i, j int) bool {
			return counts[i].Count > counts[j].Count
		})

		return counts, cur.Err()
	})

	return counts, err
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
	if err != nil {
		return err
	}

	return memcache.Delete(memcache.MemcacheUserEvents(userID).Key)
}
