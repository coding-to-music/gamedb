package pages

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/Jleagle/session-go/session"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/badoux/checkmail"
	"github.com/gamedb/gamedb/cmd/webserver/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/webserver/middleware"
	"github.com/gamedb/gamedb/cmd/webserver/oauth"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
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

	r.Get("/oauth-link/{id:[a-z]+}", oauthLinkHandler)
	r.Get("/oauth-unlink/{id:[a-z]+}", oauthUnlinkHandler)
	r.Get("/oauth-callback/{id:[a-z]+}", oauthCallbackHandler)

	return r
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	//
	t := settingsTemplate{}
	t.fill(w, r, "Settings", "Game DB settings")
	t.addAssetPasswordStrength()
	t.Domain = config.Config.GameDBDomain.Get()
	t.ProdCCs = helpers.GetProdCCs(true)

	// Get user
	t.User, err = getUserFromSession(r)
	if err != nil {
		log.Err(err, r)
	}

	steamID := t.User.GetSteamID()
	if steamID > 0 {

		// Get player
		t.Player, err = mongo.GetPlayer(steamID)
		err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
		if err != nil {
			log.Err(err, r)
		}

		// Set Steam player name to session if missing, can happen after linking
		err = session.Set(r, helpers.SessionPlayerName, t.Player.PersonaName)
		if err != nil {
			log.Err(err, r)
		}
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

		playerApps, err := mongo.GetPlayerApps(t.Player.ID, 0, 0, bson.D{})
		if err != nil {
			log.Err(err, r)
			return
		}

		var appIDs []int
		for _, v := range playerApps {
			appIDs = append(appIDs, v.AppID)
		}

		b, err := json.Marshal(appIDs)
		if err != nil {
			log.Err(err, r)
		}

		t.Games = string(b)
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
			log.Err(err)
			return
		}
		for _, v := range groups {
			groupIDs = append(groupIDs, v.GroupID)
		}

		b, err := json.Marshal(groupIDs)
		if err != nil {
			log.Err(err, r)
		}

		t.Groups = string(b)
	}()

	// Wait
	wg.Wait()

	// Template
	returnTemplate(w, r, "settings", t)
}

type settingsTemplate struct {
	GlobalTemplate
	User    sql.User
	Player  mongo.Player
	Games   string
	ProdCCs []helpers.ProductCountryCode
	Domain  string
	Groups  string
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

		if r.PostForm.Get("id") == user.SteamID.String {

			err = session.DeleteAll(r)
			log.Err(err)
			return "/", "Your account has been deleted", ""

		}

		return "/settings", "", "Invalid player ID."
	}()

	if good != "" {
		err = session.SetFlash(r, helpers.SessionGood, good)
		if err != nil {
			log.Err(err, r)
		}
	}
	if bad != "" {
		err = session.SetFlash(r, helpers.SessionBad, bad)
		if err != nil {
			log.Err(err, r)
		}
	}

	err = session.Save(w, r)
	if err != nil {
		log.Err(err, r)
	}

	http.Redirect(w, r, redirect, http.StatusFound)
}

func settingsPostHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	redirect, good, bad := func() (redirect string, good string, bad string) {

		// Get user
		user, err := getUserFromSession(r)
		if err != nil {
			log.Err(err, r)
			return "/settings", "", "User not found"
		}

		// Parse form
		err = r.ParseForm()
		if err != nil {
			log.Err(err, r)
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
				log.Err(err, r)
				return "/settings", "", "Something went wrong encrypting your password"
			}

			user.Password = string(passwordBytes)
		}

		// Country code
		if helpers.IsValidProdCC(prodCC) {
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
		db, err := sql.GetMySQLClient()
		if err != nil {
			log.Err(err, r)
			return "/settings", "", "We had trouble saving your settings"
		}

		// Have to save as a map because gorm does not save empty values otherwise
		db = db.Model(&user).Updates(map[string]interface{}{
			"email":        user.Email,
			"verified":     user.EmailVerified,
			"password":     user.Password,
			"hide_profile": user.HideProfile,
			"show_alerts":  user.ShowAlerts,
			"country_code": user.ProductCC,
		})

		if db.Error != nil {
			log.Err(db.Error, r)
			return "/settings", "", "Something went wrong saving your settings"
		}

		// Update session
		err = session.SetMany(r, map[string]string{
			helpers.SessionUserProdCC:     string(user.ProductCC),
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
		if err != nil {
			log.Err(err, r)
		}
	}
	if bad != "" {
		err = session.SetFlash(r, helpers.SessionBad, bad)
		if err != nil {
			log.Err(err, r)
		}
	}

	err = session.Save(w, r)
	if err != nil {
		log.Err(err, r)
	}

	http.Redirect(w, r, redirect, http.StatusFound)
}

func settingsNewKeyHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	good, bad := func() (good string, bad string) {

		// Get user
		user, err := getUserFromSession(r)
		if err != nil {
			log.Err(err, r)
			return "", "User not found"
		}

		user.SetAPIKey()

		// Save user
		db, err := sql.GetMySQLClient()
		log.Err(err)
		if err != nil {
			return "", "We had trouble saving your settings (1001)"
		}

		db = db.Model(&user).Update("api_key", user.APIKey)
		if db.Error != nil {
			log.Err(db.Error, r)
			return "", "We had trouble saving your settings (1002)"
		}

		// Update session
		err = session.SetMany(r, map[string]string{
			helpers.SessionUserAPIKey: user.APIKey,
		})
		if err != nil {
			log.Err(err, r)
			return "", "We had trouble saving your settings (1003)"
		}

		return "New API key generated", ""
	}()

	if good != "" {
		err = session.SetFlash(r, helpers.SessionGood, good)
		if err != nil {
			log.Err(err, r)
		}
	}
	if bad != "" {
		err = session.SetFlash(r, helpers.SessionBad, bad)
		if err != nil {
			log.Err(err, r)
		}
	}

	err = session.Save(w, r)
	if err != nil {
		log.Err(err, r)
	}

	http.Redirect(w, r, "/settings", http.StatusFound)
}

func settingsEventsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	user, err := getUserFromSession(r)
	if err != nil {
		log.Err(err)
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
			log.Err(err, r)
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
			log.Err(err, r)
		}
	}(r)

	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, total, total)
	for _, v := range events {
		response.AddRow(v.OutputForJSON(r.RemoteAddr))
	}

	returnJSON(w, r, response)
}

func settingsDonationsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	user, err := getUserFromSession(r)
	if err != nil {
		log.Err(err)
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

		total, err = mongo.CountDocuments(mongo.CollectionPatreonWebhooks, bson.D{{"user_id", user.ID}}, 0)
		log.Err(err, r)
	}(r)

	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, total, total)
	for _, v := range events {
		response.AddRow(v.OutputForJSON(r.RemoteAddr))
	}

	returnJSON(w, r, response)
}

func oauthLinkHandler(w http.ResponseWriter, r *http.Request) {

	id := oauth.ConnectionEnum(chi.URLParam(r, "id"))

	if _, ok := oauth.Connections[id]; ok {
		connection := oauth.New(id)
		connection.LinkHandler(w, r)
	}
}

func oauthUnlinkHandler(w http.ResponseWriter, r *http.Request) {

	id := oauth.ConnectionEnum(chi.URLParam(r, "id"))

	if _, ok := oauth.Connections[id]; ok {
		connection := oauth.New(id)
		connection.UnlinkHandler(w, r)
	}
}

func oauthCallbackHandler(w http.ResponseWriter, r *http.Request) {

	id := oauth.ConnectionEnum(chi.URLParam(r, "id"))

	if _, ok := oauth.Connections[id]; ok {
		connection := oauth.New(id)
		connection.LinkCallbackHandler(w, r)
	}
}
