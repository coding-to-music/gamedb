package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/captcha"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/email"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/geo"
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
	t.fill(w, r, "contact", "Contact", "Get in touch with Global Steam")
	t.HCaptchaPublic = config.C.HCaptchaPublic

	t.SessionName = session.Get(r, contactSessionName)
	t.SessionEmail = session.Get(r, contactSessionEmail)
	t.SessionMessage = session.Get(r, contactSessionMessage)

	if t.SessionEmail == "" {
		t.SessionEmail = session.Get(r, session.SessionUserEmail)
	}

	returnTemplate(w, r, t)
}

type contactTemplate struct {
	globalTemplate
	HCaptchaPublic string
	Messages       []string
	Success        bool
	SessionName    string
	SessionEmail   string
	SessionMessage string
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

		// Captcha
		if !config.IsLocal() {

			resp, err := captcha.GetCaptcha().CheckRequest(r)
			if err != nil {
				log.ErrS(err)
				return session.SessionBad, "Something has gone wrong (1002)"
			}

			if !resp.Success {
				return session.SessionBad, "Captcha Failed"
			}
		}

		// Send
		if config.C.AdminName == "" || config.C.AdminEmail == "" {
			log.Err("Missing environment variables")
		} else {

			err = email.GetProvider().Send(
				config.C.AdminEmail,
				r.PostForm.Get("name"),
				r.PostForm.Get("email"),
				"Global Steam Contact Form",
				email.ContactTemplate{
					Message: r.PostForm.Get("message"),
					IP:      geo.GetFirstIP(r.RemoteAddr),
				},
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
