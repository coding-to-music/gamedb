package pages

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/Jleagle/session-go/session"
	"github.com/Jleagle/steam-go/steamid"
	"github.com/badoux/checkmail"
	sessionHelpers "github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/cmd/webserver/pages/oauth"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
	"golang.org/x/crypto/bcrypt"
)

const loginSessionEmail = "login-email"

func LoginRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", loginHandler)
	r.Post("/", loginPostHandler)

	r.Get("/oauth/{id:[a-z]+}", oauthLoginHandler)
	r.Get("/oauth-callback/{id:[a-z]+}", oauthLCallbackHandler)

	return r
}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	_, err := getUserFromSession(r)
	if err == nil {

		err = session.SetFlash(r, sessionHelpers.SessionGood, "Login successful")
		log.Err(err, r)

		http.Redirect(w, r, "/settings", http.StatusFound)
		return
	}

	t := loginTemplate{}
	t.fill(w, r, "Login", "Login to Game DB")
	t.hideAds = true
	t.RecaptchaPublic = config.Config.RecaptchaPublic.Get()
	t.LoginEmail = sessionHelpers.Get(r, loginSessionEmail)

	returnTemplate(w, r, "login", t)
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
		err = session.Set(r, loginSessionEmail, r.PostForm.Get("email"))
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

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			err = helpers.IgnoreErrors(err, bcrypt.ErrMismatchedHashAndPassword)
			log.Err(err, r)
			return "Incorrect credentials", false
		}

		return login(r, user)
	}()

	//
	if success {

		err := session.SetFlash(r, sessionHelpers.SessionGood, message)
		log.Err(err, r)

		sessionHelpers.Save(w, r)

		// Get last page
		val := sessionHelpers.Get(r, sessionHelpers.SessionLastPage)
		if val == "" {
			val = "/settings"
		}

		//
		http.Redirect(w, r, val, http.StatusFound)

	} else {

		err := session.SetFlash(r, sessionHelpers.SessionBad, message)
		log.Err(err, r)

		sessionHelpers.Save(w, r)

		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func login(r *http.Request, user sql.User) (string, bool) {

	if !user.EmailVerified {
		return "Please verify your email address first", false
	}

	// Log user in
	sessionData := map[string]string{
		sessionHelpers.SessionUserID:         strconv.Itoa(user.ID),
		sessionHelpers.SessionUserEmail:      user.Email,
		sessionHelpers.SessionUserProdCC:     string(user.ProductCC),
		sessionHelpers.SessionUserAPIKey:     user.APIKey,
		sessionHelpers.SessionUserShowAlerts: strconv.FormatBool(user.ShowAlerts),
		sessionHelpers.SessionUserLevel:      strconv.Itoa(int(user.Level)),
	}

	steamID := user.GetSteamID()
	if steamID > 0 {
		player, err := mongo.GetPlayer(steamID)
		if err == nil {
			sessionData[sessionHelpers.SessionPlayerID] = strconv.FormatInt(player.ID, 10)
			sessionData[sessionHelpers.SessionPlayerName] = player.GetName()
			sessionData[sessionHelpers.SessionPlayerLevel] = strconv.Itoa(player.Level)
		} else {
			err = helpers.IgnoreErrors(err, steamid.ErrInvalidPlayerID, mongo.ErrNoDocuments)
			log.Err(err, r)
		}
	}

	sessionHelpers.SetMany(r, sessionData)

	// Create login event
	err := mongo.CreateUserEvent(r, user.ID, mongo.EventLogin)
	if err != nil {
		log.Err(err, r)
	}

	return "You have been logged in", true
}

func oauthLoginHandler(w http.ResponseWriter, r *http.Request) {

	id := oauth.ConnectionEnum(chi.URLParam(r, "id"))

	if _, ok := oauth.Connections[id]; ok {
		connection := oauth.New(id)
		connection.LoginHandler(w, r)
	}
}

func oauthLCallbackHandler(w http.ResponseWriter, r *http.Request) {

	id := oauth.ConnectionEnum(chi.URLParam(r, "id"))

	if _, ok := oauth.Connections[id]; ok {
		connection := oauth.New(id)
		connection.LoginCallbackHandler(w, r)
	}
}
