package pages

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/Jleagle/session-go/session"
	"github.com/badoux/checkmail"
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
	return r
}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	_, err := getUserFromSession(r)
	if err == nil {

		err = session.SetFlash(r, helpers.SessionGood, "Login successful")
		log.Err(err, r)

		http.Redirect(w, r, "/settings", http.StatusFound)
		return
	}

	t := loginTemplate{}
	t.fill(w, r, "Login", "Login to Game DB to set your currency and other things.")
	t.RecaptchaPublic = config.Config.RecaptchaPublic.Get()
	t.setFlashes(w, r, true)

	t.LoginEmail, err = session.Get(r, "login-email")
	log.Err(err, r)

	err = returnTemplate(w, r, "login", t)
	log.Err(err, r)
}

type loginTemplate struct {
	GlobalTemplate
	RecaptchaPublic string
	LoginEmail      string
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {

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

		// Remember email
		err = session.Set(r, "login-email", r.PostForm.Get("email"))
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
		user, err := sql.GetUserByKey("email", email, 0)
		if err != nil {
			err = helpers.IgnoreErrors(err, sql.ErrRecordNotFound)
			log.Err(err, r)
			return "Incorrect credentials", false
		}

		if !user.EmailVerified {
			return "Please verify your email address first", false
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			return "Incorrect credentials", false
		}

		// Log user in
		sessionData := map[string]string{
			helpers.SessionUserID:         strconv.Itoa(user.ID),
			helpers.SessionUserEmail:      user.Email,
			helpers.SessionUserCountry:    user.CountryCode,
			helpers.SessionUserShowAlerts: strconv.FormatBool(user.ShowAlerts),
		}

		player, err := mongo.GetPlayer(user.SteamID)
		if err == nil {
			sessionData[helpers.SessionPlayerID] = strconv.FormatInt(player.ID, 10)
			sessionData[helpers.SessionPlayerName] = player.PersonaName
			sessionData[helpers.SessionPlayerLevel] = strconv.Itoa(player.Level)
		} else {
			err = helpers.IgnoreErrors(err, mongo.ErrInvalidPlayerID, mongo.ErrNoDocuments)
			log.Err(err, r)
		}

		err = session.SetMany(r, sessionData)
		if err != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		// Create login event
		err = mongo.CreateUserEvent(r, user.ID, mongo.EventLogin)
		if err != nil {
			log.Err(err, r)
		}

		return "You have been logged in", true
	}()

	//
	if success {

		err := session.SetFlash(r, helpers.SessionGood, message)
		log.Err(err, r)

		err = session.Save(w, r)
		log.Err(err, r)

		http.Redirect(w, r, "/settings", http.StatusFound)

	} else {

		err := session.SetFlash(r, helpers.SessionBad, message)
		log.Err(err, r)

		err = session.Save(w, r)
		log.Err(err, r)

		http.Redirect(w, r, "/login", http.StatusFound)
	}
}
