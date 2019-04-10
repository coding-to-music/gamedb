package web

import (
	"errors"
	"net/http"

	"github.com/Jleagle/recaptcha-go"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func contactRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", contactHandler)
	r.Post("/", postContactHandler)
	return r
}

func contactHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	t := contactTemplate{}
	t.fill(w, r, "Contact", "Get in touch with Game DB.")
	t.RecaptchaPublic = config.Config.RecaptchaPublic

	err := returnTemplate(w, r, "contact", t)
	log.Err(err, r)
}

type contactTemplate struct {
	GlobalTemplate
	RecaptchaPublic string
	Messages        []string
	Success         bool
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
		err = session.WriteMany(w, r, map[string]string{
			"contact-name":    r.PostForm.Get("name"),
			"contact-email":   r.PostForm.Get("email"),
			"contact-message": r.PostForm.Get("message"),
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
		err = recaptcha.CheckFromRequest(r)
		if err != nil {

			if err == recaptcha.ErrNotChecked {
				return errors.New("please check the captcha")
			}

			log.Err(err, r)
			return ErrSomething
		}

		// Send
		message := mail.NewSingleEmail(
			mail.NewEmail(r.PostForm.Get("name"), r.PostForm.Get("email")),
			"Game DB Contact Form",
			mail.NewEmail(config.Config.AdminName, config.Config.AdminEmail),
			r.PostForm.Get("message"),
			r.PostForm.Get("message"),
		)
		client := sendgrid.NewSendClient(config.Config.SendGridAPIKey)

		_, err = client.Send(message)
		if err != nil {
			log.Err(err, r)
			return ErrSomething
		}

		// Remove backup
		err = session.WriteMany(w, r, map[string]string{
			"contact-name":    "",
			"contact-email":   "",
			"contact-message": "",
		})
		log.Err(err, r)

		return nil
	}()

	// Redirect
	if err != nil {
		err = session.SetGoodFlash(w, r, err.Error())
		log.Err(err, r)
		http.Redirect(w, r, "/contact", http.StatusTemporaryRedirect)
	} else {
		err = session.SetGoodFlash(w, r, "Message sent!")
		log.Err(err, r)
		http.Redirect(w, r, "/contact", http.StatusTemporaryRedirect)
	}
}
