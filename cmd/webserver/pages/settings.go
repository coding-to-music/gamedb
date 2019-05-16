package pages

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"sync"

	"github.com/Jleagle/session-go/session"
	"github.com/Jleagle/steam-go/steam"
	"github.com/badoux/checkmail"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
	"github.com/mxpv/patreon-go"
	"github.com/yohcop/openid-go"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func SettingsRouter() http.Handler {

	r := chi.NewRouter()
	r.Use(middlewareAuthCheck())

	r.Get("/", settingsHandler)
	r.Post("/update", settingsPostHandler)
	r.Post("/delete", deletePostHandler)
	r.Get("/events.json", settingsEventsAjaxHandler)
	r.Get("/donations.json", settingsDonationsAjaxHandler)

	// r.Get("/link-steam", linkSteamHandler)
	r.Get("/steam-callback", linkSteamCallbackHandler)
	r.Get("/unlink-steam", unlinkSteamHandler)

	r.Get("/link-patreon", linkPatreonHandler)
	r.Get("/unlink-patreon", unlinkPatreonHandler)
	r.Get("/patreon-callback", linkPatreonCallbackHandler)

	r.Get("/link-google", linkGoogleHandler)
	r.Get("/unlink-google", unlinkGoogleHandler)
	r.Get("/google-callback", linkGoogleCallbackHandler)

	r.Get("/link-discord", linkDiscordHandler)
	r.Get("/unlink-discord", unlinkDiscordHandler)
	r.Get("/discord-callback", linkDiscordCallbackHandler)

	return r
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	//
	t := settingsTemplate{}
	t.fill(w, r, "Settings", "")
	t.addAssetPasswordStrength()
	t.setFlashes(w, r, false)
	t.Domain = config.Config.GameDBDomain.Get()

	// Get user
	t.User, err = getUserFromSession(r)
	log.Err(err)

	// Set Steam name to session if missing, can happen after linking
	if t.User.SteamID != 0 {

		name, err := session.Get(r, helpers.SessionPlayerName)
		log.Err(err)

		if name == "" && err == nil {

			t.Player, err = mongo.GetPlayer(t.User.SteamID)
			err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
			log.Err(err)

			if t.Player.PersonaName != "" {
				err = session.Set(r, helpers.SessionPlayerName, t.Player.VanintyURL)
				log.Err(err)
			}
		}
	}

	err = session.Save(w, r)
	log.Err(err)

	//
	var wg sync.WaitGroup

	// Get games
	wg.Add(1)
	go func() {

		defer wg.Done()

		if t.User.SteamID == 0 {
			return
		}

		playerApps, err := mongo.GetPlayerApps(t.User.SteamID, 0, 0, mongo.D{})
		if err != nil {
			log.Err(err, r)
			return
		}

		var appIDs []int
		for _, v := range playerApps {
			appIDs = append(appIDs, v.AppID)
		}

		t.Games = string(helpers.MarshalLog(appIDs))

	}()

	// Wait
	wg.Wait()

	// Countries
	for _, v := range helpers.GetActiveCountries() {
		t.Countries = append(t.Countries, []string{string(v), steam.Countries[v]})
	}

	// Template
	err = returnTemplate(w, r, "settings", t)
	log.Err(err, r)
}

type settingsTemplate struct {
	GlobalTemplate
	User      sql.User
	Player    mongo.Player
	Games     string
	Countries [][]string
	Domain    string
}

func deletePostHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	redirect, good, bad := func() (redirect string, good string, bad string) {

		// Parse form
		err = r.ParseForm()
		if err != nil {
			log.Err(err)
			return "/settings", "", "There was an eror saving your information."
		}

		user, err := getUserFromSession(r)
		if err != nil {
			return "/settings", "", "There was an eror saving your information."
		}

		if r.PostForm.Get("id") == strconv.FormatInt(user.SteamID, 10) {

			err = session.DeleteAll(r)
			log.Err(err)
			return "/", "Your account has been deleted", ""

		} else {
			return "/settings", "", "Invalid player ID."
		}
	}()

	if good != "" {
		err = session.SetFlash(r, helpers.SessionGood, good)
		log.Err(err)
	}
	if bad != "" {
		err = session.SetFlash(r, helpers.SessionBad, bad)
		log.Err(err)
	}

	err = session.Save(w, r)
	log.Err(err)

	http.Redirect(w, r, redirect, http.StatusFound)
}

func settingsPostHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	redirect, good, bad := func() (redirect string, good string, bad string) {

		// Get user
		user, err := getUserFromSession(r)
		log.Err(err)
		if err != nil {
			return "/settings", "", "User not found"
		}

		// Parse form
		err = r.ParseForm()
		log.Err(err)
		if err != nil {
			return "/settings", "", "Could not read form data"
		}

		email := r.PostForm.Get("email")
		password := r.PostForm.Get("password")
		country := r.PostForm.Get("country_code")

		// Email
		if email != "" && email != user.Email {

			err = checkmail.ValidateFormat(r.PostForm.Get("email"))
			if err != nil {
				return "/settings", "", "Invalid email address"
			}

			user.Email = r.PostForm.Get("email")
		}

		// Password
		if email != user.Email {
			user.EmailVerified = false
		}

		if password != "" {

			if len(password) < 8 {
				return "/settings", "", "Password must be at least 8 characters long"
			}

			passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
			log.Err(err, r)
			if err != nil {
				return "/settings", "", "Something went wrong encrypting your password"
			}

			user.Password = string(passwordBytes)
		}

		// Country code
		if _, ok := steam.Countries[steam.CountryCode(country)]; ok {
			user.CountryCode = country
		} else {
			user.CountryCode = string(steam.CountryUS)
		}

		// Save hidden
		if r.PostForm.Get("hide") == "1" {
			user.HideProfile = true
		} else {
			user.HideProfile = false
		}

		// Save alerts
		if r.PostForm.Get("alerts") == "1" {
			user.ShowAlerts = true
		} else {
			user.ShowAlerts = false
		}

		// Save user
		db, err := sql.GetMySQLClient()
		log.Err(err)
		if err != nil {
			return "/settings", "", "We had trouble saving your settings"
		}

		// Have to save as a map because gorm does not save empty values otherwise
		db = db.Model(&user).Updates(map[string]interface{}{
			"email":        user.Email,
			"verified":     user.EmailVerified,
			"password":     user.Password,
			"hide_profile": user.HideProfile,
			"show_alerts":  user.ShowAlerts,
			"country_code": user.CountryCode,
		})

		log.Err(db.Error, r)
		if db.Error != nil {
			return "/settings", "", "Something went wrong saving your settings"
		}

		// Update session
		err = session.SetMany(r, map[string]string{
			helpers.SessionUserCountry:    user.CountryCode,
			helpers.SessionUserEmail:      user.Email,
			helpers.SessionUserShowAlerts: strconv.FormatBool(user.ShowAlerts),
		})
		if err != nil {
			log.Err(err, r)
			return "/settings", "", "Something went wrong saving your settings"
		}

		return "/settings", "Settings saved", ""
	}()

	if good != "" {
		err = session.SetFlash(r, helpers.SessionGood, good)
		log.Err(err)
	}
	if bad != "" {
		err = session.SetFlash(r, helpers.SessionBad, bad)
		log.Err(err)
	}

	err = session.Save(w, r)
	log.Err(err)

	http.Redirect(w, r, redirect, http.StatusFound)
}

func settingsEventsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	user, err := getUserFromSession(r)
	if err != nil {
		log.Err(err)
		return
	}

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	var wg sync.WaitGroup

	// Get events
	var events []mongo.Event
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		events, err = mongo.GetEvents(user.ID, query.getOffset64())
		if err != nil {
			log.Err(err, r)
			return
		}

	}(r)

	// Get total
	var total int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		total, err = mongo.CountEvents(user.ID)
		log.Err(err, r)

	}(r)

	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.FormatInt(total, 10)
	response.RecordsFiltered = response.RecordsTotal
	response.Draw = query.Draw

	for _, v := range events {
		response.AddRow(v.OutputForJSON(r.RemoteAddr))
	}

	response.output(w, r)
}

func settingsDonationsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	user, err := getUserFromSession(r)
	if err != nil {
		log.Err(err)
		return
	}

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	var wg sync.WaitGroup

	// Get events
	var events []mongo.Event
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

	}(r)

	// Get total
	var total int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		total, err = mongo.CountPatreonWebhooks(user.ID)
		log.Err(err, r)

	}(r)

	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.FormatInt(total, 10)
	response.RecordsFiltered = response.RecordsTotal
	response.Draw = query.Draw

	for _, v := range events {
		response.AddRow(v.OutputForJSON(r.RemoteAddr))
	}

	response.output(w, r)
}

// todo
// Steam
// For the demo, we use in-memory infinite storage nonce and discovery
// cache. In your app, do not use this as it will eat up memory and never
// free it. Use your own implementation, on a better database system.
// If you have multiple servers for example, you may need to share at least
// the nonceStore between them.
var nonceStore = openid.NewSimpleNonceStore()
var discoveryCache = openid.NewSimpleDiscoveryCache()

func linkSteamCallbackHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		err := session.Save(w, r)
		log.Err(err)

		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	// Get Steam ID
	openID, err := openid.Verify(config.Config.GameDBDomain.Get()+r.URL.String(), discoveryCache, nonceStore)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "We could not verify your Steam account")
		log.Err(err)
		return
	}

	steamIDString := path.Base(openID)
	steamID, err := strconv.ParseInt(steamIDString, 10, 64)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1001)")
		log.Err(err)
		return
	}

	user, err := getUserFromSession(r)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1004)")
		log.Err(err)
		return
	}

	// Check Steam ID not already in use
	_, err = sql.GetUserBySteamID(steamID, user.ID)
	if err == nil {
		err = session.SetFlash(r, helpers.SessionBad, "This Steam account is already linked to another Game DB account")
		log.Err(err)
		return
	} else if err != sql.ErrRecordNotFound {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1002)")
		log.Err(err)
		return
	}

	// Update user
	err = sql.UpdateUserCol(user.ID, "steam_id", steamID)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1003)")
		log.Err(err)
		return
	}

	// Success flash
	err = session.SetFlash(r, helpers.SessionGood, "Steam account linked")
	log.Err(err)

	// Create event
	err = mongo.CreateUserEvent(r, user.ID, mongo.EventLinkSteam)
	if err != nil {
		log.Err(err, r)
	}

	// Queue for an update
	player, err := mongo.GetPlayer(steamID)
	if err != nil {
		err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
		log.Err(err)
	} else {
		if player.ShouldUpdate(r.UserAgent(), mongo.PlayerUpdateManual) {
			err = queue.ProducePlayer(player.ID)
			log.Err(err, r)

			// Queued flash
			err = session.SetFlash(r, helpers.SessionGood, "Player has been queued for an update")
			log.Err(err)
		}
	}

	// Update session
	err = session.Set(r, helpers.SessionPlayerID, steamIDString)
	log.Err(err)
}

func unlinkSteamHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		err := session.Save(w, r)
		log.Err(err)

		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	userID, err := getUserIDFromSesion(r)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1001)")
		log.Err(err)
		return
	}

	// Update user
	err = sql.UpdateUserCol(userID, "steam_id", 0)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1002)")
		log.Err(err)
		return
	}

	// Clear session
	err = session.DeleteMany(r, []string{helpers.SessionPlayerID, helpers.SessionPlayerName, helpers.SessionPlayerLevel})
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1003)")
		log.Err(err)
		return
	}

	// Create event
	err = mongo.CreateUserEvent(r, userID, mongo.EventUnlinkSteam)
	if err != nil {
		log.Err(err, r)
	}

	// Flash message
	err = session.SetFlash(r, helpers.SessionGood, "Steam unlinked")
	log.Err(err)
}

