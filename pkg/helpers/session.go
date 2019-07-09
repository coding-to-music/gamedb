package helpers

import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/Jleagle/session-go/session"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/oschwald/maxminddb-golang"
)

const (
	// Set if logged in
	SessionUserID         = "user-id"
	SessionUserEmail      = "user-email"
	SessionUserCountry    = "user-country"
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
)

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

func GetCountryCode(r *http.Request) steam.CountryCode {

	ccLock.Lock()
	defer ccLock.Unlock()

	var fallback = steam.CountryUS

	// Get from URL
	q := strings.ToUpper(r.URL.Query().Get("cc"))
	if q != "" && steam.ValidCountryCode(steam.CountryCode(q)) {
		return steam.CountryCode(q)
	}

	// Get from session
	val, err := session.Get(r, SessionUserCountry)
	log.Err(err)
	if err == nil && steam.ValidCountryCode(steam.CountryCode(val)) {
		return steam.CountryCode(val)
	}

	// Get from Maxmind
	db, err := maxminddb.Open(config.Config.AssetsPath.Get() + "/files/GeoLite2-Country.mmdb")
	if err != nil {
		log.Err(err)
		return steam.CountryUS
	}
	defer func() {
		err = db.Close()
		log.Err(err)
	}()

	log.Info("IP: " + r.RemoteAddr)

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
			return fallback
		}

		for _, activeCountryCode := range GetActiveCountries() {

			if record.Country.ISOCode == string(activeCountryCode) && steam.ValidCountryCode(steam.CountryCode(val)) {
				return activeCountryCode
			}
		}
	}

	return fallback
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
