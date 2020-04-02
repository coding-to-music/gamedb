package pages

import (
	"errors"
	"net/http"

	"github.com/Jleagle/recaptcha-go"
	"github.com/Jleagle/session-go/session"
	webserverHelpers "github.com/gamedb/gamedb/cmd/webserver/helpers"
	sessionHelpers "github.com/gamedb/gamedb/cmd/webserver/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

const (
	contactSessionName    = "contact-name"
	contactSessionEmail   = "contact-email"
	contactSessionMessage = "contact-message"
)

func ContactRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", contactHandler)
	r.Post("/", postContactHandler)
	return r
}

func contactHandler(w http.ResponseWriter, r *http.Request) {

	t := contactTemplate{}
	t.fill(w, r, "Contact", "Get in touch with Game DB.")
	t.RecaptchaPublic = config.Config.RecaptchaPublic.Get()

	t.SessionName = sessionHelpers.Get(r, contactSessionName)
	t.SessionEmail = sessionHelpers.Get(r, contactSessionEmail)
	t.SessionMessage = sessionHelpers.Get(r, contactSessionMessage)

	if t.SessionEmail == "" {
		t.SessionEmail = sessionHelpers.Get(r, sessionHelpers.SessionUserEmail)
	}

	returnTemplate(w, r, "contact", t)
}

type contactTemplate struct {
	GlobalTemplate
	RecaptchaPublic string
	Messages        []string
	Success         bool
	SessionName     string
	SessionEmail    string
	SessionMessage  string
}

func postContactHandler(w http.ResponseWriter, r *http.Request) {

	err := func() (err error) {

		var ErrSomething = errors.New("something went wrong")

		// Parse form
		err = r.ParseForm()
		if err != nil {
			log.Err(err, r)
			return err
		}

		// Backup
		err = session.SetMany(r, map[string]string{
			contactSessionName:    r.PostForm.Get("name"),
			contactSessionEmail:   r.PostForm.Get("email"),
			contactSessionMessage: r.PostForm.Get("message"),
		})
		log.Err(err, r)

		// Form validation
		if r.PostForm.Get("name") == "" {
			return errors.New("Please fill in your name")
		}
		if r.PostForm.Get("email") == "" {
			return errors.New("Please fill in your email")
		}
		if r.PostForm.Get("message") == "" {
			return errors.New("Please fill in a message")
		}

		// Recaptcha
		if config.IsProd() {
			err = recaptcha.CheckFromRequest(r)
			if err != nil {

				if err == recaptcha.ErrNotChecked {
					return errors.New("please check the captcha")
				}

				log.Err(err, r)
				return ErrSomething
			}
		}

		// Send
		_, err = webserverHelpers.SendEmail(
			mail.NewEmail(config.Config.AdminName.Get(), config.Config.AdminEmail.Get()),
			mail.NewEmail(r.PostForm.Get("name"), r.PostForm.Get("email")),
			"Game DB Contact Form",
			r.PostForm.Get("message"),
		)
		if err != nil {
			log.Err(err, r)
			return ErrSomething
		}

		// Remove backup
		err = session.SetMany(r, map[string]string{
			contactSessionName:    "",
			contactSessionEmail:   "",
			contactSessionMessage: "",
		})
		log.Err(err, r)

		return nil
	}()

	// Redirect
	if err != nil {
		err = session.SetFlash(r, sessionHelpers.SessionBad, err.Error())
	} else {
		err = session.SetFlash(r, sessionHelpers.SessionGood, "Message sent!")
	}
	log.Err(err)

	sessionHelpers.Save(w, r)

	log.Err(err, r)
	http.Redirect(w, r, "/contact", http.StatusFound)
}
