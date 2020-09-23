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
	"github.com/gamedb/gamedb/cmd/frontend/helpers/geo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
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
	SessionGood FlashGroup = "good"
	SessionBad  FlashGroup = "bad"

	// Cookies
	SessionCookieName = "gamedb-session"
)

type FlashGroup session.FlashGroup

func InitSession() {

	sessionInit := session.Init{}
	sessionInit.AuthenticationKey = config.C.SessionAuthentication
	sessionInit.EncryptionKey = config.C.SessionEncryption
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

//
func GetUserIDFromSesion(r *http.Request) (id int, err error) {

	idx := Get(r, SessionUserID)

	if idx == "" {
		return id, errors.New("no user id set")
	}

	return strconv.Atoi(idx)
}

func GetPlayerIDFromSesion(r *http.Request) (id int64, err error) {

	idx := Get(r, SessionPlayerID)

	if idx == "" {
		return 0, nil
	}

	return strconv.ParseInt(idx, 10, 64)
}

var (
	maxMindLock sync.Mutex
	maxMindDB   *maxminddb.Reader
)

func GetProductCC(r *http.Request) steamapi.ProductCC {

	maxMindLock.Lock()
	defer maxMindLock.Unlock()

	cc := func() steamapi.ProductCC {

		// Get from URL
		q := strings.ToLower(r.URL.Query().Get("cc"))
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

		// Get from cloudflare
		q = strings.ToLower(r.Header.Get("cf-ipcountry"))
		if q == "gb" {
			q = "uk"
		}
		if q != "" && steamapi.IsProductCC(q) {
			return steamapi.ProductCC(q)
		}

		var err error

		// Get from Maxmind
		if maxMindDB == nil {
			maxMindDB, err = maxminddb.Open("./assets/GeoLite2-Country.mmdb")
			if err != nil {
				log.ErrS(err)
				return steamapi.ProductCCUS
			}
		}

		ip := net.ParseIP(geo.GetFirstIP(r.RemoteAddr))
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
				log.ErrS(err)
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

	maxMindLock.Lock()
	defer maxMindLock.Unlock()

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
				log.ErrS(err)
				return "US"
			}
		}

		ip := net.ParseIP(geo.GetFirstIP(r.RemoteAddr))
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
				log.ErrS(err)
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
		return mysql.UserLevel0
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		return mysql.UserLevel0
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
