package pages

import (
	"net/http"

	"github.com/Jleagle/recaptcha-go"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/email_providers"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
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
	t.RecaptchaPublic = config.C.RecaptchaPublic

	t.SessionName = session.Get(r, contactSessionName)
	t.SessionEmail = session.Get(r, contactSessionEmail)
	t.SessionMessage = session.Get(r, contactSessionMessage)

	if t.SessionEmail == "" {
		t.SessionEmail = session.Get(r, session.SessionUserEmail)
	}

	returnTemplate(w, r, "contact", t)
}

type contactTemplate struct {
	globalTemplate
	RecaptchaPublic string
	Messages        []string
	Success         bool
	SessionName     string
	SessionEmail    string
	SessionMessage  string
}

func postContactHandler(w http.ResponseWriter, r *http.Request) {

	flashGroup, message := func() (session.FlashGroup, string) {

		// Parse form
		err := r.ParseForm()
		if err != nil {
			log.ErrS(err)
			return session.SessionBad, "Something has gone wrong (1001)"
		}

		// Backup
		session.SetMany(r, map[string]string{
			contactSessionName:    r.PostForm.Get("name"),
			contactSessionEmail:   r.PostForm.Get("email"),
			contactSessionMessage: r.PostForm.Get("message"),
		})

		// Form validation
		if r.PostForm.Get("name") == "" {
			return session.SessionBad, "Please fill in your name"
		}
		if r.PostForm.Get("email") == "" {
			return session.SessionBad, "Please fill in your email"
		}
		if r.PostForm.Get("message") == "" {
			return session.SessionBad, "Please fill in a message"
		}

		// Recaptcha
		if !config.IsLocal() {

			err = recaptcha.CheckFromRequest(r)
			if err != nil {

				if err == recaptcha.ErrNotChecked {
					return session.SessionBad, "Please check the captcha"
				}

				log.ErrS(err)
				return session.SessionBad, "Something has gone wrong (1002)"
			}
		}

		// Send
		if config.C.AdminName == "" || config.C.AdminEmail == "" {
			log.Fatal("Missing environment variables")
		} else {

			err = email_providers.GetSender().Send(
				config.C.AdminName,
				config.C.AdminEmail,
				r.PostForm.Get("name"),
				r.PostForm.Get("email"),
				"Game DB Contact Form",
				r.PostForm.Get("message"),
			)

			if err != nil {
				log.ErrS(err)
				return session.SessionBad, "Something has gone wrong (1003)"
			}
		}

		// Remove backup
		session.SetMany(r, map[string]string{
			contactSessionName:    "",
			contactSessionEmail:   "",
			contactSessionMessage: "",
		})

		return session.SessionGood, "Message sent!"
	}()

	// Redirect
	session.SetFlash(r, flashGroup, message)
	session.Save(w, r)

	http.Redirect(w, r, "/contact", http.StatusFound)
}
