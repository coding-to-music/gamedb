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
		return
	}
	if provider, ok := provider.(oauth.OAuth1Provider); ok {
		providerOAuth1Redirect(w, r, provider)
		return
	}
	if provider, ok := provider.(oauth.OpenIDProvider); ok {
		providerOpenIDRedirect(w, r, provider)
		return
	}

	Error404Handler(w, r)
}

func providerCallback(w http.ResponseWriter, r *http.Request) {

	provider := oauth.New(oauth.ProviderEnum(chi.URLParam(r, "provider")))
	if provider, ok := provider.(oauth.OAuth2Provider); ok {
		providerOAuth2Callback(w, r, provider)
		return
	}
	if provider, ok := provider.(oauth.OAuth1Provider); ok {
		providerOAuth1Callback(w, r, provider)
		return
	}
	if provider, ok := provider.(oauth.OpenIDProvider); ok {
		providerOpenIDCallback(w, r, provider)
		return
	}

	Error404Handler(w, r)
}

func providerOAuth2Redirect(w http.ResponseWriter, r *http.Request, provider oauth.OAuth2Provider) {

	state := helpers.RandString(32, helpers.LettersCaps)
	name := strings.ToLower(provider.GetName())
	page := r.URL.Query().Get("page")

	session.Set(r, "oauth-page-"+name, page)
	session.Set(r, "oauth-state-"+name, state)
	session.Save(w, r)

	provider.Redirect(w, r, state)
}

func providerOAuth1Redirect(w http.ResponseWriter, r *http.Request, provider oauth.OAuth1Provider) {

	name := strings.ToLower(provider.GetName())
	page := r.URL.Query().Get("page")

	redirect, secret, err := provider.Redirect()
	if err != nil {
		log.ErrS(err)
	}

	session.Set(r, "oauth-page-"+name, page)
	session.Set(r, "oauth-state-"+name, secret)
	session.Save(w, r)

	http.Redirect(w, r, redirect, http.StatusFound)
}

func providerOpenIDRedirect(w http.ResponseWriter, r *http.Request, provider oauth.OpenIDProvider) {

	name := strings.ToLower(provider.GetName())
	page := r.URL.Query().Get("page")

	session.Set(r, "oauth-page-"+name, page)
	session.Save(w, r)

	provider.Redirect(w, r, page)
}

func providerOAuth2Callback(w http.ResponseWriter, r *http.Request, provider oauth.OAuth2Provider) {

	var page string

	defer oauthRedirect(w, r, &page)

	// Get page
	page = session.Get(r, "oauth-page-"+strings.ToLower(provider.GetName()))
	if page == "" {
		session.SetFlash(r, session.SessionBad, "An error occurred (1001)")
		return
	}

	// Validate state
	stateOut := session.Get(r, "oauth-state-"+strings.ToLower(provider.GetName()))
	if stateOut == "" {
		session.SetFlash(r, session.SessionBad, "An error occurred (1002)")
		return
	}

	stateIn := r.URL.Query().Get("state")
	if stateIn == "" {
		session.SetFlash(r, session.SessionBad, "Invalid state (1003)")
		return
	}

	if stateOut != stateIn {
		session.SetFlash(r, session.SessionBad, "Invalid state (1004)")
		return
	}

	// Swap code for token
	code := r.URL.Query().Get("code")
	if code == "" {
		session.SetFlash(r, session.SessionBad, "Invalid code (1005)")
		return
	}

	conf := provider.GetConfig()
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1006)")
		return
	}

	// Get user
	resp, err := provider.GetUser(token)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "We were unable to fetch your details from "+provider.GetName()+" (1007)")
		return
	}

	//
	oauthHandleUser(provider, resp, page, r)
}

func providerOAuth1Callback(w http.ResponseWriter, r *http.Request, provider oauth.OAuth1Provider) {

	var page string

	defer oauthRedirect(w, r, &page)

	// Get page
	page = session.Get(r, "oauth-page-"+strings.ToLower(provider.GetName()))
	if page == "" {
		session.SetFlash(r, session.SessionBad, "An error occurred (1001)")
		return
	}

	// Validate state
	state := session.Get(r, "oauth-state-"+strings.ToLower(provider.GetName()))
	if state == "" {
		session.SetFlash(r, session.SessionBad, "An error occurred (1002)")
		return
	}

	//
	requestToken, verifier, err := oauth1.ParseAuthorizationCallback(r)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1003)")
		return
	}

	config := provider.GetConfig()
	accessToken, accessSecret, err := config.AccessToken(requestToken, state, verifier)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1004)")
		return
	}

	token := oauth1.NewToken(accessToken, accessSecret)

	// Get user
	resp, err := provider.GetUser(token)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "We were unable to fetch your details from "+provider.GetName()+" (1005)")
		return
	}

	//
	oauthHandleUser(provider, resp, page, r)
}

