package session

import (
	"net/http"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gorilla/sessions"
)

const (
	// Set if logged in
	UserID         = "user-id"
	UserEmail      = "user-email"
	UserCountry    = "user-country"
	UserShowAlerts = "user-alerts"

	// Set if player exists at login
	PlayerID    = "player-id"
	PlayerLevel = "player-level"
	PlayerName  = "player-name"
)

var writeMutex sync.Mutex
var store = sessions.NewCookieStore(
	[]byte(config.Config.SessionAuthentication.Get()),
	[]byte(config.Config.SessionEncryption.Get()),
)

func getSession(r *http.Request) (*sessions.Session, error) {

	writeMutex.Lock()

	defer writeMutex.Unlock()

	session, err := store.Get(r, "gamedb-session")
	if err == nil {
		if config.IsProd() {
			session.Options = &sessions.Options{
				MaxAge:   2592000, // 30 days
				Domain:   "gamedb.online",
				Path:     "/",
				Secure:   true,
				HttpOnly: true,
			}
		} else {
			session.Options = &sessions.Options{
				MaxAge: 2592000, // 30 days
				Path:   "/",
			}
		}
	}

	return session, err
}

func Read(r *http.Request, key string) (value string, err error) {

	session, err := getSession(r)
	if err != nil {
		return "", err
	}

	if session.Values[key] == nil {
		session.Values[key] = ""
	}

	return session.Values[key].(string), nil
}

func ReadAll(r *http.Request) (ret map[string]string, err error) {

	ret = map[string]string{}

	session, err := getSession(r)
	if err != nil {
		return ret, err
	}

	for k, v := range session.Values {
		ret[k.(string)] = v.(string)
	}

	return ret, err
}

func Write(r *http.Request, name string, value string) (err error) {

	session, err := getSession(r)
	if err != nil {
		return err
	}

	session.Values[name] = value

	return nil
}

func WriteMany(r *http.Request, values map[string]string) (err error) {

	session, err := getSession(r)
	if err != nil {
		return err
	}

	for k, v := range values {
		session.Values[k] = v
	}

	return nil
}

func Clear(r *http.Request) (err error) {

	session, err := getSession(r)
	if err != nil {
		return err
	}

	session.Values = make(map[interface{}]interface{})

	return nil
}

func getFlashes(r *http.Request, group string) (flashes []interface{}, err error) {

	session, err := getSession(r)
	if err != nil {
		return nil, err
	}

	flashes = session.Flashes(group)

	return flashes, err
}

func GetGoodFlashes(r *http.Request) (flashes []interface{}, err error) {
	return getFlashes(r, "good")
}

func GetBadFlashes(r *http.Request) (flashes []interface{}, err error) {
	return getFlashes(r, "bad")
}

func setFlash(r *http.Request, flash string, group string) (err error) {

	session, err := getSession(r)
	if err != nil {
		return err
	}

	session.AddFlash(flash, group)

	return nil
}

func SetGoodFlash(r *http.Request, flash string) (err error) {
	return setFlash(r, flash, "good")
}

func SetBadFlash(r *http.Request, flash string) (err error) {
	return setFlash(r, flash, "bad")
}

func Save(w http.ResponseWriter, r *http.Request) (err error) {

	session, err := getSession(r)
	if err != nil {
		return err
	}

	return session.Save(r, w)
}
