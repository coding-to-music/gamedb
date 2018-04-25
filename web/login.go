package web

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/recaptcha"
	"github.com/steam-authority/steam-authority/session"
	"github.com/steam-authority/steam-authority/steam"
	"github.com/yohcop/openid-go"
	"golang.org/x/crypto/bcrypt"
)

// todo
// For the demo, we use in-memory infinite storage nonce and discovery
// cache. In your app, do not use this as it will eat up memory and never
// free it. Use your own implementation, on a better database system.
// If you have multiple servers for example, you may need to share at least
// the nonceStore between them.
var nonceStore = openid.NewSimpleNonceStore()
var discoveryCache = openid.NewSimpleDiscoveryCache()

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	var success bool

	if r.Method == "POST" {

		// Form validation
		if err := r.ParseForm(); err != nil {
			logger.Error(err)
			returnErrorTemplate(w, r, 500, err.Error())
			return
		}

		// Recaptcha
		var success bool
		var err error

		response := r.PostForm.Get("g-recaptcha-response")
		if response != "" {

			success, err = recaptcha.Check(os.Getenv("STEAM_RECAPTCHA_PRIVATE"), response, r.RemoteAddr)
			if err != nil {
				if err != recaptcha.ErrInvalidInputs {
					logger.Error(err)
				}
			}
		}

		if !success {
			returnErrorTemplate(w, r, 401, "Please check the captcha")
			return
		}

		// Field validation
		email := r.PostForm.Get("email")
		password := r.PostForm.Get("password")

		if email == "" || password == "" {
			returnErrorTemplate(w, r, 401, "Please fill in your username and password")
			return
		}

		players, err := datastore.GetPlayersByEmail(email)
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				returnErrorTemplate(w, r, 401, "Invalid credentials")
				return
			} else {
				logger.Error(err)
				returnErrorTemplate(w, r, 500, err.Error())
				return
			}
		}

		for _, v := range players {

			err = bcrypt.CompareHashAndPassword([]byte(v.SettingsPassword), []byte(password))
			if err == nil {

				// todo, do login
				fmt.Println("logging in")
				success = true
				break
			}
		}
	}

	t := loginTemplate{}
	t.Fill(r, "Login")
	t.Message = "xx"
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

func LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {

	// todo, get session data from db not steam

	openID, err := openid.Verify(os.Getenv("STEAM_DOMAIN")+r.URL.String(), discoveryCache, nonceStore)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	idString := path.Base(openID)

	idInt, err := strconv.Atoi(idString)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Set session from steam
	resp, err := steam.GetPlayerSummaries(idInt)
	if err != nil {
		if !strings.HasPrefix(err.Error(), "not found in steam") {
			returnErrorTemplate(w, r, 500, err.Error())
			return
		}

		logger.Error(err)
	}

	var gamesSlice []int
	gamesResp, err := steam.GetOwnedGames(idInt)

	for _, v := range gamesResp {
		gamesSlice = append(gamesSlice, v.AppID)
	}

	//gamesString, err := json.Marshal(gamesSlice)
	//if err != nil {
	//	logger.Error(err)
	//}

	// Get level
	level, err := steam.GetSteamLevel(idInt)
	if err != nil {
		logger.Error(err)
	}

	// Save session
	err = session.WriteMany(w, r, map[string]string{
		session.ID:     idString,
		session.Name:   resp.PersonaName,
		session.Avatar: resp.AvatarMedium,
		//session.Games:  string(gamesString),
		session.Level: strconv.Itoa(level),
	})
	if err != nil {
		logger.Error(err)
	}

	// Create login record
	datastore.CreateLogin(idInt, r)

	// Redirect
	http.Redirect(w, r, "/settings", 302)
	return
}

func login(player datastore.Player) {

}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	session.Clear(w, r)
	http.Redirect(w, r, "/", 303)
	return
}
