package session

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Jleagle/session-go/session"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/geo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
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
		Domain:   "gamedb.online",
		Secure:   true,
	}

	if config.IsLocal() {
		sessionInit.CookieOptions.Secure = false
		sessionInit.CookieOptions.Domain = "" // Works on any local ip
	}

	session.Initialise(sessionInit)
}

//
func GetUserIDFromSesion(r *http.Request) (id int) {

	idx := Get(r, SessionUserID)

	if idx == "" {
		return 0
	}

	id, err := strconv.Atoi(idx)
	if err != nil {
		log.ErrS(err)
	}
	return id
}

func GetPlayerIDFromSesion(r *http.Request) (id int64) {

	idx := Get(r, SessionPlayerID)

	if idx == "" {
		return 0
	}

	id, err := strconv.ParseInt(idx, 10, 64)
	if err != nil {
		log.ErrS(err)
	}
	return id
}

func GetProductCC(r *http.Request) steamapi.ProductCC {

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
		if q != "" {
			if q == "gb" {
				q = "uk" // Convert to Steam's incorrect cc
			}
			if steamapi.IsProductCC(q) {
				return steamapi.ProductCC(q)
			}
		}

		record, err := geo.GetLocation(r.RemoteAddr)
		if err != nil {
			err = helpers.IgnoreErrors(err, geo.ErrInvalidIP)
			if err != nil {
				log.Err(err.Error(), zap.String("ip", r.RemoteAddr))
			}
			return steamapi.ProductCCUS
		}

		for _, cc := range i18n.GetProdCCs(true) {
			for _, code := range cc.CountryCodes {
				if record.Country.ISOCode == code {
					return cc.ProductCode
				}
			}
		}

		return steamapi.ProductCCUS
	}()

	Set(r, SessionUserProdCC, string(cc))

	return cc
}

func GetCountryCode(r *http.Request) string {

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

		record, err := geo.GetLocation(r.RemoteAddr)
		if err != nil {
			err = helpers.IgnoreErrors(err, geo.ErrInvalidIP)
			if err != nil {
				log.Err(err.Error(), zap.String("ip", r.RemoteAddr))
			}
			return "US"
		}

		return record.Country.ISOCode
	}()

	Set(r, SessionCountryCode, cc)

	return cc
}

func GetUserLevel(r *http.Request) mysql.UserLevel {

	val := Get(r, SessionUserLevel)
	if val == "" {
		return mysql.UserLevel0
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		log.ErrS(err)
		return mysql.UserLevel0
	}

	return mysql.UserLevel(i)
}

func IsAdmin(r *http.Request) bool {

	return Get(r, SessionUserID) == "1"
}

func IsLoggedIn(r *http.Request) (val bool) {

	return Get(r, SessionUserID) != "" && Get(r, SessionUserEmail) != ""
}
