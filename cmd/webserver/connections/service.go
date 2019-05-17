package connections

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/Jleagle/session-go/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"golang.org/x/oauth2"
)

type connectionEnum string

var (
	ConnectionDiscord connectionEnum = "discord"
	ConnectionGoogle  connectionEnum = "google"
	ConnectionPatreon connectionEnum = "patreon"
	ConnectionSteam   connectionEnum = "steam"
)

type ConnectionInterface interface {
	getID(r *http.Request, token *oauth2.Token) interface{}
	getName() string
	getEnum() connectionEnum
	getConfig() oauth2.Config
	getEmptyVal() interface{}
	LinkHandler(w http.ResponseWriter, r *http.Request)
	UnlinkHandler(w http.ResponseWriter, r *http.Request)
	CallbackHandler(w http.ResponseWriter, r *http.Request)
}

func New(s connectionEnum) ConnectionInterface {

	switch s {
	case ConnectionDiscord:
		return discord{}
	case ConnectionGoogle:
		return google{}
	case ConnectionPatreon:
		return patreon{}
	case ConnectionSteam:
		return steam{}
	default:
		panic("invalid connection")
	}
}

func linkOAuth(w http.ResponseWriter, r *http.Request, c ConnectionInterface) {

	state := helpers.RandString(5, helpers.Numbers)

	err := session.Set(r, strings.ToLower(c.getName())+"-oauth-state", state)
	log.Err(err)

	err = session.Save(w, r)
	log.Err(err)

	conf := c.getConfig()
	url := conf.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func unlink(w http.ResponseWriter, r *http.Request, c ConnectionInterface, event mongo.EventEnum) {

	defer func() {
		err := session.Save(w, r)
		log.Err(err)

		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	userID, err := helpers.GetUserIDFromSesion(r)
	if err != nil || userID == 0 {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1001)")
		log.Err(err)
		return
	}

	// Update user
	err = sql.UpdateUserCol(userID, strings.ToLower(c.getName())+"_id", c.getEmptyVal())
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1002)")
		log.Err(err)
		return
	}

	// Clear session
	if c.getEnum() == ConnectionSteam {
		err = session.DeleteMany(r, []string{helpers.SessionPlayerID, helpers.SessionPlayerName, helpers.SessionPlayerLevel})
		if err != nil {
			log.Err(err)
			err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1003)")
			log.Err(err)
			return
		}
	}

	// Flash message
	err = session.SetFlash(r, helpers.SessionGood, c.getName()+" unlinked")
	log.Err(err, r)

	// Create event
	err = mongo.CreateUserEvent(r, userID, event)
	if err != nil {
		log.Err(err, r)
	}
}

func callbackOAuth(w http.ResponseWriter, r *http.Request, c ConnectionInterface, event mongo.EventEnum) {

	realState, err := session.Get(r, strings.ToLower(c.getName())+"-oauth-state")
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1001)")
		log.Err(err)
		return
	}

	err = r.ParseForm()
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1002)")
		log.Err(err)
		return
	}

	state := r.Form.Get("state")
	if state == "" || state != realState {
		err = session.SetFlash(r, helpers.SessionBad, "Invalid state")
		log.Err(err)
		return
	}

	code := r.Form.Get("code")
	if code == "" {
		err = session.SetFlash(r, helpers.SessionBad, "Invalid code")
		log.Err(err)
		return
	}

	conf := c.getConfig()
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "Invalid token")
		log.Err(err)
		return
	}

	callback(w, r, c, event, token)
}

func callback(w http.ResponseWriter, r *http.Request, c ConnectionInterface, event mongo.EventEnum, token *oauth2.Token) {

	id := c.getID(r, token)
	if id == nil {
		return
	}

	userID, err := helpers.GetUserIDFromSesion(r)
	if err != nil || userID == 0 {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1004)")
		log.Err(err)
		return
	}

	// Check ID is not already in use
	_, err = sql.GetUserByKey(strings.ToLower(c.getName())+"_id", id, userID)
	if err == nil {
		err = session.SetFlash(r, helpers.SessionBad, "This "+c.getName()+" account is already linked to another Game DB account")
		log.Err(err)
		return
	} else if err != sql.ErrRecordNotFound {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1002)")
		log.Err(err)
		return
	}

	// Update user
	err = sql.UpdateUserCol(userID, strings.ToLower(c.getName())+"_id", id)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1003)")
		log.Err(err)
		return
	}

	// Success flash
	err = session.SetFlash(r, helpers.SessionGood, c.getName()+" account linked")
	log.Err(err)

	// Create event
	err = mongo.CreateUserEvent(r, userID, event)
	if err != nil {
		log.Err(err, r)
	}

	if c.getEnum() == ConnectionSteam {

		idInt64 := id.(int64)

		// Queue for an update
		player, err := mongo.GetPlayer(idInt64)
		if err != nil {
			err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
			log.Err(err)
		} else {
			if player.ShouldUpdate(r.UserAgent(), mongo.PlayerUpdateManual) {
				err = queue.ProducePlayer(player.ID)
				log.Err(err, r)

				// Queued flash
				err = session.SetFlash(r, helpers.SessionGood, "Player has been queued for an update")
				log.Err(err)
			}
		}

		// Update session
		err = session.Set(r, helpers.SessionPlayerID, strconv.FormatInt(idInt64, 10))
		log.Err(err)
	}
}
