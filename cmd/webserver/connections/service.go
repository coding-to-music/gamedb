package connections

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/Jleagle/session-go/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
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
	ConnectionGithub  connectionEnum = "github"
	ConnectionPatreon connectionEnum = "patreon"
	ConnectionSteam   connectionEnum = "steam"
)

type ConnectionInterface interface {
	getID(r *http.Request, token *oauth2.Token) interface{}
	getName() string
	getEnum() connectionEnum
	getConfig(login bool) oauth2.Config

	//
	LinkHandler(w http.ResponseWriter, r *http.Request)
	UnlinkHandler(w http.ResponseWriter, r *http.Request)
	LinkCallbackHandler(w http.ResponseWriter, r *http.Request)

	//
	LoginHandler(w http.ResponseWriter, r *http.Request)
	LoginCallbackHandler(w http.ResponseWriter, r *http.Request)
}

func New(s connectionEnum) ConnectionInterface {

	switch s {
	case ConnectionDiscord:
		return discordConnection{}
	case ConnectionGoogle:
		return googleConnection{}
	case ConnectionPatreon:
		return patreonConnection{}
	case ConnectionSteam:
		return steamConnection{}
	case ConnectionGithub:
		return githubConnection{}
	default:
		panic("invalid connection")
	}
}

func linkOAuth(w http.ResponseWriter, r *http.Request, c ConnectionInterface, login bool) {

	state := helpers.RandString(5, helpers.Numbers)

	err := session.Set(r, strings.ToLower(c.getName())+"-oauth-state", state)
	log.Err(err, r)

	err = session.Save(w, r)
	log.Err(err, r)

	conf := c.getConfig(login)
	url := conf.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func unlink(w http.ResponseWriter, r *http.Request, c ConnectionInterface, event mongo.EventEnum) {

	defer func() {
		err := session.Save(w, r)
		log.Err(err, r)

		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	userID, err := helpers.GetUserIDFromSesion(r)
	if err != nil {
		log.Err(err, r)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1001)")
		log.Err(err, r)
		return
	}

	// Update user
	err = sql.UpdateUserCol(userID, strings.ToLower(c.getName())+"_id", nil)
	if err != nil {
		log.Err(err, r)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1002)")
		log.Err(err, r)
		return
	}

	// Clear session
	if c.getEnum() == ConnectionSteam {
		err = session.DeleteMany(r, []string{helpers.SessionPlayerID, helpers.SessionPlayerName, helpers.SessionPlayerLevel})
		if err != nil {
			log.Err(err, r)
			err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1003)")
			log.Err(err, r)
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

func callbackOAuth(r *http.Request, c ConnectionInterface, event mongo.EventEnum, login bool) {

	realState, err := session.Get(r, strings.ToLower(c.getName())+"-oauth-state")
	if err != nil {
		log.Err(err, r)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1001)")
		log.Err(err, r)
		return
	}

	err = r.ParseForm()
	if err != nil {
		log.Err(err, r)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1002)")
		log.Err(err, r)
		return
	}

	state := r.Form.Get("state")
	if state == "" || state != realState {
		err = session.SetFlash(r, helpers.SessionBad, "Invalid state")
		log.Err(err, r)
		return
	}

	code := r.Form.Get("code")
	if code == "" {
		err = session.SetFlash(r, helpers.SessionBad, "Invalid code")
		log.Err(err, r)
		return
	}

	conf := c.getConfig(login)
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		log.Err(err, r)
		err = session.SetFlash(r, helpers.SessionBad, "Invalid token")
		log.Err(err, r)
		return
	}

	callback(r, c, event, token, login)
}

func callback(r *http.Request, c ConnectionInterface, event mongo.EventEnum, token *oauth2.Token, login bool) {

	id := c.getID(r, token)
	if id == nil {
		return
	}

	userID, err := helpers.GetUserIDFromSesion(r)
	if err != nil {
		log.Err(err, r)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1004)")
		log.Err(err, r)
		return
	}

	if login {

		err = session.SetFlash(r, helpers.SessionGood, "You have been logged in")
		log.Err(err, r)

	} else {

		// Check ID is not already in use
		_, err = sql.GetUserByKey(strings.ToLower(c.getName())+"_id", id, userID)
		if err == nil {
			err = session.SetFlash(r, helpers.SessionBad, "This "+c.getName()+" account is already linked to another Game DB account")
			log.Err(err, r)
			return
		} else if err != sql.ErrRecordNotFound {
			log.Err(err, r)
			err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1002)")
			log.Err(err, r)
			return
		}

		// Update user
		err = sql.UpdateUserCol(userID, strings.ToLower(c.getName())+"_id", id)
		if err != nil {
			log.Err(err, r)
			err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1003)")
			log.Err(err, r)
			return
		}

		// Success flash
		err = session.SetFlash(r, helpers.SessionGood, c.getName()+" account linked")
		log.Err(err, r)
	}

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
			log.Err(err, r)

		} else {

			err = session.Set(r, helpers.SessionPlayerName, player.PersonaName)
			log.Err(err, r)
		}

		if player.NeedsUpdate(mongo.PlayerUpdateManual) {

			ua := r.UserAgent()
			err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID, UserAgent: &ua})
			if err == nil {
				log.Info(log.LogNameTriggerUpdate, r, ua)
			}
			err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
			if err != nil {
				log.Err(err, r)
			} else {
				err = session.SetFlash(r, helpers.SessionGood, "Player has been queued for an update")
				log.Err(err, r)
			}
		}

		// Add player to session
		err = session.Set(r, helpers.SessionPlayerID, strconv.FormatInt(idInt64, 10))
		log.Err(err, r)
	}
}
