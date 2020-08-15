package oauth

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

//
type oauthError struct {
	error error
	flash string
}

func (oe oauthError) Error() string {
	if oe.error != nil {
		return oe.error.Error()
	}
	return ""
}

//
type ConnectionEnum string

var (
	ConnectionDiscord ConnectionEnum = "discord"
	ConnectionGoogle  ConnectionEnum = "google"
	ConnectionGithub  ConnectionEnum = "github"
	ConnectionPatreon ConnectionEnum = "patreon"
	ConnectionSteam   ConnectionEnum = "steam"

	Connections = map[ConnectionEnum]bool{
		ConnectionDiscord: true,
		ConnectionGoogle:  true,
		ConnectionGithub:  true,
		ConnectionPatreon: true,
		ConnectionSteam:   true,
	}
)

//
type ConnectionInterface interface {
	getID(r *http.Request, token *oauth2.Token) (string, error)
	getName() string
	getEnum() ConnectionEnum
	getConfig(login bool) oauth2.Config

	//
	LinkHandler(w http.ResponseWriter, r *http.Request)
	UnlinkHandler(w http.ResponseWriter, r *http.Request)
	LinkCallbackHandler(w http.ResponseWriter, r *http.Request)

	//
	LoginHandler(w http.ResponseWriter, r *http.Request)
	LoginCallbackHandler(w http.ResponseWriter, r *http.Request)
}

func New(s ConnectionEnum) ConnectionInterface {

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

//
type baseConnection struct {
}

//
func (bc baseConnection) linkOAuth(w http.ResponseWriter, r *http.Request, c ConnectionInterface, login bool) {

	state := helpers.RandString(5, helpers.Numbers)

	session.Set(r, strings.ToLower(c.getName())+"-oauth-state", state)
	session.Save(w, r)

	conf := c.getConfig(login)
	url := conf.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func (bc baseConnection) unlink(w http.ResponseWriter, r *http.Request, c ConnectionInterface, event mongo.EventEnum) {

	defer func() {
		session.Save(w, r)
		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	userID, err := session.GetUserIDFromSesion(r)
	if err != nil {
		zap.S().Error(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1001)")
		return
	}

	// Update user
	err = mysql.UpdateUserCol(userID, strings.ToLower(c.getName())+"_id", nil)
	if err != nil {
		zap.S().Error(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1002)")
		return
	}

	// Clear session
	if c.getEnum() == ConnectionSteam {
		session.DeleteMany(r, []string{session.SessionPlayerID, session.SessionPlayerName, session.SessionPlayerLevel})
	}

	// Flash message
	session.SetFlash(r, session.SessionGood, c.getName()+" unlinked")

	// Create event
	err = mongo.CreateUserEvent(r, userID, event)
	if err != nil {
		zap.S().Error(err)
	}
}

func (bc baseConnection) callbackOAuth(r *http.Request, c ConnectionInterface, event mongo.EventEnum, login bool) {

	var err error

	realState := session.Get(r, strings.ToLower(c.getName())+"-oauth-state")
	if realState == "" {
		session.SetFlash(r, session.SessionBad, "An error occurred (1001)")
		return
	}

	err = r.ParseForm()
	if err != nil {
		zap.S().Error(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1002)")
		return
	}

	state := r.Form.Get("state")
	if state == "" || state != realState {
		session.SetFlash(r, session.SessionBad, "Invalid state")
		return
	}

	code := r.Form.Get("code")
	if code == "" {
		session.SetFlash(r, session.SessionBad, "Invalid code")
		return
	}

	conf := c.getConfig(login)
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		zap.S().Error(err)
		session.SetFlash(r, session.SessionBad, "Invalid token")
		return
	}

	bc.callback(r, c, event, token, login)
}

func (bc baseConnection) callback(r *http.Request, c ConnectionInterface, event mongo.EventEnum, token *oauth2.Token, login bool) {

	id, err := c.getID(r, token)
	if err != nil {
		zap.S().Error(err)
		if val, ok := err.(oauthError); ok {
			session.SetFlash(r, session.SessionBad, val.flash)
		}
		return
	}

	userID, err := session.GetUserIDFromSesion(r)
	if err != nil {
		session.SetFlash(r, session.SessionBad, "An error occurred (1004)")
		return
	}

	if login {

		session.SetFlash(r, session.SessionGood, "You have been logged in")

	} else {

		// Check ID is not already in use
		_, err = mysql.GetUserByKey(strings.ToLower(c.getName())+"_id", id, userID)
		if err == nil {
			session.SetFlash(r, session.SessionBad, "This "+c.getName()+" account is already linked to another Game DB account")
			return
		} else if err != mysql.ErrRecordNotFound {
			zap.S().Error(err)
			session.SetFlash(r, session.SessionBad, "An error occurred (1002)")
			return
		}

		// Update user
		err = mysql.UpdateUserCol(userID, strings.ToLower(c.getName())+"_id", id)
		if err != nil {
			zap.S().Error(err)
			session.SetFlash(r, session.SessionBad, "An error occurred (1003)")
			return
		}

		// Success flash
		session.SetFlash(r, session.SessionGood, c.getName()+" account linked")
	}

	// Create event
	err = mongo.CreateUserEvent(r, userID, event)
	if err != nil {
		zap.S().Error(err)
	}

	if c.getEnum() == ConnectionSteam {

		i, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			zap.S().Error(err)
		} else {

			// Queue for an update
			player, err := mongo.GetPlayer(i)
			if err != nil {

				err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
				zap.S().Error(err)

			} else {
				session.Set(r, session.SessionPlayerName, player.GetName())
			}

			if player.NeedsUpdate(mongo.PlayerUpdateManual) {

				ua := r.UserAgent()
				err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID, UserAgent: &ua})
				if err == nil {
					zap.S().Info(log.LogNameTriggerUpdate, r, ua)
				}
				err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
				if err != nil {
					zap.S().Error(err)
				} else {
					session.SetFlash(r, session.SessionGood, "Player has been queued for an update")
				}
			}

			// Add player to session
			session.Set(r, session.SessionPlayerID, strconv.FormatInt(i, 10))
		}
	}
}
