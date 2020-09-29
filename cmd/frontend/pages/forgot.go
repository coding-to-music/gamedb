package pages

import (
	"net/http"
	"strings"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/badoux/checkmail"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/email_providers"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/geo"
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

	const successMessage = "Email sent! (You might need to check the spam folder)"

	defer func() {
		time.Sleep(time.Second)
		session.Save(w, r)
		http.Redirect(w, r, "/forgot", http.StatusFound)
	}()

	// Field validation
	email := strings.TrimSpace(r.PostFormValue("email"))

	if email == "" {
		session.SetFlash(r, session.SessionBad, "Please fill in your email address")
		return
	}

	err := checkmail.ValidateFormat(email)
	if err != nil {
		session.SetFlash(r, session.SessionBad, "Invalid email address")
		return
	}

	if config.IsProd() {
		err = recaptcha.CheckFromRequest(r)
		if err != nil {
			session.SetFlash(r, session.SessionBad, "Please check the captcha")
			return
		}
	}

	// Find user
	user, err := mysql.GetUserByEmail(email)
	if err == mysql.ErrRecordNotFound {
		session.SetFlash(r, session.SessionGood, successMessage)
		return
	} else if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1001)")
		return
	}

	// Create verification code
	code, err := mysql.CreateUserVerification(user.ID)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1002)")
		return
	}

	// Send email
	body := "You or someone else has requested a new password for Game DB.<br><br>" +
		"If this was not you, please ignore this email.<br><br>Click the following link to reset your password: " +
		config.C.GameDBDomain + "/forgot/reset?code=" + code.Code +
		"<br><br>Thanks, Jleagle." +
		"<br><br>From IP: " + geo.GetFirstIP(r.RemoteAddr)

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
		session.SetFlash(r, session.SessionBad, "An error occurred (1003)")
		return
	}

	// Create login event
	err = mongo.NewEvent(r, user.ID, mongo.EventForgotPassword)
	if err != nil {
		log.ErrS(err)
	}

	session.SetFlash(r, session.SessionGood, successMessage)
}

func forgotResetPasswordHandler(w http.ResponseWriter, r *http.Request) {

	var success bool

	defer func() {

		time.Sleep(time.Second)
		session.Save(w, r)

		if success {
			http.Redirect(w, r, "/login", http.StatusFound)
		} else {
			http.Redirect(w, r, "/forgot", http.StatusFound)
		}
	}()

	// Validate code
	code := strings.TrimSpace(r.URL.Query().Get("code"))

	if len(code) != 10 {
		session.SetFlash(r, session.SessionBad, "Invalid code (1001)")
		return
	}

	// Find email from code
	userID, err := mysql.GetUserVerification(code)
	if err == mysql.ErrExpiredVerification {
		session.SetFlash(r, session.SessionBad, "Link Expired")
		return
	} else if err != nil {
		err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
		if err != nil {
			log.ErrS(err)
		}
		session.SetFlash(r, session.SessionBad, "Invalid code (1002)")
		return
	}

	// Get user
	user, err := mysql.GetUserByID(userID)
	if err != nil {
		err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
		if err != nil {
			log.ErrS(err)
		}
		session.SetFlash(r, session.SessionBad, "An error occurred (1001)")
		return
	}

	// Create password
	passwordString := helpers.RandString(10, helpers.LettersCaps+helpers.Numbers)
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(passwordString), 14)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1002)")
		return
	}

	// Send email
	body := "Your new Game DB password is:<br><br>" + passwordString + "<br><br>Thanks, Jleagle." +
		"<br><br>From IP: " + geo.GetFirstIP(r.RemoteAddr)

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
		session.SetFlash(r, session.SessionBad, "An error occurred (1003)")
		return
	}

	// Set password
	err = user.SetPassword(passwordBytes)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1004)")
		return
	}

	//
	success = true
	session.SetFlash(r, session.SessionGood, "A new password has been emailed to you (You might need to check the spam folder)")
}
