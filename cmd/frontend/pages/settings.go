package pages

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"sync"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/badoux/checkmail"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/middleware"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/oauth"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func SettingsRouter() http.Handler {

	r := chi.NewRouter()
	r.Use(middleware.MiddlewareAuthCheck())

	r.Get("/", settingsHandler)
	r.Post("/update", settingsPostHandler)
	r.Post("/delete", deletePostHandler)
	r.Get("/events.json", settingsEventsAjaxHandler)
	r.Get("/new-key", settingsNewKeyHandler)
	r.Get("/donations.json", settingsDonationsAjaxHandler)
	r.Get("/remove-provider/{provider:[a-z]+}", settingsRemoveProviderHandler)

	return r
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	//
	t := settingsTemplate{}
	t.fill(w, r, "Settings", "Game DB settings")
	t.addAssetPasswordStrength()
	t.ProdCCs = i18n.GetProdCCs(true)

	// Get user
	t.User, err = getUserFromSession(r)
	if err != nil {
		log.ErrS(err)
	}

	steamID := t.User.GetSteamID()
	if steamID > 0 {

		// Get player
		t.Player, err = mongo.GetPlayer(steamID)
		err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
		if err != nil {
			log.ErrS(err)
		}

		// Set Steam player name to session if missing, can happen after linking
		session.Set(r, session.SessionPlayerName, t.Player.GetName())
	}

	//
	var wg sync.WaitGroup

	// Get games
	wg.Add(1)
	go func() {

		defer wg.Done()

		if t.Player.ID == 0 {
			return
		}

		playerApps, err := mongo.GetPlayerApps(0, 0, bson.D{{"player_id", t.Player.ID}}, bson.D{})
		if err != nil {
			log.ErrS(err)
			return
		}

		var appIDs []int
		for _, v := range playerApps {
			appIDs = append(appIDs, v.AppID)
		}

		b, err := json.Marshal(appIDs)
		if err != nil {
			log.ErrS(err)
		}

		t.Games = template.JS(b)
	}()

	// Get groups
	wg.Add(1)
	go func() {

		defer wg.Done()

		if t.Player.ID == 0 {
			return
		}

		var groupIDs []string

		groups, err := mongo.GetPlayerGroups(t.Player.ID, 0, 0, nil)
		if err != nil {
			log.ErrS(err)
			return
		}
		for _, v := range groups {
			groupIDs = append(groupIDs, v.GroupID)
		}

		b, err := json.Marshal(groupIDs)
		if err != nil {
			log.ErrS(err)
		}

		t.Groups = template.JS(b)
	}()

	// Get badges
	wg.Add(1)
	go func() {

		defer wg.Done()

		if t.Player.ID == 0 {
			return
		}

		var badgeIDs []int
		var filter = bson.D{{Key: "player_id", Value: t.Player.ID}}

		badges, err := mongo.GetPlayerBadges(0, filter, nil)
		if err != nil {
			log.ErrS(err)
			return
		}
		for _, v := range badges {
			badgeIDs = append(badgeIDs, v.ID())
		}

		b, err := json.Marshal(badgeIDs)
		if err != nil {
			log.ErrS(err)
		}

		t.Badges = template.JS(b)
	}()

	// Get providers
	wg.Add(1)
	go func() {

		defer wg.Done()

		providers, err := mysql.GetUserProviders(t.User.ID)
		if err != nil {
			log.ErrS(err)
			return
		}

		t.UserProviders = map[oauth.ProviderEnum]mysql.UserProvider{}
		for _, v := range providers {
			t.UserProviders[v.Provider] = v
		}
	}()

	// Wait
	wg.Wait()

	t.Providers = []oauth.Provider{
		oauth.New(oauth.ProviderSteam),
		oauth.New(oauth.ProviderDiscord),
		oauth.New(oauth.ProviderGoogle),
		// oauth.New(oauth.ProviderTwitter),
		oauth.New(oauth.ProviderPatreon),
		oauth.New(oauth.ProviderGithub),
	}

	// Template
	returnTemplate(w, r, "settings", t)
}

type settingsTemplate struct {
	globalTemplate
	User          mysql.User
	Player        mongo.Player
	ProdCCs       []i18n.ProductCountryCode
	Groups        template.JS
	Badges        template.JS
	Games         template.JS
	Providers     []oauth.Provider
	UserProviders map[oauth.ProviderEnum]mysql.UserProvider
}

func deletePostHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	redirect, good, bad := func() (redirect string, good string, bad string) {

		// Parse form
		err = r.ParseForm()
		if err != nil {
			log.ErrS(err)
			return "/settings", "", "There was an eror saving your information."
		}

		user, err := getUserFromSession(r)
		if err != nil {
			log.ErrS(err)
			return "/settings", "", "There was an eror saving your information."
		}

		if r.PostForm.Get("id") == user.SteamID.String {

			session.DeleteAll(r)
			return "/", "Your account has been deleted", ""

		}

		return "/settings", "", "Invalid player ID."
	}()

	if good != "" {
		session.SetFlash(r, session.SessionGood, good)
	}
	if bad != "" {
		session.SetFlash(r, session.SessionBad, bad)
	}

	session.Save(w, r)

	http.Redirect(w, r, redirect, http.StatusFound)
}

