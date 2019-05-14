package pages

import (
	"net/http"
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
)

func ForgotRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", forgotHandler)
	r.Post("/", forgotPostHandler)
	r.Get("/reset", forgotResetPasswordHandler)

	return r
}

func forgotHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	t := forgotTemplate{}
	t.fill(w, r, "Forgot Password", "")
	t.RecaptchaPublic = config.Config.RecaptchaPublic.Get()
	t.setFlashes(w, r, true)

	t.LoginEmail, err = session.Get(r, "login-email")
	log.Err(err, r)

	err = returnTemplate(w, r, "forgot", t)
	log.Err(err, r)
}

type forgotTemplate struct {
	GlobalTemplate
	RecaptchaPublic string
	LoginEmail      string
}

func forgotPostHandler(w http.ResponseWriter, r *http.Request) {

	time.Sleep(time.Second)

	message, success := func() (message string, success bool) {

		// Parse form
		err := r.ParseForm()
		if err != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		email := r.PostForm.Get("email")

		// Field validation
		if email == "" {
			return "Please fill in your email address", false
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
		user, err := sql.GetUserByEmail(email)
		if err == sql.ErrRecordNotFound {
			return "Email sent", true
		} else if err != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		// todo, send email with link to below function

		// Create login event
		err = mongo.CreateUserEvent(r, user.ID, mongo.EventForgotPassword)
		if err != nil {
			log.Err(err, r)
		}

		return "Email sent", true
	}()

	//
	if success {

		err := session.SetFlash(r, helpers.SessionGood, message)
		log.Err(err, r)

	} else {

		err := session.SetFlash(r, helpers.SessionBad, message)
		log.Err(err, r)
	}

	err := session.Save(w, r)
	log.Err(err, r)

	http.Redirect(w, r, "/forgot", http.StatusFound)
}

func forgotResetPasswordHandler(w http.ResponseWriter, r *http.Request) {

	// code := r.URL.Query().Get("code")

}