// Patreon
var (
	patreonConfig = oauth2.Config{
		ClientID:     config.Config.PatreonClientID.Get(),
		ClientSecret: config.Config.PatreonClientSecret.Get(),
		Scopes:       []string{"identity", "identity[email]"}, // identity[email] scope is only needed as the Patreon package we are using only handles v1 API
		RedirectURL:  config.Config.GameDBDomain.Get() + "/settings/patreon-callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  patreon.AuthorizationURL,
			TokenURL: patreon.AccessTokenURL,
		},
	}
)

func linkPatreonHandler(w http.ResponseWriter, r *http.Request) {

	state := helpers.RandString(5, helpers.Numbers)

	err := session.Set(r, "patreon-oauth-state", state)
	log.Err(err)

	err = session.Save(w, r)
	log.Err(err)

	url := patreonConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func linkPatreonCallbackHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {

		err := session.Save(w, r)
		log.Err(err)

		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	// Oauth checks
	realState, err := session.Get(r, "patreon-oauth-state")
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1001)")
		log.Err(err)
		return
	}

	err = r.ParseForm()
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1002)")
		log.Err(err)
		return
	}

	state := r.Form.Get("state")
	if state != realState {
		err = session.SetFlash(r, helpers.SessionBad, "Invalid state")
		log.Err(err)
		return
	}

	code := r.Form.Get("code")
	if code == "" {
		err = session.SetFlash(r, helpers.SessionBad, "Invalid code")
		log.Err(err)
		return
	}

	// Get token
	token, err := patreonConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "Invalid token")
		log.Err(err)
		return
	}

	// Get Patreon user
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token.AccessToken})
	tc := oauth2.NewClient(context.TODO(), ts)

	patreonUser, err := patreon.NewClient(tc).FetchUser()
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1003)")
		log.Err(err)
		return
	}

	idx, err := strconv.Atoi(patreonUser.Data.ID)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1004)")
		log.Err(err)
		return
	}

	// Get user
	user, err := getUserFromSession(r)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1005)")
		log.Err(err)
		return
	}

	// Check Steam ID not already in use
	_, err = sql.GetUserByPatreonID(idx, user.ID)
	if err == nil {
		err = session.SetFlash(r, helpers.SessionBad, "This Patreon account is already linked to another Game DB account")
		log.Err(err)
		return
	} else if err != sql.ErrRecordNotFound {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1006)")
		log.Err(err)
		return
	}

	// Update user
	err = sql.UpdateUserCol(user.ID, "patreon_id", idx)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1007)")
		log.Err(err)
		return
	}

	// Create event
	err = mongo.CreateUserEvent(r, user.ID, mongo.EventLinkPatreon)
	if err != nil {
		log.Err(err, r)
	}

	// Flash message
	err = session.SetFlash(r, helpers.SessionGood, "Patreon account linked")
	log.Err(err)
}

func unlinkPatreonHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		err := session.Save(w, r)
		log.Err(err)

		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	userID, err := getUserIDFromSesion(r)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1001)")
		log.Err(err)
		return
	}

	// Update user
	err = sql.UpdateUserCol(userID, "patreon_id", 0)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1002)")
		log.Err(err)
		return
	}

	// Flash message
	err = session.SetFlash(r, helpers.SessionGood, "Patreon unlinked")
	log.Err(err, r)

	// Create event
	err = mongo.CreateUserEvent(r, userID, mongo.EventUnlinkPatreon)
	if err != nil {
		log.Err(err, r)
	}
}

// Google
var (
	googleConfig = oauth2.Config{
		ClientID:     config.Config.GoogleOauthClientID.Get(),
		ClientSecret: config.Config.GoogleOauthClientSecret.Get(),
		Scopes:       []string{"profile"},
		RedirectURL:  config.Config.GameDBDomain.Get() + "/settings/google-callback",
		Endpoint:     google.Endpoint,
	}
)

