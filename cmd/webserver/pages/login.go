package pages

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/Jleagle/session-go/session"
	"github.com/badoux/checkmail"
	webserverHelpers "github.com/gamedb/gamedb/cmd/webserver/helpers"
	"github.com/gamedb/gamedb/cmd/webserver/oauth"
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

		err = session.SetFlash(r, webserverHelpers.SessionGood, "Login successful")
		log.Err(err, r)

		http.Redirect(w, r, "/settings", http.StatusFound)
		return
	}

	t := loginTemplate{}
	t.fill(w, r, "Login", "Login to Game DB")
	t.hideAds = true
	t.RecaptchaPublic = config.Config.RecaptchaPublic.Get()

	t.LoginEmail, err = session.Get(r, loginSessionEmail)
	log.Err(err, r)

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
			log.Err(err)
			return "Incorrect credentials", false
		}

		return login(r, user)
	}()

	//
	if success {

		err := session.SetFlash(r, webserverHelpers.SessionGood, message)
		log.Err(err, r)

		err = session.Save(w, r)
		log.Err(err, r)

		// Get last page
		val, err := session.Get(r, webserverHelpers.SessionLastPage)
		if err != nil {
			log.Err(err, r)
		}

		if val == "" {
			val = "/settings"
		}

		//
		http.Redirect(w, r, val, http.StatusFound)

	} else {

		err := session.SetFlash(r, webserverHelpers.SessionBad, message)
		log.Err(err, r)

		err = session.Save(w, r)
		log.Err(err, r)

		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func login(r *http.Request, user sql.User) (string, bool) {

	if !user.EmailVerified {
		return "Please verify your email address first", false
	}

	// Log user in
	sessionData := map[string]string{
		webserverHelpers.SessionUserID:         strconv.Itoa(user.ID),
		webserverHelpers.SessionUserEmail:      user.Email,
		webserverHelpers.SessionUserProdCC:     string(user.ProductCC),
		webserverHelpers.SessionUserAPIKey:     user.APIKey,
		webserverHelpers.SessionUserShowAlerts: strconv.FormatBool(user.ShowAlerts),
		webserverHelpers.SessionUserLevel:      strconv.Itoa(int(user.PatreonLevel)),
	}

	steamID := user.GetSteamID()
	if steamID > 0 {
		player, err := mongo.GetPlayer(steamID)
		if err == nil {
			sessionData[webserverHelpers.SessionPlayerID] = strconv.FormatInt(player.ID, 10)
			sessionData[webserverHelpers.SessionPlayerName] = player.PersonaName
			sessionData[webserverHelpers.SessionPlayerLevel] = strconv.Itoa(player.Level)
		} else {
			err = helpers.IgnoreErrors(err, helpers.ErrInvalidPlayerID, mongo.ErrNoDocuments)
			log.Err(err, r)
		}
	}

	err := session.SetMany(r, sessionData)
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
