package db

import (
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/steam-authority/steam-authority/helpers"
)

const (
	EventLogin   = "login"
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

func (event Event) GetIP() string {

	var ips = strings.Split(event.IP, ", ")
	if len(ips) > 0 && ips[0] != "" {
		return ips[0]
	}
	return "-"
}

func (event Event) GetCreatedUnix() int64 {
	return event.CreatedAt.Unix()
}

func (event Event) GetType() string {

	switch event.Type {
	case EventLogin:
		return "User Login"
	case EventRefresh:
		return "Profile Update"
	default:
		return strings.Title(event.Type)
	}
}

func CreateEvent(r *http.Request, playerID int64, eventType string) (err error) {

	login := new(Event)
	login.CreatedAt = time.Now()
	login.PlayerID = playerID
	login.Type = eventType
	login.UserAgent = r.Header.Get("User-Agent")
	login.IP = r.Header.Get("X-Forwarded-For")

	_, err = SaveKind(login.GetKey(), login)
	return err
}

func GetEvents(playerID int64, limit int, eventType string) (logins []Event, err error) {

	client, ctx, err := GetDSClient()
	if err != nil {
		return logins, err
	}

	q := datastore.NewQuery(KindEvent)
	q = q.Filter("player_id =", playerID)
	q = q.Order("-created_at")

	if eventType != "" {
		q = q.Filter("type =", eventType)
	}

	if limit > 0 {
		q = q.Limit(limit)
	}

	_, err = client.GetAll(ctx, q, &logins)
	return logins, err
}
