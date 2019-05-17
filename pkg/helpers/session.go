package helpers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Jleagle/session-go/session"
)

const (
	// Set if logged in
	SessionUserID         = "user-id"
	SessionUserEmail      = "user-email"
	SessionUserCountry    = "user-country"
	SessionUserShowAlerts = "user-alerts"

	// Set if player exists at login
	SessionPlayerID    = "player-id"
	SessionPlayerLevel = "player-level"
	SessionPlayerName  = "player-name"

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
