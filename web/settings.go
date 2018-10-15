package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/logging"
	"github.com/steam-authority/steam-authority/session"
	"golang.org/x/crypto/bcrypt"
)

var (
	errNotLoggedIn = errors.New("not logged in")
)

func SettingsHandler(w http.ResponseWriter, r *http.Request) {

	player, err := getPlayer(r, 0)
	if err != nil {
		if err == errNotLoggedIn {
			session.SetBadFlash(w, r, "please login")
			http.Redirect(w, r, "/login", 302)
			return
		} else {
			logging.Error(err)
			returnErrorTemplate(w, r, 500, err.Error())
			return
		}
	}

	//
	var wg sync.WaitGroup

	// Get donations
	var donations []db.Donation
	wg.Add(1)
	go func(player db.Player) {

		if player.Donated > 0 {
			donations, err = db.GetDonations(player.PlayerID, 10)
			logging.Error(err)
		}

		wg.Done()

	}(player)

	// Get games
	var games string
	wg.Add(1)
	go func(player db.Player) {

		resp, err := player.GetAllPlayerApps("app_name", 0)
		if err != nil {
			logging.Error(err)
			return
		}
		var gamesSlice []int
		for _, v := range resp {
			gamesSlice = append(gamesSlice, v.AppID)
		}

		bytes, err := json.Marshal(gamesSlice)
		logging.Error(err)

		games = string(bytes)

		wg.Done()

	}(player)

	// Get User
	var user db.User
	wg.Add(1)
	go func(player db.Player) {

		user, err = getUser(r, 0)
		if err != nil {
			logging.Error(err)
			return
		}

		wg.Done()

	}(player)

	// Wait
	wg.Wait()

	// Template
	t := settingsTemplate{}
	t.Fill(w, r, "Settings")
	t.Player = player
	t.User = user
	t.Donations = donations
	t.Games = games

	returnTemplate(w, r, "settings", t)
}

type settingsTemplate struct {
	GlobalTemplate
	Player    db.Player
	User      db.User
	Donations []db.Donation
	Games     string
	Messages  []interface{}
}

func SettingsPostHandler(w http.ResponseWriter, r *http.Request) {

	// Get user
	user, err := getUser(r, 0)
	if err != nil {
		logging.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		logging.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Save password
	password := r.PostForm.Get("password")

	if len(password) > 0 {
		if len(password) < 8 {
			session.SetBadFlash(w, r, "Password must be at least 8 characters long")
			http.Redirect(w, r, "/settings", 302)
			return
		} else {
			passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
			if err != nil {
				logging.Error(err)
				session.SetBadFlash(w, r, "Something went wrong encrypting your password")
				http.Redirect(w, r, "/settings", 302)
				return
			} else {
				user.Password = string(passwordBytes)
			}
		}
	}

	// Save email
	user.Email = r.PostForm.Get("email")

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
	_, err = user.UpdateInsert()
	if err != nil {
		logging.Error(err)
		session.SetBadFlash(w, r, "Something went wrong saving settings")
	} else {
		session.SetGoodFlash(w, r, "Settings saved")
	}

	http.Redirect(w, r, "/settings", 302)
	return

}

func SettingsEventsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	query.FillFromURL(r.URL.Query())

	//
	var wg sync.WaitGroup

	// Get events
	var events []db.Event

	wg.Add(1)
	go func() {

		playerID, err := getPlayerIDFromSession(r)
		if err != nil {

			logging.Error(err)

		} else {

			client, ctx, err := db.GetDSClient()
			if err != nil {

				logging.Error(err)

			} else {

				q := datastore.NewQuery(db.KindEvent).Filter("player_id =", playerID).Limit(100)
				q, err = query.SetOrderOffsetDS(q, map[string]string{})
				q = q.Order("-created_at")
				if err != nil {

					logging.Error(err)

				} else {

					_, err := client.GetAll(ctx, q, &events)
					logging.Error(err)
				}
			}
		}

		wg.Done()
	}()

	// Get total
	var total int
	wg.Add(1)
	go func() {

		playerID, err := getPlayerIDFromSession(r)
		if err != nil {

			logging.Error(err)

		} else {

			total, err = db.CountPlayerEvents(playerID)
			logging.Error(err)

		}

		wg.Done()
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(total)
	response.RecordsFiltered = strconv.Itoa(total)
	response.Draw = query.Draw

	for _, v := range events {
		response.AddRow(v.OutputForJSON(r))
	}

	response.output(w)

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
	idx, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return playerID, err
	}

	return idx, nil
}

func getPlayer(r *http.Request, playerID int64) (player db.Player, err error) {

	if playerID == 0 {
		playerID, err = getPlayerIDFromSession(r)
		if err != nil {
			return player, err
		}
	}

	player, err = db.GetPlayer(playerID)
	if err != nil {
		return player, err
	}

	return player, nil
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
