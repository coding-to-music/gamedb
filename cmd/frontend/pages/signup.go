package pages

import (
	"net/http"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/badoux/checkmail"
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
	t.fill(w, r, "Login", "Login to Game DB to set your currency and other things.")
	t.addAssetPasswordStrength()
	t.hideAds = true
	t.Domain = config.C.GameDBDomain
	t.RecaptchaPublic = config.C.RecaptchaPublic
	t.SignupEmail = session.Get(r, signupSessionEmail)
	t.Providers = oauth.Providers

	returnTemplate(w, r, "signup", t)
}

type signupTemplate struct {
	globalTemplate
	RecaptchaPublic string
	Domain          string
	SignupEmail     string
	Providers       []oauth.Provider
}

func (t signupTemplate) includes() []string {
	return []string{"includes/login_header.gohtml"}
}

func signupPostHandler(w http.ResponseWriter, r *http.Request) {

	message, success := func() (message string, success bool) {

		// Parse form
		err := r.ParseForm()
		if err != nil {
			log.ErrS(err)
			return "An error occurred (1001)", false
		}

		email := r.PostForm.Get("email")
		password := r.PostForm.Get("password")
		password2 := r.PostForm.Get("password2")

		// Remember email
		session.Set(r, signupSessionEmail, email)

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
		_, err = mysql.GetUserByEmail(email)
		if err == nil {
			return "An account with this email already exists", false
		}

		err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
		if err != nil {
			log.ErrS(err)
		}

		// Create user
		_, err = mysql.NewUser(r, email, password, session.GetProductCC(r), false)
		if err != nil {
			log.ErrS(err)
			return "An error occurred (1002)", false
		}

		return "Please check your email to verify your account (You might need to check the spam folder)", true
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

func verifyHandler(w http.ResponseWriter, r *http.Request) {

	message, success := func() (message string, success bool) {

		// Validate code
		code := r.URL.Query().Get("code")

		if len(code) != 10 {
			return "Invalid code (1001)", false
		}

		// Find email from code
		userID, err := mysql.GetUserVerification(code)
		if err == mysql.ErrExpiredVerification {
			return "Link Expired", false
		} else if err != nil {
			err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
			if err != nil {
				log.ErrS(err)
			}
			return "Invalid code (1002)", false
		}

		// Enable user
		err = mysql.VerifyUser(userID)
		if err != nil {
			log.ErrS(err)
			return "Invalid code (1003)", false
		}

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

		//
		return "Email has been verified", true
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
