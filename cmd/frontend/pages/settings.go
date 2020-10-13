package pages

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"sync"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/badoux/checkmail"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/geo"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/middleware"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/oauth"
	"github.com/go-chi/chi"
	"github.com/mssola/user_agent"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func SettingsRouter() http.Handler {

	r := chi.NewRouter()
	r.Use(middleware.MiddlewareAuthCheck())

	r.Get("/", settingsHandler)
	r.Post("/update", settingsPostHandler)
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
	t.fill(w, r, "settings", "Settings", "Game DB settings")
	t.addAssetPasswordStrength()
	t.ProdCCs = i18n.GetProdCCs(true)

	// Get user
	t.User, err = getUserFromSession(r)
	if err == ErrLoggedOut {
		returnErrorTemplate(w, r, errorTemplate{Code: http.StatusForbidden})
		return
	} else if err != nil {
		log.ErrS(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500})
		return
	}

	// Get player
	t.Player, err = getPlayerFromSession(r)
	if err != nil {
		err = helpers.IgnoreErrors(err, ErrLoggedOut)
		if err != nil {
			log.ErrS(err)
		}
	} else {
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

		if _, ok := t.UserProviders[oauth.ProviderSteam]; !ok {
			t.Banners = append(t.Banners, "<a href='/oauth/out/steam?page=settings'>Link your Steam account.</a>")
		}
	}()

	// Wait
	wg.Wait()

	t.Providers = oauth.Providers

	// Template
	returnTemplate(w, r, t)
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
	Banners       []template.HTML
}

func settingsPostHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		session.Save(w, r)
		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	// Get user
	user, err := getUserFromSession(r)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "User not found")
		return
	}

	// Parse form
	err = r.ParseForm()
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "Could not read form data")
		return
	}

	// Email
	email := r.PostForm.Get("email")
	if email != user.Email {

		err = checkmail.ValidateFormat(r.PostForm.Get("email"))
		if err != nil {
			session.SetFlash(r, session.SessionBad, "Invalid email address")
			return
		}

		user.Email = email
		user.EmailVerified = false
	}

	// Password
	password := r.PostForm.Get("password")
	if password != "" {

		if len(password) < 8 {
			session.SetFlash(r, session.SessionBad, "Password must be at least 8 characters long")
			return
		}

		passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
		if err != nil {
			log.ErrS(err)
			session.SetFlash(r, session.SessionBad, "Something went wrong encrypting your password")
			return
		}

		user.Password = string(passwordBytes)
	}

	// Country code
	prodCC := steamapi.ProductCC(r.PostForm.Get("prod_cc"))

	if i18n.IsValidProdCC(prodCC) {
		user.ProductCC = prodCC
	} else {
		user.ProductCC = steamapi.ProductCCUS
	}

	// Save hidden
	// if r.PostForm.Get("hide") == "1" {
	// 	user.HideProfile = true
	// } else {
	// 	user.HideProfile = false
	// }

	// Save alerts
	// if r.PostForm.Get("alerts") == "1" {
	// 	user.ShowAlerts = true
	// } else {
	// 	user.ShowAlerts = false
	// }

	// Save user
	db, err := mysql.GetMySQLClient()
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "We had trouble saving your settings")
		return
	}

	// Have to save as a map because gorm does not save empty values otherwise
	db = db.Model(&user).Updates(map[string]interface{}{
		"email":          user.Email,
		"email_verified": user.EmailVerified,
		"password":       user.Password,
		"country_code":   user.ProductCC,
		// "hide_profile":   user.HideProfile,
		// "show_alerts":    user.ShowAlerts,
	})

	if db.Error != nil {
		log.ErrS(db.Error)
		session.SetFlash(r, session.SessionBad, "Something went wrong saving your settings")
		return
	}

	// Save player
	playerID := session.GetPlayerIDFromSesion(r)

	if playerID > 0 {
		filter := bson.D{{"_id", playerID}}
		update := bson.D{{"private", r.PostForm.Get("private") == "1"}}

		_, err = mongo.UpdateOne(mongo.CollectionPlayers, filter, update)
		if err != nil {
			log.ErrS(err)
			session.SetFlash(r, session.SessionBad, "We had trouble saving your settings")
			return
		}

		err = memcache.Delete(memcache.MemcachePlayer(playerID).Key)
		if err != nil {
			log.ErrS(err)
		}
	}

	// Update session
	session.SetMany(r, map[string]string{
		session.SessionUserProdCC: string(user.ProductCC),
		session.SessionUserEmail:  user.Email,
		// session.SessionUserShowAlerts: strconv.FormatBool(user.ShowAlerts),
	})

	session.SetFlash(r, session.SessionGood, "Settings saved")
}

func settingsNewKeyHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		session.Save(w, r)
		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	// Get user
	user, err := getUserFromSession(r)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "User not found")
		return
	}

	user.SetAPIKey()

	// Save user
	db, err := mysql.GetMySQLClient()
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "We had trouble saving your settings (1001)")
		return
	}

	db = db.Model(&user).Update("api_key", user.APIKey)
	if db.Error != nil {
		log.ErrS(db.Error)
		session.SetFlash(r, session.SessionBad, "We had trouble saving your settings (1002)")
		return
	}

	// Update session
	session.SetMany(r, map[string]string{
		session.SessionUserAPIKey: user.APIKey,
	})

	session.SetFlash(r, session.SessionGood, "New API key generated")
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
	for _, event := range events {

		// Parse user agent
		ua := user_agent.New(event.UserAgent)
		browser, version := ua.Browser()
		agent := ua.OSInfo().Name + " " + ua.OSInfo().Version + " - " + browser + " " + version

		// Get IP location
		var location []string
		record, err := geo.GetCountryCode(event.IP)
		if err == nil {
			if val, ok := record.Country.Names["en"]; ok {
				location = append(location, val)
			}
			if val, ok := record.City.Names["en"]; ok {
				location = append(location, val)
			}
		}

		if len(location) == 0 {
			location = append(location, geo.GetFirstIP(event.IP))
		}

		//
		response.AddRow([]interface{}{
			event.CreatedAt.Unix(),       // 0
			event.GetCreatedNice(),       // 1
			event.GetType(),              // 2
			geo.GetFirstIP(event.IP),     // 3
			event.UserAgent,              // 4
			agent,                        // 5
			geo.GetFirstIP(r.RemoteAddr), // 6
			event.GetIcon(),              // 7
			strings.Join(location, ", "), // 8
		})
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
		response.AddRow([]interface{}{
			v.IP,
		})
	}

	returnJSON(w, r, response)
}

func settingsRemoveProviderHandler(w http.ResponseWriter, r *http.Request) {

	provider := oauth.New(oauth.ProviderEnum(chi.URLParam(r, "provider")))
	if provider == nil {
		Error404Handler(w, r)
		return
	}

	defer func() {
		session.Save(w, r)
		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	userID := session.GetUserIDFromSesion(r)
	if userID == 0 {
		session.SetFlash(r, session.SessionBad, "An error occurred (1001)")
		return
	}

	// Update user
	err := mysql.DeleteUserProvider(provider.GetEnum(), userID)
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
	err = mongo.NewEvent(r, userID, mongo.EventUnlink(provider.GetEnum()))
	if err != nil {
		log.ErrS(err)
	}
}
