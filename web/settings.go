package web

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
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

	setNoCacheHeaders(w)

	player, err := getPlayer(r)
	if err != nil {
		if err == errNotLoggedIn {
			err := session.SetBadFlash(w, r, "please login")
			log.Err(err, r)
			http.Redirect(w, r, "/login", 302)
			return
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving your data.", Error: err})
		return
	}

	//
	var wg sync.WaitGroup

	// Get donations
	var donations []db.Donation
	wg.Add(1)
	go func(player db.Player) {

		defer wg.Done()

		if player.Donated > 0 {
			donations, err = db.GetDonations(player.PlayerID, 10)
			log.Err(err, r)
		}

	}(player)

	// Get games
	var games string
	wg.Add(1)
	go func(player db.Player) {

		defer wg.Done()

		appIDs, err := player.GetAppIDs()
		if err != nil {
			log.Err(err, r)
			return
		}

		games = string(helpers.MarshalLog(appIDs))

	}(player)

	// Get User
	var user db.User
	wg.Add(1)
	go func(player db.Player) {

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
	t.Fill(w, r, "Settings", "")
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
	Player    db.Player
	User      db.User
	Donations []db.Donation
	Games     string
	Messages  []interface{}
	Countries [][]string
}

func settingsPostHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

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
			err := session.SetBadFlash(w, r, "Password must be at least 8 characters long")
			log.Err(err, r)
			http.Redirect(w, r, "/settings", 302)
			return
		}

		passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
		if err != nil {
			log.Err(err, r)
			err := session.SetBadFlash(w, r, "Something went wrong encrypting your password")
			log.Err(err, r)
			http.Redirect(w, r, "/settings", 302)
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

	http.Redirect(w, r, "/settings", 302)
}

func settingsEventsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	log.Err(err, r)

	//
	var wg sync.WaitGroup

	// Get events
	var events []db.Event

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		playerID, err := getPlayerIDFromSession(r)
		if err != nil {

			log.Err(err, r)
			return
		}

		client, ctx, err := db.GetDSClient()
		if err != nil {

			log.Err(err, r)
			return
		}

		q := datastore.NewQuery(db.KindEvent).Filter("player_id =", playerID).Order("-created_at").Limit(100)
		q, err = query.SetOffsetDS(q)
		if err != nil {

			log.Err(err, r)
			return
		}

		_, err = client.GetAll(ctx, q, &events)
		log.Err(err, r)

	}(r)

	// Get total
	var total int
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		playerID, err := getPlayerIDFromSession(r)
		if err != nil {

			log.Err(err, r)
			return
		}

		total, err = db.CountPlayerEvents(playerID)
		log.Err(err, r)

	}(r)

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(total)
	response.RecordsFiltered = strconv.Itoa(total)
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

func getPlayer(r *http.Request) (player db.Player, err error) {

	playerID, err := getPlayerIDFromSession(r)
	if err != nil {
		return player, err
	}

	return db.GetPlayer(playerID)
}

func getUser(r *http.Request, playerID int64) (user db.User, err error) {

	if playerID == 0 {
		playerID, err = getPlayerIDFromSession(r)
		if err != nil {
			return user, err
		}
	}

	user, err = db.GetUser(playerID)
	if err != nil {
		return user, err
	}

	return user, nil
}
