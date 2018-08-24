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
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/session"
	"github.com/yohcop/openid-go"
	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	t := loginTemplate{}
	t.Fill(w, r, "Login")
	t.RecaptchaPublic = os.Getenv("STEAM_RECAPTCHA_PUBLIC")

	returnTemplate(w, r, "login", t)
	return
}

type loginTemplate struct {
	GlobalTemplate
	RecaptchaPublic string
}

func LoginPostHandler(w http.ResponseWriter, r *http.Request) {

	err := func() (err error) {

		var ErrInvalidCreds = errors.New("invalid username or password")
		var ErrInvalidCaptcha = errors.New("please check the captcha")

		// Parse form
		if err := r.ParseForm(); err != nil {
			return err
		}

		// Save email so they don't need to keep typing it
		session.Write(w, r, "login-email", r.PostForm.Get("email"))

		// Recaptcha
		err = recaptcha.CheckFromRequest(r)
		if err != nil {
			if err == recaptcha.ErrNotChecked {
				return ErrInvalidCaptcha
			} else {
				return err
			}
		}

		// Field validation
		email := r.PostForm.Get("email")
		password := r.PostForm.Get("password")

		if email == "" || password == "" {
			return ErrInvalidCreds
		}

		// Get users that match the email
		users, err := mysql.GetUsersByEmail(email)
		if err != nil {
			return err
		}

		if len(users) == 0 {
			return ErrInvalidCreds
		}

		// Check password matches
		var user mysql.User
		var success bool
		for _, user := range users {

			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err == nil {
				success = true
				break
			}
		}

		if !success {
			return ErrInvalidCreds
		}

		// Get player from user
		player, err := datastore.GetPlayer(user.PlayerID)
		if err != nil {
			return errors.New("no corresponding player")
		}

		// Log user in
		err = login(w, r, player)
		if err != nil {
			return err
		}

		// Remove form prefill on success
		session.Write(w, r, "login-email", "")

		return nil
	}()

	// Redirect
	if err != nil {
		logger.Error(err)
		session.SetGoodFlash(w, r, err.Error())
		http.Redirect(w, r, "/login", 302)
	} else {
		session.SetGoodFlash(w, r, "Login successful")
		http.Redirect(w, r, "/settings", 302)
	}

	return
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
	idInt, err := strconv.ParseInt(path.Base(openID), 10, 64)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Check if we have the player
	player, err := datastore.GetPlayer(idInt)

	// Get player if they're new
	if player.PersonaName == "" {
		errs := player.Update("")
		for _, v := range errs {
			logger.Error(v) // todo, Handle these better
		}
	}

	err = login(w, r, player)
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
		session.PlayerID:    strconv.FormatInt(player.PlayerID, 10),
		session.PlayerName:  player.PersonaName,
		session.PlayerLevel: strconv.Itoa(player.Level),
	})
	if err != nil {
		return err
	}

	// Create login record
	err = datastore.CreateEvent(r, player.PlayerID, datastore.EVENT_LOGIN)
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
