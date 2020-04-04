package session

import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/Jleagle/session-go/session"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/oschwald/maxminddb-golang"
)

const (
	// Set if logged in
	SessionUserID         = "user-id"
	SessionUserEmail      = "user-email"
	SessionUserProdCC     = "user-country"
	SessionUserShowAlerts = "user-alerts"
	SessionUserAPIKey     = "user-api-key"
	SessionUserLevel      = "user-level"

	// Set if player exists at login
	SessionPlayerID    = "player-id"
	SessionPlayerLevel = "player-level"
	SessionPlayerName  = "player-name"

	//
	SessionLastPage    = "last-page"
	SessionCountryCode = "country-code"

	// Flash groups
	SessionGood session.FlashGroup = "good"
	SessionBad  session.FlashGroup = "bad"

	// Cookies
	SessionCookieName = "gamedb-session"
)

func InitSession() {

	sessionInit := session.Init{}
	sessionInit.AuthenticationKey = config.Config.SessionAuthentication.Get()
	sessionInit.EncryptionKey = config.Config.SessionEncryption.Get()
	sessionInit.CookieName = SessionCookieName
	sessionInit.CookieOptions = sessions.Options{
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode, // Can't be strict, stops oauth callbacks working
		MaxAge:   2419200,              // 30 days
		Path:     "/",
		Domain:   "",
		Secure:   config.IsProd(),
	}

	session.Initialise(sessionInit)
}

func Get(r *http.Request, key string) (value string) {

	val, err := session.Get(r, key)
	logSessionError(err)
	return val
}

func Set(r *http.Request, name string, value string) {

	err := session.Set(r, name, value)
	logSessionError(err)
}

func SetMany(r *http.Request, values map[string]string) {

	err := session.SetMany(r, values)
	logSessionError(err)
}

func GetFlashes(r *http.Request, group session.FlashGroup) (flashes []string) {

	flashes, err := session.GetFlashes(r, group)
	logSessionError(err)

	return flashes
}

func Save(w http.ResponseWriter, r *http.Request) {

	err := session.Save(w, r)
	logSessionError(err)
}

func logSessionError(err error) {

	if err != nil {

		if val, ok := err.(securecookie.Error); ok {
			if val.IsUsage() || val.IsDecode() {
				log.Info(val.Error())
				return
			}
		}

		log.Err(err)
	}
}

//
func GetUserIDFromSesion(r *http.Request) (id int, err error) {

	idx := Get(r, SessionUserID)

	if idx == "" {
		return id, errors.New("no user id set")
	}

	return strconv.Atoi(idx)
}

var (
	ccLock    sync.Mutex
	maxMindDB *maxminddb.Reader
)

func GetProductCC(r *http.Request) steamapi.ProductCC {

	ccLock.Lock()
	defer ccLock.Unlock()

	cc := func() steamapi.ProductCC {

		// Get from URL
		q := strings.ToUpper(r.URL.Query().Get("cc"))
		if q != "" && steamapi.IsProductCC(q) {
			return steamapi.ProductCC(q)
		}

		// Get from session
		val := Get(r, SessionUserProdCC)
		if val != "" && steamapi.IsProductCC(val) {
			return steamapi.ProductCC(val)
		}

		// If local
		if strings.Contains(r.RemoteAddr, "[::1]:") {
			return steamapi.ProductCCUK
		}

		var err error

		// Get from Maxmind
		if maxMindDB == nil {
			maxMindDB, err = maxminddb.Open("./assets/GeoLite2-Country.mmdb")
			if err != nil {
				log.Err(err)
				return steamapi.ProductCCUS
			}
		}

		ip := net.ParseIP(r.RemoteAddr)
		if ip != nil {

			// More fields available @ https://github.com/oschwald/geoip2-golang/blob/master/reader.go
			// Only using what we need is faster
			var record struct {
				Country struct {
					ISOCode           string `maxminddb:"iso_code"`
					IsInEuropeanUnion bool   `maxminddb:"is_in_european_union"`
				} `maxminddb:"country"`
			}

			err = maxMindDB.Lookup(ip, &record)
			if err != nil {
				log.Err(err)
				return steamapi.ProductCCUS
			}

			for _, cc := range i18n.GetProdCCs(true) {
				for _, code := range cc.CountryCodes {
					if record.Country.ISOCode == code {
						return cc.ProductCode
					}
				}
			}
		}

		return steamapi.ProductCCUS
	}()

	Set(r, SessionUserProdCC, string(cc))

	return cc
}

func GetCountryCode(r *http.Request) string {

	ccLock.Lock()
	defer ccLock.Unlock()

	cc := func() string {

		// Get from session
		val := Get(r, SessionCountryCode)
		if val != "" {
			return val
		}

		// If local
		if strings.Contains(r.RemoteAddr, "[::1]:") {
			return "GB"
		}

		var err error

		// Get from Maxmind
		if maxMindDB == nil {
			maxMindDB, err = maxminddb.Open("./assets/GeoLite2-Country.mmdb")
			if err != nil {
				log.Err(err)
				return "US"
			}
		}

		ip := net.ParseIP(r.RemoteAddr)
		if ip != nil {

			// More fields available @ https://github.com/oschwald/geoip2-golang/blob/master/reader.go
			// Only using what we need is faster
			var record struct {
				Country struct {
					ISOCode           string `maxminddb:"iso_code"`
					IsInEuropeanUnion bool   `maxminddb:"is_in_european_union"`
				} `maxminddb:"country"`
			}

			err = maxMindDB.Lookup(ip, &record)
			if err != nil {
				log.Err(err)
				return "US"
			}

			return record.Country.ISOCode
		}

		return "US"
	}()

	Set(r, SessionCountryCode, cc)

	return cc
}

func GetUserLevel(r *http.Request) int {

	val := Get(r, SessionUserLevel)
	if val == "" {
		return 0
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}

	return i
}

func IsAdmin(r *http.Request) bool {

	return Get(r, SessionUserID) == "1"
}

func IsLoggedIn(r *http.Request) (val bool) {

	read := Get(r, SessionUserEmail)
	return read != ""
}
