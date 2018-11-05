package session

import (
	"net/http"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gorilla/sessions"
	"github.com/spf13/viper"
)

const (
	PlayerID    = "id"
	PlayerName  = "name"
	PlayerLevel = "level"
	UserCountry = "country"
)

var store *sessions.CookieStore

// Called from main
func Init() {
	store = sessions.NewCookieStore(
		[]byte(viper.GetString("SESSION_AUTHENTICATION")),
		[]byte(viper.GetString("SESSION_ENCRYPTION")),
	)
}

func getSession(r *http.Request) (*sessions.Session, error) {

	session, err := store.Get(r, "gamedb-session")
	if err != nil {
		return session, err
	}

	session.Options = &sessions.Options{
		MaxAge: 0, // Session
		Path:   "/",
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

	return session.Values[key].(string), err
}

func GetCountryCode(r *http.Request) steam.CountryCode {

	val, err := Read(r, UserCountry)
	if err != nil {
		return steam.CountryUS
	}
	return steam.CountryCode(val)
}

func ReadAll(r *http.Request) (value map[interface{}]interface{}, err error) {

	session, err := getSession(r)
	if err != nil {
		return value, err
	}

	return session.Values, err
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

	err = session.Save(r, w)
	if err != nil {
		return err
	}

	return nil
}

func getFlashes(w http.ResponseWriter, r *http.Request, group string) (flashes []interface{}, err error) {

	session, err := getSession(r)
	if err != nil {
		return nil, err
	}

	flashes = session.Flashes(group)

	err = session.Save(r, w)
	if err != nil {
		return nil, err
	}

	return flashes, nil
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
