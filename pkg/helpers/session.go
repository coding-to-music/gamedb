package helpers

import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/session-go/session"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
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
	SessionLastPage = "last-page"

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
		SameSite: http.SameSiteDefaultMode, // Can't be strict, stops oauth callbacks working
		MaxAge:   2419200,                  // 30 days
		Path:     "/",
		Domain:   "", // https://scotthelme.co.uk/tough-cookies/
		Secure:   false,
	}

	if config.IsProd() {
		sessionInit.CookieName = "__Host-" + sessionInit.CookieName // https://scotthelme.co.uk/tough-cookies/
		sessionInit.CookieOptions.Secure = true
	}

	session.Initialise(sessionInit)
}

// Expires the session cookie if it's corrupt
func HandleSessionError(w http.ResponseWriter, r *http.Request, err error) {

	if err != nil && strings.Contains(err.Error(), "base64 decode failed") {

		cook, _ := r.Cookie(SessionCookieName)
		cook.Expires = time.Now().Add(-time.Second)
		cook.Value = ""

		http.SetCookie(w, cook)
	}
}

//
func GetUserIDFromSesion(r *http.Request) (id int, err error) {

	idx, err := session.Get(r, SessionUserID)
	if err != nil {
		return id, err
	}

	if idx == "" {
		return id, errors.New("no user id set")
	}

	return strconv.Atoi(idx)
}

var ccLock sync.Mutex

func GetProductCC(r *http.Request) steam.ProductCC {

	ccLock.Lock()
	defer ccLock.Unlock()

	cc := func() steam.ProductCC {

		// Get from URL
		q := strings.ToUpper(r.URL.Query().Get("cc"))
		if q != "" && steam.IsProductCC(q) {
			return steam.ProductCC(q)
		}

		// Get from session
		val, err := session.Get(r, SessionUserProdCC)
		log.Err(err)
		if err == nil && steam.IsProductCC(val) {
			return steam.ProductCC(val)
		}

		// Get from Maxmind
		db, err := maxminddb.Open("./files/GeoLite2-Country.mmdb")
		if err != nil {
			log.Err(err)
			return steam.ProductCCUS
		}
		defer func() {
			err = db.Close()
			log.Err(err)
		}()

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

			err = db.Lookup(ip, &record)
			if err != nil {
				log.Err(err)
				return steam.ProductCCUS
			}

			for _, cc := range GetProdCCs(true) {
				for _, code := range cc.CountryCodes {
					if record.Country.ISOCode == string(code) {
						return cc.ProductCode
					}
				}
			}
		}

		return steam.ProductCCUS
	}()

	err := session.Set(r, SessionUserProdCC, string(cc))
	log.Err(err)

	return cc
}

func GetUserLevel(r *http.Request) int {

	val, err := session.Get(r, SessionUserLevel)
	if err != nil {
		return 0
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}

	return i
}

func IsAdmin(r *http.Request) bool {

	id, err := session.Get(r, SessionUserID)
	log.Err(err)

	return id == "1"
}

func IsLoggedIn(r *http.Request) (val bool, err error) {

	read, err := session.Get(r, SessionUserEmail)
	return read != "", err
}
