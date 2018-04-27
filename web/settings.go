package web

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/session"
	"golang.org/x/crypto/bcrypt"
)

func SettingsHandler(w http.ResponseWriter, r *http.Request) {

	loggedIn, err := session.IsLoggedIn(r)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	if !loggedIn {
		http.Redirect(w, r, "/login", 302)
		return
	}

	// Get session
	id, err := session.Read(r, session.UserID)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Convert ID
	idx, err := strconv.Atoi(id)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Get player
	player, err := datastore.GetPlayer(idx)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Save form data
	if r.Method == "POST" {

		// Parse form
		if err := r.ParseForm(); err != nil {
			logger.Error(err)
			returnErrorTemplate(w, r, 500, err.Error())
			return
		}

		// Save password
		password := r.PostForm.Get("password")

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
			}

			player.SettingsPassword = string(passwordBytes)
		}

		// Save email
		player.SettingsEmail = r.PostForm.Get("email")

		// Save hidden
		if r.PostForm.Get("hide") == "1" {
			player.SettingsHidden = true
		} else {
			player.SettingsHidden = false
		}

		// Save alerts
		if r.PostForm.Get("alerts") == "1" {
			player.SettingsAlerts = true
		} else {
			player.SettingsAlerts = false
		}

		err = player.Save()
		if err != nil {
			logger.Error(err)
			session.SetBadFlash(w, r, "Something went wrong saving settings")
		} else {
			session.SetGoodFlash(w, r, "Settings saved")
		}

		http.Redirect(w, r, "/settings", 302)
		return
	}

	// Get logins
	logins, err := datastore.GetLogins(idx, 20)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Get donations
	var donations []datastore.Donation
	if player.Donated > 0 {
		donations, err = datastore.GetDonations(player.PlayerID, 10)
		if err != nil {
			logger.Error(err)
			returnErrorTemplate(w, r, 500, err.Error())
			return
		}
	}

	// Get games
	games := player.GetGames()
	var gamesSlice []int
	for _, v := range games {
		gamesSlice = append(gamesSlice, v.AppID)
	}

	gamesString, err := json.Marshal(gamesSlice)
	if err != nil {
		logger.Error(err)
	}

	// Template
	template := settingsTemplate{}
	template.Fill(w, r, "Settings")
	template.Logins = logins
	template.Player = *player
	template.Donations = donations
	template.Games = string(gamesString)

	returnTemplate(w, r, "settings", template)
}

type settingsTemplate struct {
	GlobalTemplate
	Player    datastore.Player
	Logins    []datastore.Login
	Donations []datastore.Donation
	Games     string
	Messages  []interface{}
}
