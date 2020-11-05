package pages

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/badoux/checkmail"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/geo"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/oauth"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/middleware"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
	"github.com/mssola/user_agent"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

func SettingsRouter() http.Handler {

	r := chi.NewRouter()
	r.Use(middleware.MiddlewareAuthCheck)

	r.Get("/", settingsHandler)
	r.Post("/update", settingsPostHandler)
	r.Get("/events.json", settingsEventsAjaxHandler)
	r.Get("/new-key", settingsNewKeyHandler)
	r.Get("/donations.json", settingsDonationsAjaxHandler)
	r.Get("/remove-provider/{provider:[a-z]+}", settingsRemoveProviderHandler)
	r.Get("/join-discord-server", joinDiscordServerHandler)

	return r
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	//
	t := settingsTemplate{}
	t.fill(w, r, "settings", "Settings", "Game DB settings")
	t.addAssetPasswordStrength()
	t.addAssetChosen()
	t.addAssetBootbox()
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

		// Get providers user has
		providers, err := mysql.GetUserProviders(t.User.ID)
		if err != nil {
			log.ErrS(err)
			return
		}

		t.UserProviders = map[oauth.ProviderEnum]mysql.UserProvider{}
		for _, v := range providers {
			t.UserProviders[v.Provider] = v
		}

		// Check if user has not linked steam yet
		if _, ok := t.UserProviders[oauth.ProviderSteam]; !ok {
			t.Banners = append(t.Banners, "<a href='/oauth/out/steam?page=settings'>Link your Steam account.</a>")
		}

		// Check if user is in discord server
		if val, ok := t.UserProviders[oauth.ProviderDiscord]; ok {

			var inGuild bool
			var item = memcache.MemcacheUserInDiscord(val.ID)

			err = memcache.GetSetInterface(item.Key, item.Expiration, &inGuild, func() (interface{}, error) {

				discord, err := discordgo.New("Bot " + config.C.DiscordOAuthBotToken)
				if err != nil {
					return false, err
				}

				_, err = discord.GuildMember(helpers.GuildID, val.ID)
				if val, ok := err.(*discordgo.RESTError); ok && val.Response.StatusCode == 404 {
					return false, nil // 404, not in guild
				}

				if err != nil {
					return false, err // unknown error
				} else {
					return true, nil // no error, in guild
				}
			})

			if !inGuild {
				t.Banners = append(t.Banners, "<a href='/settings/join-discord-server'>Join the Discord server!</a>")
			}
		}
	}()

	// Get event types
	wg.Add(1)
	go func() {

		defer wg.Done()

		types, err := mongo.GetEventCounts(t.User.ID)
		if err != nil {
			log.ErrS(err)
			return
		}

		for _, v := range types {
			t.EventTypes = append(t.EventTypes, settingsEventTemplate{
				ID:    v.ID,
				Name:  mongo.EventEnum(v.ID).ToString(),
				Count: v.Count,
			})
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
	EventTypes    []settingsEventTemplate
}

type settingsEventTemplate struct {
	ID    string
	Name  string
	Count int
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

func joinDiscordServerHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		session.Save(w, r)
		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	userID := session.GetUserIDFromSesion(r)
	if userID == 0 {
		session.SetFlash(r, session.SessionBad, "Can't find user session")
		return
	}

	provider, err := mysql.GetUserProviderByUserID(oauth.ProviderDiscord, userID)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "Can't find user session")
		return
	}

	var token oauth2.Token
	err = json.Unmarshal([]byte(provider.Token), &token)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "Something went wrong (1001)")
		return
	}

	if token.Expiry.Before(time.Now().Add(time.Minute)) {
		http.Redirect(w, r, config.C.DiscordServerInviteURL, http.StatusFound)
		return
	}

	discord, err := discordgo.New("Bot " + config.C.DiscordOAuthBotToken)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "Something went wrong (1002)")
		return
	}

	err = discord.GuildMemberAdd(token.AccessToken, helpers.GuildID, provider.ID, "", nil, false, false)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "Something went wrong (1003)")
		return
	}

	session.SetFlash(r, session.SessionGood, "You are now in the server")
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

	userID := session.GetUserIDFromSesion(r)
	if userID == 0 {
		return
	}

	query := datatable.NewDataTableQuery(r, true)
	types := query.GetSearchSlice("type")

	var filter = bson.D{{"user_id", userID}}
	if len(types) > 0 {
		filter = append(filter, bson.E{Key: "type", Value: bson.M{"$in": types}})
	}

	var wg sync.WaitGroup

	// Get events
	var events []mongo.Event
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		events, err = mongo.GetEvents(filter, query.GetOffset64())
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionEvents, bson.D{{"user_id", userID}}, 86400)
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, total, total, nil)
	for _, event := range events {

		// Parse user agent
		ua := user_agent.New(event.UserAgent)
		browser, version := ua.Browser()
		agent := ua.OSInfo().Name + " " + ua.OSInfo().Version + " - " + browser + " " + version

		// Get IP location
		var location []string
		record, err := geo.GetLocation(event.IP)
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
			event.Type.ToString(),        // 2
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

	query := datatable.NewDataTableQuery(r, false)
	userID := session.GetUserIDFromSesion(r)
	if userID == 0 {
		return
	}

	var wg sync.WaitGroup

	// Get donations
	var donations []mysql.Donation
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		donations, err = mysql.GetDonationsByUser(userID, query.GetOffset())
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get total
	var total int
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mysql.GetDonationCountByUser(userID)
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, int64(total), int64(total), nil)
	for _, donation := range donations {
		response.AddRow([]interface{}{
			donation.CreatedAt.Unix(),                       // 0
			donation.CreatedAt.Format(helpers.DateYearTime), // 1
			donation.AmountUSD,                              // 2
			strings.Title(donation.Source),                  // 3
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

	user, err := getUserFromSession(r)
	if err != nil {
		err = helpers.IgnoreErrors(err, ErrLoggedOut)
		if err != nil {
			log.ErrS(err)
		}
		session.SetFlash(r, session.SessionBad, "An error occurred (1001)")
		return
	}

	if user.Password == "" {

		providers, err := mysql.GetUserProviders(user.ID)
		if err != nil {
			log.ErrS(err)
			session.SetFlash(r, session.SessionBad, "An error occurred (1002)")
			return
		}

		if len(providers) < 2 {
			session.SetFlash(r, session.SessionBad, "You need at least one provider if you are not using a password")
			return
		}
	}

	// Update user
	err = mysql.DeleteUserProvider(provider.GetEnum(), user.ID)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1003)")
		return
	}

	// Clear session
	if provider.GetEnum() == oauth.ProviderSteam {
		session.DeleteMany(r, []string{session.SessionPlayerID, session.SessionPlayerName, session.SessionPlayerLevel})
	}

	// Flash message
	session.SetFlash(r, session.SessionGood, provider.GetName()+" removed")

	// Create event
	err = mongo.NewEvent(r, user.ID, mongo.EventUnlink(provider.GetEnum()))
	if err != nil {
		log.ErrS(err)
	}
}
