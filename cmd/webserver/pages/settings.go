package pages

import (
	"errors"
	"net/http"
	"strconv"
	"sync"

	"github.com/Jleagle/steam-go/steam"
	"github.com/badoux/checkmail"
	"github.com/gamedb/gamedb/cmd/webserver/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
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
	// r.Post("/delete", deletePostHandler)
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
		err := session.SetBadFlash(r, "Please login")
		log.Err(err, r)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	//
	t := settingsTemplate{}
	t.fill(w, r, "Settings", "")
	t.addAssetPasswordStrength()
	t.setFlashes(w, r, true)

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
		err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
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

func deletePostHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, 0)

	var err error

	redirect, good, bad := func() (redirect string, good string, bad string) {

		loggedIn, err := session.IsLoggedIn(r)
		log.Err(err)

		if !loggedIn {
			return "/login", "", "Please login"
		}

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

		if r.PostForm.Get("id") == strconv.FormatInt(user.PlayerID, 10) {

			err = session.Clear(r)
			log.Err(err)
			return "/", "Your account has been deleted", ""

		} else {
			return "/settings", "", "Invalid player ID."
		}
	}()

	if good != "" {
		err = session.SetGoodFlash(r, good)
		log.Err(err)
	}
	if bad != "" {
		err = session.SetBadFlash(r, bad)
		log.Err(err)
	}

	err = session.Save(w, r)
	log.Err(err)

	http.Redirect(w, r, redirect, http.StatusFound)
}

func settingsPostHandler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, 0)

	var err error

	redirect, good, bad := func() (redirect string, good string, bad string) {

		loggedIn, err := session.IsLoggedIn(r)
		log.Err(err)
		if !loggedIn || err != nil {
			return "/login", "", "Please login"
		}

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

		// Email
		email := r.PostForm.Get("email")

		if email != "" && email != user.Email {

			err = checkmail.ValidateFormat(r.PostForm.Get("email"))
			if err != nil {
				return "/settings", "", "Invalid email address"
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
		code := r.PostForm.Get("country_code")

		if _, ok := steam.Countries[steam.CountryCode(code)]; ok {
			user.CountryCode = code
		} else {
			user.CountryCode = string(steam.CountryUS)
		}

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
		log.Err(err)
		if err != nil {
			return "/settings", "", "We had trouble saving your settings"
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
			return "/settings", "", "Something went wrong saving your settings"
		}

		// Update session
		err = session.WriteMany(w, r, map[string]string{
			session.UserCountry:    user.CountryCode,
			session.UserEmail:      user.Email,
			session.UserShowAlerts: strconv.Itoa(int(user.ShowAlerts)),
		})
		log.Err(err, r)
		if err != nil {
			return "/settings", "", "Something went wrong saving your settings"
		}

		return "/settings", "Settings saved", ""
	}()

	if good != "" {
		err = session.SetGoodFlash(r, good)
		log.Err(err)
	}
	if bad != "" {
		err = session.SetBadFlash(r, bad)
		log.Err(err)
	}

	err = session.Save(w, r)
	log.Err(err)

	http.Redirect(w, r, redirect, http.StatusFound)
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
