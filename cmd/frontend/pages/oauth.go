package pages

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/oauth"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

func OauthRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/out/{provider:[a-z]+}", oauthOutHandler)
	r.Get("/in/{provider:[a-z]+}", oauthInHandler)
	return r
}

func oauthOutHandler(w http.ResponseWriter, r *http.Request) {

	provider := oauth.New(oauth.ProviderEnum(chi.URLParam(r, "provider")))
	if provider == nil {
		Error404Handler(w, r)
		return
	}

	state := oauth.State{
		State: helpers.RandString(10, helpers.LettersCaps),
		Page:  r.URL.Query().Get("page"),
	}.Marshal()

	session.Set(r, "oauth-state-"+strings.ToLower(provider.GetName()), state)
	session.Save(w, r)

	conf := provider.GetConfig()
	http.Redirect(w, r, conf.AuthCodeURL(state), http.StatusFound)
}

const (
	oauthRedirectLogin    = "login"
	oauthRedirectSignup   = "signup"
	oauthRedirectSettings = "settings"
)

func oauthInHandler(w http.ResponseWriter, r *http.Request) {

	provider := oauth.New(oauth.ProviderEnum(chi.URLParam(r, "provider")))
	if provider == nil {
		Error404Handler(w, r)
		return
	}

	var token *oauth2.Token
	var err error
	var redirect string

	func() {

		// If OAuth
		if provider.GetEnum() != oauth.ProviderSteam {

			// Handle outgoing generated state
			realStateString := session.Get(r, "oauth-state-"+strings.ToLower(provider.GetName()))
			if realStateString == "" {
				session.SetFlash(r, session.SessionBad, "An error occurred (1001)")
				return
			}

			realState := oauth.State{}
			realState.Unmarshal(realStateString)

			redirect = realState.Page

			// Handle incoming state from provider
			stateString := r.URL.Query().Get("state")
			if stateString == "" {
				session.SetFlash(r, session.SessionBad, "Invalid state")
				return
			}

			state := oauth.State{}
			state.Unmarshal(stateString)

			if state.State == "" || state.State != realState.State {
				session.SetFlash(r, session.SessionBad, "Invalid state")
				return
			}

			// Swap code for token
			code := r.URL.Query().Get("code")
			if code == "" {
				session.SetFlash(r, session.SessionBad, "Invalid code")
				return
			}

			conf := provider.GetConfig()
			token, err = conf.Exchange(context.Background(), code)
			if err != nil {
				log.ErrS(err)
				session.SetFlash(r, session.SessionBad, "An error occurred (1003)")
				return
			}
		}

		// Api call to get user info
		resp, err := provider.GetUser(r, token)
		if err != nil {
			log.ErrS(err)
			session.SetFlash(r, session.SessionBad, "We were unable to fetch your details from "+provider.GetName()+" (1001)")
			return
		}

		if resp.Email == "" {
			if provider.GetEnum() == oauth.ProviderSteam {
				session.SetFlash(r, session.SessionBad, "Sorry you can not yet sign up using Steam")
			} else {
				session.SetFlash(r, session.SessionBad, "We were unable to fetch your details from "+provider.GetName()+" (1002)")
			}
			return
		}

		// Look for existing user by email
		user, err := mysql.GetUserByEmail(resp.Email)
		if err == mysql.ErrRecordNotFound {

			// Create new user
			user, err = mysql.NewUser(resp.Email, "", session.GetProductCC(r), true)
			if err != nil {
				log.ErrS(err)
				session.SetFlash(r, session.SessionGood, "Account could not be created")
			}

			// err = email_providers.GetSender().Send(
			// 	resp.Username,
			// 	resp.Email,
			// 	"",
			// 	"",
			// 	"Welcome to Game DB",
			// 	"todo", // todo
			// )
			// if err != nil {
			// 	log.Err(err.Error())
			// }

			// session.SetFlash(r, session.SessionGood, "Account created")

		} else if err != nil {
			log.ErrS(err)
			session.SetFlash(r, session.SessionGood, "An error occurred (1004)")
			return
		}

		login(r, user)

		// Check ID is not already in use
		used, err := mysql.CheckExistingUserProvider(provider.GetEnum(), resp.ID, user.ID)
		if err != nil {
			log.ErrS(err)
			session.SetFlash(r, session.SessionBad, "An error occurred (1005)")
			return
		}
		if used {
			session.SetFlash(r, session.SessionBad, "This "+provider.GetName()+" account is already linked to another Game DB account")
			return
		}

		err = mysql.UpdateUserProvider(user.ID, provider.GetEnum(), resp.Token, resp.ID, resp.Email, resp.Avatar)
		if err != nil {
			log.ErrS(err)
			session.SetFlash(r, session.SessionBad, "An error occurred (1006)")
			return
		}

		// Success flash
		if redirect == oauthRedirectSettings {
			session.SetFlash(r, session.SessionGood, provider.GetName()+" account linked")
		}
		if redirect == oauthRedirectSignup {
			session.SetFlash(r, session.SessionGood, "Account created")
		}

		// Create event
		err = mongo.CreateUserEvent(r, user.ID, mongo.EventLink(provider.GetEnum()))
		if err != nil {
			log.ErrS(err)
		}

		if provider.GetEnum() == oauth.ProviderSteam {

			i, err := strconv.ParseInt(resp.ID, 10, 64)
			if err != nil {
				log.ErrS(err)
			} else {

				// Queue for an update
				player, err := mongo.GetPlayer(i)
				if err != nil {

					err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
					if err != nil {
						log.ErrS(err)
					}

				} else {
					session.Set(r, session.SessionPlayerName, player.GetName())
				}

				if player.NeedsUpdate(mongo.PlayerUpdateManual) {

					ua := r.UserAgent()
					err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID, UserAgent: &ua})
					if err == nil {
						log.Info("player queued", zap.String("ua", ua))
					}
					err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
					if err != nil {
						log.ErrS(err)
					} else {
						session.SetFlash(r, session.SessionGood, "Player has been queued for an update")
					}
				}

				// Add player to session
				session.Set(r, session.SessionPlayerID, strconv.FormatInt(i, 10))
			}
		}

	}()

	session.Save(w, r)

	switch redirect {
	case oauthRedirectLogin:
		http.Redirect(w, r, "/login", http.StatusFound)
	case oauthRedirectSignup:
		http.Redirect(w, r, "/signup", http.StatusFound)
	case oauthRedirectSettings:
		http.Redirect(w, r, "/settings", http.StatusFound)
	default:
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
