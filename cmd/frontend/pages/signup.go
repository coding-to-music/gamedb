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
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
	influx "github.com/influxdata/influxdb1-client"
	"golang.org/x/crypto/bcrypt"
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

	_, err := getUserFromSession(r)
	if err == nil {

		session.SetFlash(r, session.SessionGood, "Login successful")
		session.Save(w, r)

		http.Redirect(w, r, "/settings", http.StatusFound)
		return
	}

	t := signupTemplate{}
	t.fill(w, r, "Login", "Login to Game DB to set your currency and other things.")
	t.hideAds = true
	t.Domain = config.C.GameDBDomain
	t.RecaptchaPublic = config.C.RecaptchaPublic
	t.SignupEmail = session.Get(r, signupSessionEmail)

	returnTemplate(w, r, "signup", t)
}

type signupTemplate struct {
	globalTemplate
	RecaptchaPublic string
	Domain          string
	SignupEmail     string
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
			return "An error occurred", false
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
		_, err = mysql.GetUserByKey("email", email, 0)
		if err == nil {
			return "An account with this email already exists", false
		}

		err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
		if err != nil {
			log.ErrS(err)
		}

		// Create user
		db, err := mysql.GetMySQLClient()
		if err != nil {
			log.ErrS(err)
			return "An error occurred (1001)", false
		}

		passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
		if err != nil {
			log.ErrS(err)
			return "An error occurred (1002)", false
		}

		user := mysql.User{
			Email:         email,
			EmailVerified: false,
			Password:      string(passwordBytes),
			ProductCC:     session.GetProductCC(r),
			Level:         mysql.UserLevel1,
			LoggedInAt:    time.Unix(0, 0), // Fixes a gorm bug
		}

		user.SetAPIKey()

		db = db.Create(&user)
		if db.Error != nil {
			log.ErrS(db.Error)
			return "An error occurred (1003)", false
		}

		// Create verification code
		code, err := mysql.CreateUserVerification(user.ID)
		if err != nil {
			log.ErrS(err)
			return "An error occurred (1004)", false
		}

		// Send email
		body := "Please click the below link to verify your email address<br />" +
			config.C.GameDBDomain + "/signup/verify?code=" + code.Code +
			"<br><br>Thanks, Jleagle." +
			"<br><br>From IP: " + r.RemoteAddr

		err = email_providers.GetSender().Send(
			email,
			email,
			"",
			"",
			"Game DB Email Verification",
			body,
		)
		if err != nil {
			log.ErrS(err)
			return "An error occurred (1005)", false
		}

		// Create event
		err = mongo.CreateUserEvent(r, user.ID, mongo.EventSignup)
		if err != nil {
			log.ErrS(err)
		}

		// Influx
		point := influx.Point{
			Measurement: string(influxHelper.InfluxMeasurementSignups),
			Fields: map[string]interface{}{
				"signup": 1,
			},
			Time:      time.Now(),
			Precision: "s",
		}

		_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
		if err != nil {
			log.ErrS(err)
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
		err = mysql.UpdateUserCol(userID, "email_verified", true)
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
