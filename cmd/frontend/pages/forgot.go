package pages

import (
	"net/http"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/badoux/checkmail"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/email_providers"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
	"golang.org/x/crypto/bcrypt"
)

func ForgotRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", forgotHandler)
	r.Post("/", forgotPostHandler)
	r.Get("/reset", forgotResetPasswordHandler)

	return r
}

func forgotHandler(w http.ResponseWriter, r *http.Request) {

	if session.IsLoggedIn(r) {

		http.Redirect(w, r, "/settings", http.StatusFound)
		return
	}

	t := forgotTemplate{}
	t.fill(w, r, "Forgot Password", "")
	t.hideAds = true
	t.RecaptchaPublic = config.C.RecaptchaPublic
	t.LoginEmail = session.Get(r, "login-email")

	returnTemplate(w, r, "forgot", t)
}

type forgotTemplate struct {
	globalTemplate
	RecaptchaPublic string
	LoginEmail      string
}

func (t forgotTemplate) includes() []string {
	return []string{"includes/login_header.gohtml"}
}

func forgotPostHandler(w http.ResponseWriter, r *http.Request) {

	message, success := func() (message string, success bool) {

		// Parse form
		err := r.ParseForm()
		if err != nil {
			log.ErrS(err)
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
		user, err := mysql.GetUserByKey("email", email, 0)
		if err == mysql.ErrRecordNotFound {
			return "Email sent", true
		} else if err != nil {
			log.ErrS(err)
			return "An error occurred", false
		}

		// Create verification code
		code, err := mysql.CreateUserVerification(user.ID)
		if err != nil {
			log.ErrS(err)
			return "An error occurred", false
		}

		// Send email
		body := "You or someone else has requested a new password for Game DB.<br><br>" +
			"If this was not you, please ignore this email.<br><br>Click the following link to reset your password: " +
			config.C.GameDBDomain + "/forgot/reset?code=" + code.Code +
			"<br><br>Thanks, Jleagle." +
			"<br><br>From IP: " + r.RemoteAddr

		err = email_providers.GetSender().Send(
			email,
			email,
			"",
			"",
			"Game DB Forgotten Password",
			body,
		)
		if err != nil {
			log.ErrS(err)
			return "An error occurred", false
		}

		// Create login event
		err = mongo.CreateUserEvent(r, user.ID, mongo.EventForgotPassword)
		if err != nil {
			log.ErrS(err)
		}

		return "Email sent", true
	}()

	//
	if success {

		session.SetFlash(r, session.SessionGood, message)
		session.Save(w, r)

		http.Redirect(w, r, "/login", http.StatusFound)

	} else {

		time.Sleep(time.Second)

		session.SetFlash(r, session.SessionBad, message)
		session.Save(w, r)

		http.Redirect(w, r, "/forgot", http.StatusFound)
	}
}

func forgotResetPasswordHandler(w http.ResponseWriter, r *http.Request) {

	message, success := func() (message string, success bool) {

		// Validate code
		code := r.URL.Query().Get("code")

		if len(code) != 10 {
			return "Invalid code (1001)", false
		}

		// Find email from code
		userID, err := mysql.GetUserVerification(code)
		if err != nil {
			err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
			if err != nil {
				log.ErrS(err)
			}
			return "Invalid code (1002)", false
		}

		// if userVerify.Expires.Unix() < time.Now().Unix() {
		// return "This verify code has expired", false
		// }

		// Get user
		user, err := mysql.GetUserByID(userID)
		if err != nil {
			err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
			if err != nil {
				log.ErrS(err)
			}
			return "An error occurred (1001)", false
		}

		// Create password
		passwordString := helpers.RandString(10, helpers.LettersCaps+helpers.Numbers)
		passwordBytes, err := bcrypt.GenerateFromPassword([]byte(passwordString), 14)
		if err != nil {
			log.ErrS(err)
			return "An error occurred (1002)", false
		}

		// Send email
		body := "Your new Game DB password is:<br><br>" + passwordString + "<br><br>Thanks, Jleagle." +
			"<br><br>From IP: " + r.RemoteAddr

		err = email_providers.GetSender().Send(
			user.Email,
			user.Email,
			"",
			"",
			"Game DB Password Reset",
			body,
		)
		if err != nil {
			log.ErrS(err)
			return "An error occurred", false
		}

		// Set password
		err = mysql.UpdateUserCol(userID, "password", string(passwordBytes))
		if err != nil {
			log.ErrS(err)
			return "An error occurred (1003)", false
		}

		//
		return "A new password has been emailed to you", true
	}()

	//
	if success {

		session.SetFlash(r, session.SessionGood, message)
		session.Save(w, r)

		http.Redirect(w, r, "/login", http.StatusFound)

	} else {

		time.Sleep(time.Second)

		session.SetFlash(r, session.SessionBad, message)
		session.Save(w, r)

		http.Redirect(w, r, "/signup", http.StatusFound)
	}
}