func linkGoogleHandler(w http.ResponseWriter, r *http.Request) {

	state := helpers.RandString(5, helpers.Numbers)

	err := session.Set(r, "google-oauth-state", state)
	log.Err(err)

	err = session.Save(w, r)
	log.Err(err)

	url := googleConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func unlinkGoogleHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		err := session.Save(w, r)
		log.Err(err)

		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	userID, err := getUserIDFromSesion(r)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1001)")
		log.Err(err)
		return
	}

	// Update user
	err = sql.UpdateUserCol(userID, "google_id", 0)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1002)")
		log.Err(err)
		return
	}

	// Create event
	err = mongo.CreateUserEvent(r, userID, mongo.EventUnlinkGoogle)
	if err != nil {
		log.Err(err, r)
	}

	// Flash message
	err = session.SetFlash(r, helpers.SessionGood, "Google unlinked")
	log.Err(err)

}

func linkGoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {

		err := session.Save(w, r)
		log.Err(err)

		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	// Oauth checks
	realState, err := session.Get(r, "google-oauth-state")
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1001)")
		log.Err(err)
		return
	}

	err = r.ParseForm()
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1002)")
		log.Err(err)
		return
	}

	state := r.Form.Get("state")
	if state != realState {
		err = session.SetFlash(r, helpers.SessionBad, "Invalid state")
		log.Err(err)
		return
	}

	code := r.Form.Get("code")
	if code == "" {
		err = session.SetFlash(r, helpers.SessionBad, "Invalid code")
		log.Err(err)
		return
	}

	// Get token
	token, err := googleConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "Invalid token")
		log.Err(err)
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "Invalid token")
		log.Err(err)
		return
	}
	defer func(response *http.Response) {
		err := response.Body.Close()
		log.Err(err)
	}(response)

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1004)")
		log.Err(err)
		return
	}

	userInfo := struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
		Picture    string `json:"picture"`
		Locale     string `json:"locale"`
	}{}

	err = json.Unmarshal(b, &userInfo)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1005)")
		log.Err(err)
		return
	}

	// Get user
	user, err := getUserFromSession(r)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1007)")
		log.Err(err)
		return
	}

	// Check Steam ID not already in use
	_, err = sql.GetUserByGoogleID(userInfo.ID, user.ID)
	if err == nil {
		err = session.SetFlash(r, helpers.SessionBad, "This Google account is already linked to another Game DB account")
		log.Err(err)
		return
	} else if err != sql.ErrRecordNotFound {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1008)")
		log.Err(err)
		return
	}

	// Update user
	err = sql.UpdateUserCol(user.ID, "google_id", userInfo.ID)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1009)")
		log.Err(err)
		return
	}

	// Create event
	err = mongo.CreateUserEvent(r, user.ID, mongo.EventLinkGoogle)
	if err != nil {
		log.Err(err, r)
	}

	// Flash message
	err = session.SetFlash(r, helpers.SessionGood, "Google account linked")
	log.Err(err)
}

// Discord
var (
	discordConfig = oauth2.Config{
		ClientID:     config.Config.DiscordClientID.Get(),
		ClientSecret: config.Config.DiscordClientSescret.Get(),
		Scopes:       []string{"identify"},
		RedirectURL:  config.Config.GameDBDomain.Get() + "/settings/discord-callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discordapp.com/api/oauth2/authorize",
			TokenURL: "https://discordapp.com/api/oauth2/token",
		},
	}
)

func linkDiscordHandler(w http.ResponseWriter, r *http.Request) {

	state := helpers.RandString(5, helpers.Numbers)

	err := session.Set(r, "discord-oauth-state", state)
	log.Err(err)

	err = session.Save(w, r)
	log.Err(err)

	url := discordConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func unlinkDiscordHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		err := session.Save(w, r)
		log.Err(err)

		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	userID, err := getUserIDFromSesion(r)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1001)")
		log.Err(err)
		return
	}

	// Update user
	err = sql.UpdateUserCol(userID, "discord_id", 0)
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1002)")
		log.Err(err)
		return
	}

	// Create event
	err = mongo.CreateUserEvent(r, userID, mongo.EventUnlinkDiscord)
	if err != nil {
		log.Err(err, r)
	}

	// Flash message
	err = session.SetFlash(r, helpers.SessionGood, "Steam unlinked")
	log.Err(err)

}

func linkDiscordCallbackHandler(w http.ResponseWriter, r *http.Request) {

}
