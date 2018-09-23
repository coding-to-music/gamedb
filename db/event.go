package db

import (
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/steam-authority/steam-authority/helpers"
)

const (
	EVENT_LOGIN = "login"
)

type Event struct {
	CreatedAt time.Time `datastore:"created_at"`
	Type      string    `datastore:"type"`
	PlayerID  int64     `datastore:"player_id"`
	UserAgent string    `datastore:"user_agent,noindex"`
	IP        string    `datastore:"ip,noindex"`
}

func (login Event) GetKey() (key *datastore.Key) {
	return datastore.IncompleteKey(KindEvent, nil)
}

func (login Event) GetCreatedNice() (t string) {
	return login.CreatedAt.Format(helpers.DateTime)
}

func (login Event) GetCreatedUnix() int64 {
	return login.CreatedAt.Unix()
}

func CreateEvent(r *http.Request, playerID int64, eventType string) (err error) {

	login := new(Event)
	login.CreatedAt = time.Now()
	login.PlayerID = playerID
	login.Type = eventType
	login.UserAgent = r.Header.Get("User-Agent")
	login.IP = r.Header.Get("X-Forwarded-For")

	if login.IP == "" {
		login.IP = "127.0.0.1"
	}

	_, err = SaveKind(login.GetKey(), login)

	return err
}

func GetEvents(playerID int64, limit int, eventType string) (logins []Event, err error) {

	client, ctx, err := GetDSClient()
	if err != nil {
		return logins, err
	}

	q := datastore.NewQuery(KindEvent).Order("-created_at").Limit(limit)
	q = q.Filter("player_id =", playerID)

	if eventType != "" {
		q = q.Filter("type =", eventType)
	}

	_, err = client.GetAll(ctx, q, &logins)
	if err != nil {
		return
	}

	return logins, err
}
