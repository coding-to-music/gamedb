package pages

import (
	"errors"
	"net/http"
	"strconv"
	"sync"

	"github.com/Jleagle/steam-go/steam"
	"github.com/badoux/checkmail"
	"github.com/gamedb/website/pkg/helpers"
	"github.com/gamedb/website/pkg/log"
	"github.com/gamedb/website/pkg/mongo"
	"github.com/gamedb/website/pkg/session"
	"github.com/gamedb/website/pkg/sql"
	"github.com/go-chi/chi"
	"golang.org/x/crypto/bcrypt"
)

var (
	errNotLoggedIn = errors.New("not logged in")
)

func SettingsRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", settingsHandler)
	r.Post("/", settingsPostHandler)
	r.Get("/events.json", settingsEventsAjaxHandler)
	return r
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, 0)

	loggedIn, err := session.IsLoggedIn(r)
	log.Err(err)

	if !loggedIn {
		err := session.SetBadFlash(w, r, "Please login")
		log.Err(err, r)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	//
	t := settingsTemplate{}
	t.fill(w, r, "Settings", "")
	t.addAssetPasswordStrength()

	//
	var wg sync.WaitGroup

	// Get games
	wg.Add(1)
	go func() {

		defer wg.Done()

		id, err := getPlayerIDFromSession(r)
		if err != nil {
			log.Err(err, r)
			return
		}

		playerApps, err := mongo.GetPlayerApps(id, 0, 0, mongo.D{})
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

	// Get User
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.User, err = getUserFromSession(r)
		log.Err(err)
	}()

	// Get Player
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Player, err = getPlayerFromSession(r)
		log.Err(err)
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
	Player    mongo.Player
	User      sql.User
	Games     string
	Messages  []interface{}
	Countries [][]string
}

func settingsPostHandler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, 0)

	// Get user
	user, err := getUserFromSession(r)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an eror saving your information.", Error: err})
		return
	}

	// Parse form
	err = r.ParseForm()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an eror saving your information.", Error: err})
		return
	}

	// Email
	email := r.PostForm.Get("email")
	if email != "" {

		err = checkmail.ValidateFormat(r.PostForm.Get("email"))
		if err != nil {
			err = session.SetBadFlash(w, r, "Invalid email address")
			http.Redirect(w, r, "/settings", http.StatusFound)
			return
		}

		user.Email = r.PostForm.Get("email")
	}

	// Password
	password := r.PostForm.Get("password")

	if email != user.Email {
		user.Verified = false
	}

	if password != "" {

		if len(password) < 8 {
			err := session.SetBadFlash(w, r, "Password must be at least 8 characters long")
			log.Err(err, r)
			http.Redirect(w, r, "/settings", http.StatusFound)
			return
		}

		passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
		if err != nil {
			log.Err(err, r)
			err := session.SetBadFlash(w, r, "Something went wrong encrypting your password")
			log.Err(err, r)
			http.Redirect(w, r, "/settings", http.StatusFound)
			return
		}

		user.Password = string(passwordBytes)
	}

	// Country code
	user.CountryCode = r.PostForm.Get("country_code")

	// Save hidden
	if r.PostForm.Get("hide") == "1" {
		user.HideProfile = 1
	} else {
		user.HideProfile = 0
	}

	// Save alerts
	if r.PostForm.Get("alerts") == "1" {
		user.ShowAlerts = 1
	} else {
		user.ShowAlerts = 0
	}

	// Save user

	db, err := sql.GetMySQLClient()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an eror saving your information.", Error: err})
		return
	}

	// Have to save as a map because gorm does not save empty values otherwise
	db = db.Model(&user).Updates(map[string]interface{}{
		"email":        user.Email,
		"verified":     user.Verified,
		"password":     user.Password,
		"hide_profile": user.HideProfile,
		"show_alerts":  user.ShowAlerts,
		"country_code": user.CountryCode,
	})

	log.Err(db.Error, r)

	if db.Error != nil {
		err = session.SetBadFlash(w, r, "Something went wrong saving settings")
		log.Err(err, r)
	} else {
		err = session.SetGoodFlash(w, r, "Settings saved")
		log.Err(err, r)
	}

	// Update session
	err = session.WriteMany(w, r, map[string]string{
		session.UserCountry: user.CountryCode,
	})
	log.Err(err, r)

	http.Redirect(w, r, "/settings", http.StatusFound)
}

func settingsEventsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{"draw", "start"})
	if ret {
		return
	}

	setCacheHeaders(w, 0)

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	// Get player ID
	playerID, err := getPlayerIDFromSession(r)
	if err != nil {
		log.Err(err, r)
		return
	}

	var wg sync.WaitGroup

	// Get events
	var events []mongo.Event
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		events, err = mongo.GetEvents(playerID, query.getOffset64())
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

		total, err = mongo.CountEvents(playerID)
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

func getPlayerIDFromSession(r *http.Request) (playerID int64, err error) {

	// Check if logged in
	loggedIn, err := session.IsLoggedIn(r)
	if err != nil {
		return playerID, errNotLoggedIn
	}

	if !loggedIn {
		return playerID, errNotLoggedIn
	}

	// Get session
	id, err := session.Read(r, session.PlayerID)
	if err != nil {
		return playerID, err
	}

	// Convert ID
	return strconv.ParseInt(id, 10, 64)
}

func getPlayerFromSession(r *http.Request) (player mongo.Player, err error) {

	playerID, err := getPlayerIDFromSession(r)
	if err != nil {
		return player, err
	}

	return mongo.GetPlayer(playerID)
}

func getUserFromSession(r *http.Request) (user sql.User, err error) {

	playerID, err := getPlayerIDFromSession(r)
	if err != nil {
		return user, err
	}

	return sql.GetUser(playerID)
}