func providerOpenIDCallback(w http.ResponseWriter, r *http.Request, provider oauth.OpenIDProvider) {

	var page string

	defer oauthRedirect(w, r, &page)

	// Get page
	page = session.Get(r, "oauth-page-"+strings.ToLower(provider.GetName()))
	if page == "" {
		session.SetFlash(r, session.SessionBad, "An error occurred (1001)")
		return
	}

	// Get user
	resp, err := provider.GetUser(r)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "We were unable to fetch your details from "+provider.GetName()+" (1002)")
		return
	}

	if resp.ID == "" {
		session.SetFlash(r, session.SessionBad, "We were unable to fetch your details from "+provider.GetName()+" (1003)")
		return
	}

	//
	oauthHandleUser(provider, resp, page, r)
}

func oauthRedirect(w http.ResponseWriter, r *http.Request, page *string) {

	// Save for flashes
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

func oauthHandleUser(provider oauth.Provider, resp oauth.User, page string, r *http.Request) {

	// Check page is valid
	switch page {
	case authPageLogin, authPageSignup, authPageSettings:
	default:
		session.SetFlash(r, session.SessionBad, "Invalid page (1101)")
		return
	}

	//
	if page == authPageSignup && !provider.HasEmail() {

		session.SetFlash(r, session.SessionBad, provider.GetName()+" currently can't be used to sign up (1102)")
		return
	}

	//
	if (provider.HasEmail() && resp.Email == "") || (resp.ID == "") {

		session.SetFlash(r, session.SessionBad, "We were unable to fetch your details from "+provider.GetName()+" (1103)")
		return
	}

	// Get the user from DB
	var user mysql.User
	var err error

	if page == authPageSettings {

		// Just get user from session
		user, err = getUserFromSession(r)
		if err != nil {
			log.ErrS(err)
			session.SetFlash(r, session.SessionBad, "Account could not be created (1104)")
			return
		}

	} else if provider.GetType() == oauth.TypeOAuth {

		// Look for existing user by email
		user, err = mysql.GetUserByEmail(resp.Email)
		if err == mysql.ErrRecordNotFound {

			// Create new user
			user, err = mysql.NewUser(r, resp.Email, "", session.GetProductCC(r), true)
			if err != nil {
				log.ErrS(err)
				session.SetFlash(r, session.SessionBad, "Account could not be created (1105)")
				return
			}

		} else if err != nil {
			log.ErrS(err)
			session.SetFlash(r, session.SessionBad, "An error occurred (1106)")
			return
		}

	} else if provider.GetType() == oauth.TypeOpenID {

		// Look for existing user by email
		userProvider, err := mysql.GetUserProviderByProviderID(provider.GetEnum(), resp.ID)
		if err != nil {
			err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
			if err != nil {
				log.ErrS(err)
			}
			session.SetFlash(r, session.SessionBad, "Unable to find a Game DB account linked with this "+provider.GetName()+" account (1107)")
			return
		}

		user, err = mysql.GetUserByID(userProvider.UserID)
		if err == mysql.ErrRecordNotFound {
			session.SetFlash(r, session.SessionBad, "Unable to find a Game DB account linked with this "+provider.GetName()+" account (1108)")
			return
		} else if err != nil {
			log.ErrS(err)
			session.SetFlash(r, session.SessionBad, "An error occurred (1109)")
			return
		}
	}

	//
	switch page {
	case authPageLogin, authPageSignup:
		login(r, user)
	}

	// Check ID is not already in use
	used, err := mysql.CheckExistingUserProvider(provider.GetEnum(), resp.ID, user.ID)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1110)")
		return
	}

	if used {
		session.SetFlash(r, session.SessionBad, "This "+provider.GetName()+" account ("+resp.Username+") is already linked to another Game DB account")
		return
	}

	// Update provider in DB
	err = mysql.UpdateUserProvider(user.ID, provider.GetEnum(), resp)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1111)")
		return
	}

	// Success flash
	switch page {
	case authPageSettings:
		session.SetFlash(r, session.SessionGood, provider.GetName()+" account linked")
	case authPageSignup:
		session.SetFlash(r, session.SessionGood, "Account created with "+provider.GetName())
	}

	// Create event
	switch page {
	case authPageSettings, authPageSignup:

		err = mongo.NewEvent(r, user.ID, mongo.EventLink(provider.GetEnum()))
		if err != nil {
			log.ErrS(err)
		}
	}

	// Queue player
	if provider.GetEnum() == oauth.ProviderSteam {

		i, err := strconv.ParseInt(resp.ID, 10, 64)
		if err == nil {

			ua := r.UserAgent()
			err = queue.ProducePlayer(queue.PlayerMessage{ID: i, UserAgent: &ua})
			if err == nil {
				log.Info("player queued", zap.String("ua", ua))
			}
			err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
			if err != nil {
				log.ErrS(err)
			}
		}
	}
}
