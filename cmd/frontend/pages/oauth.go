package pages

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/dghubble/oauth1"
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
)

const (
	authPageLogin    = "login"
	authPageSignup   = "signup"
	authPageSettings = "settings"
)

func OauthRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/out/{provider:[a-z]+}", providerRedirect)
	r.Get("/in/{provider:[a-z]+}", providerCallback)

	return r
}

func providerRedirect(w http.ResponseWriter, r *http.Request) {

	provider := oauth.New(oauth.ProviderEnum(chi.URLParam(r, "provider")))
	if provider, ok := provider.(oauth.OAuth2Provider); ok {
		providerOAuth2Redirect(w, r, provider)
	}
	if provider, ok := provider.(oauth.OAuth1Provider); ok {
		providerOAuth1Redirect(w, r, provider)
	}
	if provider, ok := provider.(oauth.OpenIDProvider); ok {
		providerOpenIDRedirect(w, r, provider)
	}

	Error404Handler(w, r)
}

func providerCallback(w http.ResponseWriter, r *http.Request) {

	provider := oauth.New(oauth.ProviderEnum(chi.URLParam(r, "provider")))
	if provider, ok := provider.(oauth.OAuth2Provider); ok {
		providerOAuth2Callback(w, r, provider)
	}
	if provider, ok := provider.(oauth.OAuth1Provider); ok {
		providerOAuth1Callback(w, r, provider)
	}
	if provider, ok := provider.(oauth.OpenIDProvider); ok {
		providerOpenIDCallback(w, r, provider)
	}

	Error404Handler(w, r)
}

func providerOAuth2Redirect(w http.ResponseWriter, r *http.Request, provider oauth.OAuth2Provider) {

	state := oauth.State{
		State: helpers.RandString(10, helpers.LettersCaps),
		Page:  r.URL.Query().Get("page"),
	}.Marshal()

	session.Set(r, "oauth-state-"+strings.ToLower(provider.GetName()), state)
	session.Save(w, r)

	provider.Redirect(w, r, state)
}

func providerOAuth1Redirect(w http.ResponseWriter, r *http.Request, provider oauth.OAuth1Provider) {

	u, secret, err := provider.Redirect()
	if err != nil {
		log.ErrS(err)
	}

	state := oauth.State{
		State: secret,
		Page:  r.URL.Query().Get("page"),
	}.Marshal()

	session.Set(r, "oauth-state-"+strings.ToLower(provider.GetName()), state)
	session.Save(w, r)

	http.Redirect(w, r, u, http.StatusFound)
}

func providerOpenIDRedirect(w http.ResponseWriter, r *http.Request, provider oauth.OpenIDProvider) {

	provider.Redirect(w, r, r.URL.Query().Get("page"))
}

func providerOAuth2Callback(w http.ResponseWriter, r *http.Request, provider oauth.OAuth2Provider) {

	var page string

	defer oauthRedirect(w, r, &page)

	// Get generated state
	realStateString := session.Get(r, "oauth-state-"+strings.ToLower(provider.GetName()))
	if realStateString == "" {
		session.SetFlash(r, session.SessionBad, "An error occurred (1001)")
		return
	}

	realState := oauth.State{}
	realState.Unmarshal(realStateString)

	page = realState.Page

	// Get incoming state from provider
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
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1003)")
		return
	}

	// Api call to get user info
	resp, err := provider.GetUser(token)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "We were unable to fetch your details from "+provider.GetName()+" (1001)")
		return
	}

	var user mysql.User

	if resp.Email == "" {
		session.SetFlash(r, session.SessionBad, "We were unable to fetch your details from "+provider.GetName()+" (1002)")
		return
	}

	// Look for existing user by email
	user, err = mysql.GetUserByEmail(resp.Email)
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

	//
	oauthHandleUser(provider, resp, user, page, r)
}