func settingsPostHandler(w http.ResponseWriter, r *http.Request) {

	redirect, good, bad := func() (redirect string, good string, bad string) {

		// Get user
		user, err := getUserFromSession(r)
		if err != nil {
			log.ErrS(err)
			return "/settings", "", "User not found"
		}

		// Parse form
		err = r.ParseForm()
		if err != nil {
			log.ErrS(err)
			return "/settings", "", "Could not read form data"
		}

		email := r.PostForm.Get("email")
		password := r.PostForm.Get("password")
		prodCC := steamapi.ProductCC(r.PostForm.Get("prod_cc"))

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
			if err != nil {
				log.ErrS(err)
				return "/settings", "", "Something went wrong encrypting your password"
			}

			user.Password = string(passwordBytes)
		}

		// Country code
		if i18n.IsValidProdCC(prodCC) {
			user.ProductCC = prodCC
		} else {
			user.ProductCC = steamapi.ProductCCUS
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
		db, err := mysql.GetMySQLClient()
		if err != nil {
			log.ErrS(err)
			return "/settings", "", "We had trouble saving your settings"
		}

		// Have to save as a map because gorm does not save empty values otherwise
		db = db.Model(&user).Updates(map[string]interface{}{
			"email":          user.Email,
			"email_verified": user.EmailVerified,
			"password":       user.Password,
			"hide_profile":   user.HideProfile,
			"show_alerts":    user.ShowAlerts,
			"country_code":   user.ProductCC,
		})

		if db.Error != nil {
			log.ErrS(db.Error)
			return "/settings", "", "Something went wrong saving your settings"
		}

		// Update session
		session.SetMany(r, map[string]string{
			session.SessionUserProdCC:     string(user.ProductCC),
			session.SessionUserEmail:      user.Email,
			session.SessionUserShowAlerts: strconv.FormatBool(user.ShowAlerts),
		})

		return "/settings", "Settings saved", ""
	}()

	if good != "" {
		session.SetFlash(r, session.SessionGood, good)
	}
	if bad != "" {
		session.SetFlash(r, session.SessionBad, bad)
	}

	session.Save(w, r)

	http.Redirect(w, r, redirect, http.StatusFound)
}

func settingsNewKeyHandler(w http.ResponseWriter, r *http.Request) {

	good, bad := func() (good string, bad string) {

		// Get user
		user, err := getUserFromSession(r)
		if err != nil {
			log.ErrS(err)
			return "", "User not found"
		}

		user.SetAPIKey()

		// Save user
		db, err := mysql.GetMySQLClient()
		if err != nil {
			log.ErrS(err)
			return "", "We had trouble saving your settings (1001)"
		}

		db = db.Model(&user).Update("api_key", user.APIKey)
		if db.Error != nil {
			log.ErrS(db.Error)
			return "", "We had trouble saving your settings (1002)"
		}

		// Update session
		session.SetMany(r, map[string]string{
			session.SessionUserAPIKey: user.APIKey,
		})

		return "New API key generated", ""
	}()

	if good != "" {
		session.SetFlash(r, session.SessionGood, good)
	}
	if bad != "" {
		session.SetFlash(r, session.SessionBad, bad)
	}

	session.Save(w, r)

	http.Redirect(w, r, "/settings", http.StatusFound)
}

func settingsEventsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	user, err := getUserFromSession(r)
	if err != nil {
		log.ErrS(err)
		return
	}

	query := datatable.NewDataTableQuery(r, true)

	var wg sync.WaitGroup

	// Get events
	var events []mongo.Event
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		events, err = mongo.GetEvents(user.ID, query.GetOffset64())
		if err != nil {
			log.ErrS(err)
			return
		}

	}(r)

	// Get total
	var total int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		total, err = mongo.CountDocuments(mongo.CollectionEvents, bson.D{{"user_id", user.ID}}, 86400)
		if err != nil {
			log.ErrS(err)
		}
	}(r)

	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, total, total, nil)
	for _, v := range events {
		response.AddRow(v.OutputForJSON(r.RemoteAddr))
	}

	returnJSON(w, r, response)
}

func settingsDonationsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	user, err := getUserFromSession(r)
	if err != nil {
		log.ErrS(err)
		return
	}

	query := datatable.NewDataTableQuery(r, true)

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

		total, err = mongo.CountDocuments(mongo.CollectionWebhooks, bson.D{{"user_id", user.ID}}, 0)
		if err != nil {
			log.ErrS(err)
		}
	}(r)

	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, total, total, nil)
	for _, v := range events {
		response.AddRow(v.OutputForJSON(r.RemoteAddr))
	}

	returnJSON(w, r, response)
}

func settingsRemoveProviderHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		session.Save(w, r)
		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	provider := oauth.New(oauth.ProviderEnum(chi.URLParam(r, "provider")))
	if provider == nil {
		session.SetFlash(r, session.SessionBad, "Invalid Provider")
		return
	}

	userID, err := session.GetUserIDFromSesion(r)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1001)")
		return
	}

	// Update user
	err = mysql.DeleteUserProvider(provider.GetEnum(), userID)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1002)")
		return
	}

	// Clear session
	if provider.GetEnum() == oauth.ProviderSteam {
		session.DeleteMany(r, []string{session.SessionPlayerID, session.SessionPlayerName, session.SessionPlayerLevel})
	}

	// Flash message
	session.SetFlash(r, session.SessionGood, provider.GetName()+" removed")

	// Create event
	err = mongo.CreateUserEvent(r, userID, mongo.EventUnlink(provider.GetEnum()))
	if err != nil {
		log.ErrS(err)
	}
}
