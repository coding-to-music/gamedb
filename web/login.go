package web

import (
	"errors"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/recaptcha"
	"github.com/steam-authority/steam-authority/session"
	"github.com/yohcop/openid-go"
	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {

		var ErrInvalidCreds = errors.New("invalid username or password")
		var ErrInvalidCaptcha = errors.New("please check the captcha")

		success, code, err := func() (success bool, code int, err error) {

			// Parse form
			if err := r.ParseForm(); err != nil {
				return false, 500, err
			}

			// Recaptcha
			success, err = recaptcha.CheckFromRequest(r)
			if err != nil {
				return false, 500, err
			}

			if !success {
				return false, 401, ErrInvalidCaptcha
			}

			// Field validation
			email := r.PostForm.Get("email")
			password := r.PostForm.Get("password")

			if email == "" || password == "" {
				return false, 401, ErrInvalidCreds
			}

			players, err := datastore.GetPlayersByEmail(email)
			if err != nil {
				if err == datastore.ErrNoSuchEntity {
					return false, 401, ErrInvalidCreds
				} else {
					return false, 500, err
				}
			}

			if len(players) == 0 {
				return false, 401, ErrInvalidCreds
			}

			var player datastore.Player

			for _, player := range players {

				err = bcrypt.CompareHashAndPassword([]byte(player.SettingsPassword), []byte(password))
				if err == nil {
					success = true
					break
				}
			}

			if success {

				// Save session
				err = session.WriteMany(w, r, map[string]string{
					session.UserID:    strconv.Itoa(player.PlayerID),
					session.UserName:  player.PersonaName,
					session.UserLevel: strconv.Itoa(player.Level),
				})
				if err != nil {
					logger.Error(err)
					returnErrorTemplate(w, r, 500, err.Error())
					return
				}

				// Create login record
				err = datastore.CreateLogin(player.PlayerID, r)
				if err != nil {
					logger.Error(err)
					returnErrorTemplate(w, r, 500, err.Error())
					return
				}
			}

		}()

		if err == ErrInvalidCreds || err == ErrInvalidCaptcha {
			code = 401
		}
	}

	t := loginTemplate{}
	t.Fill(r, "Login")
	t.Message = message
	t.Success = success
	t.RecaptchaPublic = os.Getenv("STEAM_RECAPTCHA_PUBLIC")

	returnTemplate(w, r, "login", t)
	return

}

type loginTemplate struct {
	GlobalTemplate
	Username        string
	Message         string
	Success         bool
	RecaptchaPublic string
}

func LoginOpenIDHandler(w http.ResponseWriter, r *http.Request) {

	loggedIn, err := session.IsLoggedIn(r)
	if err != nil {
		logger.Error(err)
	}

	if loggedIn {
		http.Redirect(w, r, "/settings", 303)
		return
	}

	var url string
	url, err = openid.RedirectURL("http://steamcommunity.com/openid", os.Getenv("STEAM_DOMAIN")+"/login/callback", os.Getenv("STEAM_DOMAIN")+"/")
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	http.Redirect(w, r, url, 303)
	return
}

// todo
// For the demo, we use in-memory infinite storage nonce and discovery
// cache. In your app, do not use this as it will eat up memory and never
// free it. Use your own implementation, on a better database system.
// If you have multiple servers for example, you may need to share at least
// the nonceStore between them.
var nonceStore = openid.NewSimpleNonceStore()
var discoveryCache = openid.NewSimpleDiscoveryCache()

func LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {

	// Get ID from OpenID
	openID, err := openid.Verify(os.Getenv("STEAM_DOMAIN")+r.URL.String(), discoveryCache, nonceStore)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Convert to int
	idInt, err := strconv.Atoi(path.Base(openID))
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Check if we have the player
	player, err := datastore.GetPlayer(idInt)
	if player.PlayerID == 0 {
		errs := player.Update("")
	}

	// Save session
	err = session.WriteMany(w, r, map[string]string{
		session.UserID:    strconv.Itoa(player.PlayerID),
		session.UserName:  player.PersonaName,
		session.UserLevel: strconv.Itoa(player.Level),
	})
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Create login record
	err = datastore.CreateLogin(player.PlayerID, r)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Redirect
	http.Redirect(w, r, "/settings", 302)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	session.Clear(w, r)
	http.Redirect(w, r, "/", 303)
	return
}
