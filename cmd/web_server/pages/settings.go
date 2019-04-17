package pages

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"sync"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/pkg"
	"github.com/go-chi/chi"
	"golang.org/x/crypto/bcrypt"
)

var (
	errNotLoggedIn = errors.New("not logged in")
)

func settingsRouter() http.Handler {
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

	player, err := getPlayer(r)
	if err != nil {
		if err == errNotLoggedIn {
			err := pkg.SetBadFlash(w, r, "please login")
			log.Err(err, r)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving your data.", Error: err})
		return
	}

	//
	var wg sync.WaitGroup

	// Get donations
	var donations []pkg.Donation
	wg.Add(1)
	go func(player pkg.Player) {

		defer wg.Done()

		// if player.Donated > 0 {
		// 	donations, err = db.GetDonations(player.ID, 10)
		// 	log.Err(err, r)
		// }

	}(player)

	// Get games
	var games string
	wg.Add(1)
	go func(player pkg.Player) {

		defer wg.Done()

		playerApps, err := pkg.GetPlayerApps(player.ID, 0, 0, pkg.D{})
		if err != nil {
			log.Err(err, r)
			return
		}

		var appIDs []int
		for _, v := range playerApps {
			appIDs = append(appIDs, v.AppID)
		}

		games = string(pkg.MarshalLog(appIDs))

	}(player)

	// Get User
	var user pkg.User
	wg.Add(1)
	go func(player pkg.Player) {

		defer wg.Done()

		user, err = getUser(r, 0)
		if err != nil {
			log.Err(err, r)
			return
		}

	}(player)

	// Wait
	wg.Wait()

	// Countries
	var countries [][]string
	for k, v := range steam.Countries {
		countries = append(countries, []string{string(k), v})
	}
	sort.Slice(countries, func(i, j int) bool {
		return countries[i][1] < countries[j][1]
	})

	// Template
	t := settingsTemplate{}
	t.fill(w, r, "Settings", "")
	t.addAssetPasswordStrength()
	t.Player = player
	t.User = user
	t.Donations = donations
	t.Games = games
	t.Countries = countries

	err = returnTemplate(w, r, "settings", t)
	log.Err(err, r)
}

type settingsTemplate struct {
	GlobalTemplate
	Player    pkg.Player
	User      pkg.User
	Donations []pkg.Donation
	Games     string
	Messages  []interface{}
	Countries [][]string
}

func settingsPostHandler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, 0)

	// Get user
	user, err := getUser(r, 0)
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

	// Save password
	password := r.PostForm.Get("password")

	if len(password) > 0 {

		if len(password) < 8 {
			err := pkg.SetBadFlash(w, r, "Password must be at least 8 characters long")
			log.Err(err, r)
			http.Redirect(w, r, "/settings", http.StatusFound)
			return
		}

		passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
		if err != nil {
			log.Err(err, r)
			err := pkg.SetBadFlash(w, r, "Something went wrong encrypting your password")
			log.Err(err, r)
			http.Redirect(w, r, "/settings", http.StatusFound)
			return
		}

		user.Password = string(passwordBytes)
	}

	// Save email
	user.Email = r.PostForm.Get("email")
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
	err = user.Save()
	log.Err(err, r)
	if err != nil {
		err = pkg.SetBadFlash(w, r, "Something went wrong saving settings")
		log.Err(err, r)
	} else {
		err = pkg.SetGoodFlash(w, r, "Settings saved")
		log.Err(err, r)
	}

	// Update session
	err = pkg.WriteMany(w, r, map[string]string{
		pkg.UserCountry: user.CountryCode,
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
	var events []pkg.Event
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		events, err = pkg.GetEvents(playerID, query.getOffset64())
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

		total, err = pkg.CountEvents(playerID)
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
	loggedIn, err := pkg.IsLoggedIn(r)
	if err != nil {
		return playerID, errNotLoggedIn
	}

	if !loggedIn {
		return playerID, errNotLoggedIn
	}

	// Get session
	id, err := pkg.Read(r, pkg.PlayerID)
	if err != nil {
		return playerID, err
	}

	// Convert ID
	return strconv.ParseInt(id, 10, 64)
}

func getPlayer(r *http.Request) (player pkg.Player, err error) {

	playerID, err := getPlayerIDFromSession(r)
	if err != nil {
		return player, err
	}

	return pkg.GetPlayer(playerID)
}

func getUser(r *http.Request, playerID int64) (user pkg.User, err error) {

	if playerID == 0 {
		playerID, err = getPlayerIDFromSession(r)
		if err != nil {
			return user, err
		}
	}

	user, err = pkg.GetUser(playerID)
	if err != nil {
		return user, err
	}

	return user, nil
}
