package pages

import (
	"net/http"
	"strings"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/badoux/checkmail"
	"github.com/gamedb/gamedb/cmd/webserver/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
	"github.com/nlopes/slack"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"golang.org/x/crypto/bcrypt"
)

func SignupRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", signupHandler)
	r.Post("/", signupPostHandler)
	r.Get("/verify", verifyHandler)
	return r
}

func signupHandler(w http.ResponseWriter, r *http.Request) {

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

	t := signupTemplate{}
	t.fill(w, r, "Login", "Login to Game DB to set your currency and other things.")
	t.RecaptchaPublic = config.Config.RecaptchaPublic.Get()
	t.Domain = config.Config.GameDBDomain.Get()
	t.setFlashes(w, r, true)

	t.SignupEmail, err = session.Read(r, "signup-email")
	log.Err(err, r)

	err = returnTemplate(w, r, "signup", t)
	log.Err(err, r)
}

type signupTemplate struct {
	GlobalTemplate
	RecaptchaPublic string
	Domain          string
	SignupEmail     string
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

		if len(password) < 8 {
			return "Password must be at least 8 characters", false
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
		_, err = sql.GetUserByEmail(email)
		if err == nil {
			return "An account with this email already exists", true
		} else {
			err = helpers.IgnoreErrors(err, sql.ErrRecordNotFound)
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
			CountryCode:   string(steam.CountryUS),
		})

		if db.Error != nil {
			log.Err(db.Error, r)
			return "An error occurred", false
		}

		// Create verification code
		code, err := sql.CreateUserVerification(email)
		if err != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		// Send email
		var link = func() string {
			if config.IsLocal() {
				return "http://localhost:" + config.Config.WebserverPort.Get() + "/signup/verify?code=" + code.Code
			}
			return "https://gamedb.online/signup/verify?code=" + code.Code
		}()

		body := "Please click the below link to verify your email address\n" + link

		client := sendgrid.NewSendClient(config.Config.SendGridAPIKey.Get())
		_, err = client.Send(mail.NewSingleEmail(
			mail.NewEmail("Game DB", "no-reply@gamedb.online"),
			"Game DB Email Verification",
			mail.NewEmail(email, email),
			body,
			strings.ReplaceAll(body, "\n", "<br />"),
		))
		if err != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		// Slack message
		err = slack.PostWebhook(config.Config.SlackWebhook.Get(), &slack.WebhookMessage{
			Text: "New signup: " + email,
		})
		log.Err(err)

		return "Please check your email to verify your email", true
	}()

	//
	if success {

		err := session.SetGoodFlash(r, message)
		log.Err(err, r)

		err = session.Save(w, r)
		log.Err(err, r)

		http.Redirect(w, r, "/login", http.StatusFound)

	} else {

		err := session.SetBadFlash(r, message)
		log.Err(err, r)

		err = session.Save(w, r)
		log.Err(err, r)

		http.Redirect(w, r, "/signup", http.StatusFound)
	}
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{"code"})
	if ret {
		return
	}

	time.Sleep(time.Second)

	message, success := func() (message string, success bool) {

		// Validate code
		code := r.URL.Query().Get("code")

		if len(code) != 10 {
			return "Invalid code (1001)", false
		}

		// Find email from code
		db, err := sql.GetMySQLClient()
		if err != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		var userVerify sql.UserVerification
		db = db.Where("code = ?", code).Find(&userVerify)
		if db.Error != nil {
			err = helpers.IgnoreErrors(err, sql.ErrRecordNotFound)
			log.Err(db.Error, r)
			return "Invalid code (1002)", false
		}

		// if userVerify.Expires.Unix() < time.Now().Unix() {
		// return "This verify code has expired", false
		// }

		// Enable user
		db, err = sql.GetMySQLClient()
		if err != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		var user = sql.User{Email: userVerify.Email}

		db = db.Model(&user).Update("email_verified", true)
		if db.Error != nil {
			log.Err(db.Error, r)
			return "Invalid code (1003)", false
		}

		//
		return "Email has been verified", true
	}()

	//
	if success {

		err := session.SetGoodFlash(r, message)
		log.Err(err, r)

		err = session.Save(w, r)
		log.Err(err, r)

		http.Redirect(w, r, "/login", http.StatusFound)

	} else {

		err := session.SetBadFlash(r, message)
		log.Err(err, r)

		err = session.Save(w, r)
		log.Err(err, r)

		http.Redirect(w, r, "/signup", http.StatusFound)
	}
}
