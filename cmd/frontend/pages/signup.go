package pages

import (
	"net/http"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/badoux/checkmail"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/email"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/oauth"
	"github.com/go-chi/chi"
	influx "github.com/influxdata/influxdb1-client"
)

const signupSessionEmail = "signup-email"

func SignupRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", signupHandler)
	r.Post("/", signupPostHandler)
	r.Get("/verify", verifyHandler)
	return r
}

func signupHandler(w http.ResponseWriter, r *http.Request) {

	if session.IsLoggedIn(r) {
		http.Redirect(w, r, "/settings", http.StatusFound)
		return
	}

	t := signupTemplate{}
	t.fill(w, r, "signup", "Login", "Login to Game DB to set your currency and other things.")
	t.addAssetPasswordStrength()
	t.hideAds = true
	t.Domain = config.C.GameDBDomain
	t.RecaptchaPublic = config.C.RecaptchaPublic
	t.SignupEmail = session.Get(r, signupSessionEmail)
	t.Providers = oauth.Providers

	returnTemplate(w, r, t)
}

type signupTemplate struct {
	globalTemplate
	RecaptchaPublic string
	Domain          string
	SignupEmail     string
	Providers       []oauth.Provider
}

func signupPostHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		time.Sleep(time.Second)
		session.Save(w, r)
		http.Redirect(w, r, "/signup", http.StatusFound)
	}()

	// Parse form
	err := r.ParseForm()
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1001)")
		return
	}

	userEmail := r.PostForm.Get("email")
	password := r.PostForm.Get("password")
	password2 := r.PostForm.Get("password2")

	// Remember email
	session.Set(r, signupSessionEmail, userEmail)

	// Field validation
	if userEmail == "" {
		session.SetFlash(r, session.SessionBad, "Please fill in your email address")
		return
	}

	if password == "" || password2 == "" {
		session.SetFlash(r, session.SessionBad, "Please fill in your password")
		return
	}

	if len(password) < 8 {
		session.SetFlash(r, session.SessionBad, "Password must be at least 8 characters")
		return
	}

	if password != password2 {
		session.SetFlash(r, session.SessionBad, "Passwords do not match")
		return
	}

	err = checkmail.ValidateFormat(userEmail)
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

	// Check user doesnt exist already
	_, err = mysql.GetUserByEmail(userEmail)
	if err == nil {
		session.SetFlash(r, session.SessionBad, "An account with this email already exists")
		return
	}

	err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
	if err != nil {
		log.ErrS(err)
	}

	// Create user
	_, err = mysql.NewUser(r, userEmail, password, session.GetProductCC(r), false)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "An error occurred (1002)")
		return
	}

	session.SetFlash(r, session.SessionGood, "Please check your email to verify your account (You might need to check the spam folder)")
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		time.Sleep(time.Second)
		session.Save(w, r)
		http.Redirect(w, r, "/signup", http.StatusFound)
	}()

	// Validate code
	code := r.URL.Query().Get("code")

	if len(code) != 10 {
		session.SetFlash(r, session.SessionBad, "Invalid code (1001)")
		return
	}

	// Find email from code
	userID, err := mysql.GetUserVerification(code)
	if err == mysql.ErrExpiredVerification {
		session.SetFlash(r, session.SessionBad, "The link you clicked has expired")
		return
	} else if err == mysql.ErrRecordNotFound {
		session.SetFlash(r, session.SessionBad, "The requested link can't be found")
		return
	} else if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "Invalid code (1002)")
		return
	}

	user, err := mysql.GetUserByID(userID)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "User not found")
		return
	}

	// Enable user
	err = mysql.VerifyUser(userID)
	if err != nil {
		log.ErrS(err)
		session.SetFlash(r, session.SessionBad, "Invalid code (1003)")
		return
	}
	user.EmailVerified = true

	// Influx
	point := influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementSignups),
		Fields: map[string]interface{}{
			"validate": 1,
		},
		Time:      time.Now(),
		Precision: "s",
	}

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
	if err != nil {
		log.ErrS(err)
	}

	err = email.NewSignup(user.Email, r)
	if err != nil {
		log.ErrS(err)
	}

	login(r, user)

	//
	session.SetFlash(r, session.SessionGood, "Email has been verified")
}
