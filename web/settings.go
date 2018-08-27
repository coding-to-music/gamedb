package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
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
			logger.Error(err)
			returnErrorTemplate(w, r, 500, err.Error())
			return
		}
	}

	//
	var wg sync.WaitGroup

	// Get logins
	var events []datastore.Event
	wg.Add(1)
	go func(player datastore.Player) {

		events, err = datastore.GetEvents(player.PlayerID, 20, datastore.EVENT_LOGIN)
		logger.Error(err)

		wg.Done()

	}(player)

	// Get donations
	var donations []datastore.Donation
	wg.Add(1)
	go func(player datastore.Player) {

		if player.Donated > 0 {
			donations, err = datastore.GetDonations(player.PlayerID, 10)
			logger.Error(err)
		}

		wg.Done()

	}(player)

	// Get games
	var games string
	wg.Add(1)
	go func(player datastore.Player) {

		resp, err := player.GetGames()
		if err != nil {
			logger.Error(err)
			return
		}
		var gamesSlice []int
		for _, v := range resp {
			gamesSlice = append(gamesSlice, v.AppID)
		}

		bytes, err := json.Marshal(gamesSlice)
		logger.Error(err)

		games = string(bytes)

		wg.Done()

	}(player)

	// Get User
	var user mysql.User
	wg.Add(1)
	go func(player datastore.Player) {

		user, err = getUser(r, 0)
		if err != nil {
			logger.Error(err)
			return
		}

		wg.Done()

	}(player)

	// Wait
	wg.Wait()

	// Template
	t := settingsTemplate{}
	t.Fill(w, r, "Settings")
	t.Events = events
	t.Player = player
	t.User = user
	t.Donations = donations
	t.Games = games

	returnTemplate(w, r, "settings", t)
}

func SettingsPostHandler(w http.ResponseWriter, r *http.Request) {

	// Get user
	user, err := getUser(r, 0)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		logger.Error(err)
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
				logger.Error(err)
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

	err = user.SaveOrUpdateUser()
	if err != nil {
		logger.Error(err)
		session.SetBadFlash(w, r, "Something went wrong saving settings")
	} else {
		session.SetGoodFlash(w, r, "Settings saved")
	}

	http.Redirect(w, r, "/settings", 302)
	return

}

type settingsTemplate struct {
	GlobalTemplate
	Player    datastore.Player
	User      mysql.User
	Events    []datastore.Event
	Donations []datastore.Donation
	Games     string
	Messages  []interface{}
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

func getPlayer(r *http.Request, playerID int64) (player datastore.Player, err error) {

	if playerID == 0 {
		playerID, err = getPlayerIDFromSession(r)
		if err != nil {
			return player, err
		}
	}

	player, err = datastore.GetPlayer(playerID)
	if err != nil {
		return player, err
	}

	return player, nil
}

func getUser(r *http.Request, playerID int64) (user mysql.User, err error) {

	if playerID == 0 {
		playerID, err = getPlayerIDFromSession(r)
		if err != nil {
			return user, err
		}
	}

	user, err = mysql.GetUser(playerID)
	if err != nil {
		return user, err
	}

	return user, nil
}
