package pages

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/badoux/checkmail"
	"github.com/gamedb/gamedb/cmd/webserver/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
	"golang.org/x/crypto/bcrypt"
)

func LoginRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", loginHandler)
	r.Post("/", loginPostHandler)
	r.Post("/signup", signupPostHandler)
	r.Get("/logout", logoutHandler)
	return r
}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	_, err := getUserFromSession(r)
	if err == nil {

		err = session.SetGoodFlash(r, "Login successful")
		log.Err(err, r)

		http.Redirect(w, r, "/settings", http.StatusFound)
		return
	}

	t := loginTemplate{}
	t.fill(w, r, "Login", "Login to Game DB to set your currency and other things.")
	t.RecaptchaPublic = config.Config.RecaptchaPublic.Get()
	t.Domain = config.Config.GameDBDomain.Get()
	t.setFlashes(w, r, true)

	t.LoginEmail, err = session.Read(r, "login-email")
	log.Err(err, r)
	t.SignupEmail, err = session.Read(r, "signup-email")
	log.Err(err, r)

	err = returnTemplate(w, r, "login", t)
	log.Err(err, r)
}

type loginTemplate struct {
	GlobalTemplate
	RecaptchaPublic string
	Domain          string
	LoginEmail      string
	SignupEmail     string
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	time.Sleep(time.Second / 2)

	message, success := func() (message string, success bool) {

		// Parse form
		err := r.ParseForm()
		if err != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		email := r.PostForm.Get("email")
		password := r.PostForm.Get("password")

		// Remember email
		err = session.Write(r, "login-email", r.PostForm.Get("email"))
		if err != nil {
			log.Err(err, r)
		}

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
			err = recaptcha.CheckFromRequest(r)
			if err != nil {
				return "Please check the captcha", false
			}
		}

		// Find user
		user, err := sql.GetUser(email, true)
		if err != nil {
			log.Err(err, r)
			return "Incorrect credentials", false
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			return "Incorrect credentials", false
		}

		// Log user in
		sessionData := map[string]string{
			session.UserEmail:   user.Email,
			session.UserCountry: user.CountryCode,
		}

		player, err := mongo.GetPlayer(user.SteamID)
		if err == nil {

			sessionData[session.PlayerID] = strconv.FormatInt(player.ID, 10)
			sessionData[session.PlayerName] = player.PersonaName
			sessionData[session.PlayerLevel] = strconv.Itoa(player.Level)
		} else {
			log.Err(err, r)
		}

		err = session.WriteMany(w, r, sessionData)
		if err != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		// Create login event
		err = mongo.CreateEvent(r, player.ID, mongo.EventLogin)
		if err != nil {
			log.Err(err, r)
		}

		return "You have been logged in", true
	}()

	//
	if success {

		err := session.SetGoodFlash(r, message)
		log.Err(err, r)

		http.Redirect(w, r, "/settings", http.StatusFound)

	} else {

		err := session.SetBadFlash(r, message)
		log.Err(err, r)

		http.Redirect(w, r, "/login", http.StatusFound)
	}

	err := session.Save(w, r)
	log.Err(err, r)
}

func signupPostHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	time.Sleep(time.Second)

	message, success := func() (message string, success bool) {

		// Parse form
		err := r.ParseForm()
		if err != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		email := r.PostForm.Get("email")
		password := r.PostForm.Get("password")
		password2 := r.PostForm.Get("password2")

		// Remember email
		err = session.Write(r, "signup-email", email)
		if err != nil {
			log.Err(err, r)
		}

		// Field validation
		if email == "" {
			return "Please fill in your email address", false
		}

		if password == "" || password2 == "" {
			return "Please fill in your password", false
		}

		if password != password2 {
			return "Passwords do not match", false
		}

		err = checkmail.ValidateFormat(email)
		if err != nil {
			return "Invalid email address", false
		}

		if config.IsProd() {
			err = recaptcha.CheckFromRequest(r)
			if err != nil {
				return "Please check the captcha", false
			}
		}

		// Check user doesnt exist
		_, err = sql.GetUser(email, false)
		if err == nil {
			return "An account with this email already exists", true
		} else {
			log.Err(err, r)
		}

		// Create user
		db, err := sql.GetMySQLClient()
		if err != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
		log.Err(err, r)
		if err != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		db = db.Create(&sql.User{
			Email:         email,
			EmailVerified: false,
			Password:      string(passwordBytes),
		})

		if db.Error != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		// todo, send email

		return "Account created", true
	}()

	//
	if success {

		err := session.SetGoodFlash(r, message)
		log.Err(err, r)

		http.Redirect(w, r, "/settings", http.StatusFound)

	} else {

		err := session.SetBadFlash(r, message)
		log.Err(err, r)

		http.Redirect(w, r, "/login", http.StatusFound)
	}

	err := session.Save(w, r)
	log.Err(err, r)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	id, err := getPlayerIDFromSession(r)
	err = helpers.IgnoreErrors(err, errNotLoggedIn)
	log.Err(err, r)

	err = mongo.CreateEvent(r, id, mongo.EventLogout)
	log.Err(err, r)

	err = session.Clear(r)
	log.Err(err, r)

	err = session.SetGoodFlash(r, "You have been logged out")
	log.Err(err, r)

	err = session.Save(w, r)
	log.Err(err, r)

	http.Redirect(w, r, "/", http.StatusFound)
}
