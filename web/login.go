package web

import (
	"errors"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/Jleagle/recaptcha-go"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/session"
	"github.com/yohcop/openid-go"
	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	if r.Method == "POST" {

		var ErrInvalidCreds = errors.New("invalid username or password")
		var ErrInvalidCaptcha = errors.New("please check the captcha")

		err = func() (err error) {

			// Parse form
			if err := r.ParseForm(); err != nil {
				return err
			}

			// Recaptcha
			err = recaptcha.CheckFromRequest(r)
			if err != nil {
				if err == recaptcha.ErrNotChecked {
					return ErrInvalidCaptcha
				} else {
					logger.Error(err)
					return err
				}
			}

			// Field validation
			email := r.PostForm.Get("email")
			password := r.PostForm.Get("password")

			if email == "" || password == "" {
				return ErrInvalidCreds
			}

			// Get players that match the email
			players, err := datastore.GetPlayersByEmail(email)
			if err != nil {
				if err == datastore.ErrNoSuchEntity {
					return ErrInvalidCreds
				} else {
					return err
				}
			}

			if len(players) == 0 {
				return ErrInvalidCreds
			}

			var player datastore.Player
			var success bool
			for _, v := range players {

				err = bcrypt.CompareHashAndPassword([]byte(v.SettingsPassword), []byte(password))
				if err == nil {
					success = true
					player = v
					break
				}
			}

			if success {

				err = login(w, r, player)
				if err != nil {
					return err
				}

				return nil
			} else {
				return ErrInvalidCreds
			}
		}()

		// Redirect
		if err == nil {
			session.SetGoodFlash(w, r, "Login successful")
			http.Redirect(w, r, "/settings", 302)
			return
		}
	}

	t := loginTemplate{}
	t.Fill(w, r, "Login")
	t.Success = err == nil
	if err != nil {
		t.Message = err.Error()
	}
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
		for _, v := range errs {
			logger.Error(v) // Handle these better
		}
	}

	err = login(w, r, *player)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Redirect
	http.Redirect(w, r, "/settings", 302)
	return
}

func login(w http.ResponseWriter, r *http.Request, player datastore.Player) (err error) {

	// Save session
	err = session.WriteMany(w, r, map[string]string{
		session.UserID:    strconv.Itoa(player.PlayerID),
		session.UserName:  player.PersonaName,
		session.UserLevel: strconv.Itoa(player.Level),
	})
	if err != nil {
		return err
	}

	// Create login record
	err = datastore.CreateLogin(player.PlayerID, r)
	if err != nil {
		return err
	}

	return nil
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	session.Clear(w, r)
	http.Redirect(w, r, "/", 303)
	return
}