func providerOAuth1Callback(w http.ResponseWriter, r *http.Request, provider oauth.OAuth1Provider) {

	var page string

	defer oauthRedirect(w, r, &page)

	// Get state from session
	stateString := session.Get(r, "oauth-state-"+strings.ToLower(provider.GetName()))
	if stateString == "" {
		session.SetFlash(r, session.SessionBad, "An error occurred (1001)")
		return
	}

	state := oauth.State{}
	state.Unmarshal(stateString)

	page = state.Page

	//
	requestToken, verifier, err := oauth1.ParseAuthorizationCallback(r)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1001)")
		return
	}

	config := provider.GetConfig()
	accessToken, accessSecret, err := config.AccessToken(requestToken, state.State, verifier)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1003)")
		return
	}

	token := oauth1.NewToken(accessToken, accessSecret)

	// Api call to get user info
	resp, err := provider.GetUser(token)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "We were unable to fetch your details from "+provider.GetName()+" (1001)")
		return
	}

	if resp.Email == "" {
		session.SetFlash(r, session.SessionBad, "We were unable to fetch your details from "+provider.GetName()+" (1002)")
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

	//
	oauthHandleUser(provider, resp, user, page, r)
}

func providerOpenIDCallback(w http.ResponseWriter, r *http.Request, provider oauth.OpenIDProvider) {

	var page = r.URL.Query().Get("page")

	defer oauthRedirect(w, r, &page)

	// Api call to get user info
	resp, err := provider.GetUser(r)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "We were unable to fetch your details from "+provider.GetName()+" (1001)")
		return
	}

	if resp.ID == "" {
		session.SetFlash(r, session.SessionBad, "We were unable to fetch your details from "+provider.GetName()+" (1002)")
		return
	}

	var user mysql.User

	switch page {
	case authPageLogin:

		// Look for existing user by email
		i, err := strconv.Atoi(resp.ID)
		if err != nil {
			log.Err(err.Error(), zap.String("id", resp.ID))
			session.SetFlash(r, session.SessionBad, "We were unable to fetch your details from "+provider.GetName()+" (1002)")
			return
		}

		userProvider, err := mysql.GetUserProvider(provider.GetEnum(), i)
		if err != nil {
			log.ErrS(err)
			session.SetFlash(r, session.SessionGood, "Unable to find a Game DB account linked with this Steam account (1001)")
			return
		}

		user, err = mysql.GetUserByID(userProvider.UserID)
		if err == mysql.ErrRecordNotFound {
			session.SetFlash(r, session.SessionGood, "Unable to find a Game DB account linked with this Steam account (1002)")
			return
		} else if err != nil {
			log.ErrS(err)
			session.SetFlash(r, session.SessionGood, "An error occurred (1005)")
			return
		}

	case authPageSettings:

		userID, err := session.GetUserIDFromSesion(r)
		if err != nil {
			log.ErrS(err)
			session.SetFlash(r, session.SessionGood, "An error occurred (1005)")
			return
		}

		user, err = mysql.GetUserByID(userID)
		if err != nil {
			log.ErrS(err)
			session.SetFlash(r, session.SessionGood, "An error occurred (1005)")
			return
		}

	default:
		session.SetFlash(r, session.SessionGood, "Steam can't be used to sign up yet")
		return
	}

	//
	oauthHandleUser(provider, resp, user, page, r)

	//
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

func oauthRedirect(w http.ResponseWriter, r *http.Request, page *string) {

	session.Save(w, r)

	switch *page {
	case authPageLogin:
		http.Redirect(w, r, "/login", http.StatusFound)
	case authPageSignup:
		http.Redirect(w, r, "/signup", http.StatusFound)
	case authPageSettings:
		http.Redirect(w, r, "/settings", http.StatusFound)
	default:
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func oauthHandleUser(provider oauth.Provider, resp oauth.User, user mysql.User, page string, r *http.Request) {

	//
	switch page {
	case authPageLogin, authPageSignup:
		login(r, user)
	}

	// Check ID is not already in use
	used, err := mysql.CheckExistingUserProvider(provider.GetEnum(), resp.ID, user.ID)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1006)")
		return
	}

	if used {
		session.SetFlash(r, session.SessionBad, "This "+provider.GetName()+" account is already linked to another Game DB account")
		return
	}

	// Update provider in DB
	err = mysql.UpdateUserProvider(user.ID, provider.GetEnum(), resp)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1007)")
		return
	}

	// Success flash
	switch page {
	case authPageSettings:
		session.SetFlash(r, session.SessionGood, provider.GetName()+" account linked")
	case authPageSignup:
		session.SetFlash(r, session.SessionGood, "Account created")
	}

	// Create event
	err = mongo.NewEvent(r, user.ID, mongo.EventLink(provider.GetEnum()))
	if err != nil {
		log.ErrS(err)
	}
}
