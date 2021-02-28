package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steamid"
	"github.com/badoux/checkmail"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/captcha"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/oauth"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

const loginSessionEmail = "login-email"

func LoginRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", loginHandler)
	r.Post("/", loginPostHandler)

	return r
}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	if session.IsLoggedIn(r) {
		http.Redirect(w, r, "/settings", http.StatusFound)
		return
	}

	t := loginTemplate{}
	t.fill(w, r, "login", "Login", "Login to Global Steam")
	t.hideAds = true
	t.HCaptchaPublic = config.C.HCaptchaPublic
	t.LoginEmail = session.Get(r, loginSessionEmail)
	t.Providers = oauth.Providers

	returnTemplate(w, r, t)
}

type loginTemplate struct {
	globalTemplate
	HCaptchaPublic string
	LoginEmail     string
	Providers      []oauth.Provider
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {

	message, success := func() (message string, success bool) {

		// Parse form
		err := r.ParseForm()
		if err != nil {
			log.ErrS(err)
			return "An error occurred", false
		}

		email := r.PostForm.Get("email")
		password := r.PostForm.Get("password")

		// Remember email
		session.Set(r, loginSessionEmail, r.PostForm.Get("email"))

		// Field validation
		if email == "" {
			return "Please fill in your email address", false
		}

		if password == "" {
			return "Please fill in your password", false
		}

		err = checkmail.ValidateFormat(email)
		if err != nil {
			return "Invalid email address", false
		}

		if config.IsProd() {

			resp, err := captcha.GetCaptcha().CheckRequest(r)
			if err != nil {
				log.ErrS(err)
				return "Something went wrong", false
			}

			if !resp.Success {
				return "Please check the captcha", false
			}
		}

		// Find user
		user, err := mysql.GetUserByEmail(email)
		if err != nil {
			err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
			if err != nil {
				log.ErrS(err)
			}
			return "Incorrect credentials", false
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			err = helpers.IgnoreErrors(err, bcrypt.ErrMismatchedHashAndPassword)
			if err != nil {
				log.ErrS(err)
			}
			return "Incorrect credentials", false
		}

		return login(r, user)
	}()

	//
	if success {

		session.SetFlash(r, session.SessionGood, message)
		session.Save(w, r)

		// Get last page
		val := session.Get(r, session.SessionLastPage)
		if val == "" {
			val = "/settings"
		}

		//
		http.Redirect(w, r, val, http.StatusFound)

	} else {

		time.Sleep(time.Second)

		session.SetFlash(r, session.SessionBad, message)
		session.Save(w, r)

		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func login(r *http.Request, user mysql.User) (string, bool) {

	if !user.EmailVerified {
		return "Please verify your email address first", false
	}

	// Log user in
	session.SetMany(r, map[string]string{
		session.SessionUserID:     strconv.Itoa(user.ID),
		session.SessionUserEmail:  user.Email,
		session.SessionUserProdCC: string(user.ProductCC),
		session.SessionUserAPIKey: user.APIKey,
		session.SessionUserLevel:  strconv.Itoa(int(user.Level)),
		// session.SessionUserShowAlerts: strconv.FormatBool(user.ShowAlerts),
	})

	playerID := mysql.GetUserSteamID(user.ID)
	if playerID > 0 {
		player, err := mongo.GetPlayer(playerID)
		if err == nil {
			session.SetMany(r, map[string]string{
				session.SessionPlayerID:    strconv.FormatInt(player.ID, 10),
				session.SessionPlayerName:  player.GetName(),
				session.SessionPlayerLevel: strconv.Itoa(player.Level),
			})
		} else {
			err = helpers.IgnoreErrors(err, steamid.ErrInvalidPlayerID, mongo.ErrNoDocuments)
			if err != nil {
				log.ErrS(err)
			}
		}
	}

	// Create login event
	err := mongo.NewEvent(r, user.ID, mongo.EventLogin)
	if err != nil {
		log.ErrS(err)
	}

	err = user.TouchLoggedInTime()
	if err != nil {
		log.ErrS(err)
	}

	return "You have been logged in", true
}
