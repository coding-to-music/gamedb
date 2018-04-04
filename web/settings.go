package web

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/session"
	"github.com/steam-authority/steam-authority/steam"
	"github.com/yohcop/openid-go"
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

	loggedIn, err := session.IsLoggedIn(r)
	if err != nil {
		logger.Error(err)
		if err.Error() != "securecookie: the value is not valid" {
			returnErrorTemplate(w, r, 500, err.Error())
			return
		}
	}

	if loggedIn {
		http.Redirect(w, r, "/settings", 303)
		return
	}

	var url string
	url, err = openid.RedirectURL("http://steamcommunity.com/openid", os.Getenv("STEAM_DOMAIN")+"/login-callback", os.Getenv("STEAM_DOMAIN")+"/")
	if err != nil {
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
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	idString := path.Base(openID)

	idInt, err := strconv.Atoi(idString)
	if err != nil {
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
	}

	var gamesSlice []int
	gamesResp, err := steam.GetOwnedGames(idInt)

	for _, v := range gamesResp {
		gamesSlice = append(gamesSlice, v.AppID)
	}

	gamesString, err := json.Marshal(gamesSlice)
	if err != nil {
		logger.Error(err)
	}

	// Get level
	level, err := steam.GetSteamLevel(idInt)
	if err != nil {
		logger.Error(err)
	}

	// Save session
	session.WriteMany(w, r, map[string]string{
		session.ID:     idString,
		session.Name:   resp.PersonaName,
		session.Avatar: resp.AvatarMedium,
		session.Games:  string(gamesString),
		session.Level:  strconv.Itoa(level),
	})

	// Create login record
	datastore.CreateLogin(idInt, r)

	// Redirect
	http.Redirect(w, r, "/settings", 302)
	return
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	session.Clear(w, r)
	http.Redirect(w, r, "/", 303)
	return
}

func SettingsHandler(w http.ResponseWriter, r *http.Request) {

	loggedIn, err := session.IsLoggedIn(r)
	if err != nil {
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	if !loggedIn {
		http.Redirect(w, r, "/login", 302)
		return
	}

	// Get session
	id, err := session.Read(r, session.ID)
	if err != nil {
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Convert ID
	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Get logins
	logins, err := datastore.GetLogins(idx, 20)
	if err != nil {
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Template
	template := settingsTemplate{}
	template.Fill(r, "Settings")
	template.Logins = logins

	returnTemplate(w, r, "settings", template)

}

func SaveSettingsHandler(w http.ResponseWriter, r *http.Request) {

}

type settingsTemplate struct {
	GlobalTemplate
	User   datastore.Player
	Logins []datastore.Login
}
