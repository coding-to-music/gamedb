package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/captcha"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/email"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/geo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/session"
	"github.com/go-chi/chi/v5"
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
	t.fill(w, r, "forgot", "Forgot Password", "")
	t.hideAds = true
	t.HCaptchaPublic = config.C.HCaptchaPublic
	t.LoginEmail = session.Get(r, "login-email")

	returnTemplate(w, r, t)
}

type forgotTemplate struct {
	globalTemplate
	HCaptchaPublic string
	LoginEmail     string
}

func forgotPostHandler(w http.ResponseWriter, r *http.Request) {

	const successMessage = "Email sent! (You might need to check the spam folder)"

	defer func() {
		time.Sleep(time.Second)
		session.Save(w, r)
		http.Redirect(w, r, "/forgot", http.StatusFound)
	}()

	// Field validation
	userEmail := strings.TrimSpace(r.PostFormValue("email"))

	if userEmail == "" {
		session.SetFlash(r, session.SessionBad, "Please fill in your email address")
		return
	}

	err := checkmail.ValidateFormat(userEmail)
	if err != nil {
		session.SetFlash(r, session.SessionBad, "Invalid email address")
		return
	}

	if config.IsProd() {

		resp, err := captcha.Client().CheckRequest(r)
		if err != nil {
			log.ErrS(err)
			session.SetFlash(r, session.SessionBad, "An error occurred (1002)")
			return
		}

		if !resp.Success {
			session.SetFlash(r, session.SessionBad, "Please check the captcha")
			return
		}
	}

	// Find user
	user, err := mysql.GetUserByEmail(userEmail)
	if err == mysql.ErrRecordNotFound {

		// Send email
		err = email.GetProvider().Send(
			userEmail,
			"",
			"",
			"Global Steam Forgotten Password",
			email.ForgotMissingTemplate{
				Email: userEmail,
				IP:    geo.GetFirstIP(r.RemoteAddr),
			},
		)
		if err != nil {
			log.ErrS(err)
			session.SetFlash(r, session.SessionBad, "An error occurred (1003)")
		} else {
			session.SetFlash(r, session.SessionGood, successMessage)
		}
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
	err = email.GetProvider().Send(
		userEmail,
		"",
		"",
		"Global Steam Forgotten Password",
		email.Forgot1Template{
			Domain: config.C.GlobalSteamDomain,
			Code:   code.Code,
			IP:     geo.GetFirstIP(r.RemoteAddr),
		},
	)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1003)")
		return
	}

	// Create event
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
	err = email.GetProvider().Send(
		user.Email,
		"",
		"",
		"Global Steam Password Reset",
		email.Forgot2Template{
			Password: passwordString,
			IP:       geo.GetFirstIP(r.RemoteAddr),
		},
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
