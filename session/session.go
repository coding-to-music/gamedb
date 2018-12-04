package session

import (
	"net/http"
	"sync"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/log"
	"github.com/gorilla/sessions"
	"github.com/spf13/viper"
)

const (
	PlayerID    = "id"
	PlayerName  = "name"
	UserEmail   = "email"
	PlayerLevel = "level"
	UserCountry = "country"
)

var store *sessions.CookieStore
var writeMutex = new(sync.Mutex)

// Called from main
func Init() {
	store = sessions.NewCookieStore(
		[]byte(viper.GetString("SESSION_AUTHENTICATION")),
		[]byte(viper.GetString("SESSION_ENCRYPTION")),
	)
}

func getSession(r *http.Request) (*sessions.Session, error) {

	writeMutex.Lock()

	session, err := store.Get(r, "gamedb-session")

	if viper.GetString("ENV") == string(log.EnvProd) {
		session.Options = &sessions.Options{
			MaxAge:   86400,
			Domain:   "gamedb.online",
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
		}
	} else {
		session.Options = &sessions.Options{
			MaxAge: 0,
			Path:   "/",
		}
	}

	writeMutex.Unlock()

	return session, err
}

func Read(r *http.Request, key string) (value string, err error) {

	session, err := getSession(r)
	if err != nil {
		log.Log(log.SeverityDebug, "1")
		return "", err
	}

	log.Log(log.SeverityDebug, "2")

	if session.Values[key] == nil {
		log.Log(log.SeverityDebug, "3")
		session.Values[key] = ""
	}

	log.Log(log.SeverityDebug, "4")
	log.Log(log.SeverityDebug, session.Values[key].(string))
	log.Log(log.SeverityDebug, "5")

	return session.Values[key].(string), nil
}

func GetCountryCode(r *http.Request) steam.CountryCode {

	val, err := Read(r, UserCountry)
	if err != nil || val == "" {
		log.Log(err)
		return steam.CountryUS
	}

	return steam.CountryCode(val)
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

func Write(w http.ResponseWriter, r *http.Request, name string, value string) (err error) {

	session, err := getSession(r)
	if err != nil {
		return err
	}

	session.Values[name] = value

	return session.Save(r, w)
}

func WriteMany(w http.ResponseWriter, r *http.Request, values map[string]string) (err error) {

	session, err := getSession(r)
	if err != nil {
		return err
	}

	for k, v := range values {
		session.Values[k] = v
	}

	return session.Save(r, w)
}

func Clear(w http.ResponseWriter, r *http.Request) (err error) {

	session, err := getSession(r)
	if err != nil {
		return err
	}

	session.Values = make(map[interface{}]interface{})

	return session.Save(r, w)
}

func getFlashes(w http.ResponseWriter, r *http.Request, group string) (flashes []interface{}, err error) {

	session, err := getSession(r)
	if err != nil {
		return nil, err
	}

	flashes = session.Flashes(group)

	err = session.Save(r, w)

	return flashes, err
}

func GetGoodFlashes(w http.ResponseWriter, r *http.Request) (flashes []interface{}, err error) {
	return getFlashes(w, r, "good")
}

func GetBadFlashes(w http.ResponseWriter, r *http.Request) (flashes []interface{}, err error) {
	return getFlashes(w, r, "bad")
}

func setFlash(w http.ResponseWriter, r *http.Request, flash string, group string) (err error) {

	session, err := getSession(r)
	if err != nil {
		return err
	}

	session.AddFlash(flash, group)

	return session.Save(r, w)
}

func SetGoodFlash(w http.ResponseWriter, r *http.Request, flash string) (err error) {
	return setFlash(w, r, flash, "good")
}

func SetBadFlash(w http.ResponseWriter, r *http.Request, flash string) (err error) {
	return setFlash(w, r, flash, "bad")
}

func IsLoggedIn(r *http.Request) (val bool, err error) {
	read, err := Read(r, PlayerID)
	return read != "", err
}
