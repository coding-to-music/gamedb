package datastore

import (
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
)

type Login struct {
	CreatedAt time.Time `datastore:"created_at"`
	PlayerID  int       `datastore:"player_id"`
	UserAgent string    `datastore:"user_agent,noindex"`
	IP        string    `datastore:"ip,noindex"`
}

func (login Login) GetKey() (key *datastore.Key) {
	return datastore.IncompleteKey(KindLogin, nil)
}

func (login Login) GetTime() (t string) {
	return login.CreatedAt.Format(time.RFC822)
}

func CreateLogin(playerID int, r *http.Request) (err error) {

	login := new(Login)
	login.CreatedAt = time.Now()
	login.PlayerID = playerID
	login.UserAgent = r.Header.Get("User-Agent")
	login.IP = r.RemoteAddr

	_, err = SaveKind(login.GetKey(), login)

	return err
}

func GetLogins(playerID int, limit int) (logins []Login, err error) {

	client, ctx, err := getDSClient()
	if err != nil {
		return logins, err
	}

	q := datastore.NewQuery(KindLogin).Order("-created_at").Limit(limit)
	q = q.Filter("player_id =", playerID)

	client.GetAll(ctx, q, &logins)

	return logins, err
}
