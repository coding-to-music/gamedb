package web

import (
	"errors"
	"net/http"

	"github.com/Jleagle/recaptcha-go"
	"github.com/gamedb/website/logging"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/spf13/viper"
)

func contactRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", ContactHandler)
	r.Post("/", PostContactHandler)
	return r
}

func ContactHandler(w http.ResponseWriter, r *http.Request) {

	t := contactTemplate{}
	t.Fill(w, r, "Contact")
	t.RecaptchaPublic = viper.GetString("RECAPTCHA_PUBLIC")
	t.Description = "Get in touch."

	err := returnTemplate(w, r, "contact", t)
	logging.Error(err)
}

type contactTemplate struct {
	GlobalTemplate
	RecaptchaPublic string
	Messages        []string
	Success         bool
}

func PostContactHandler(w http.ResponseWriter, r *http.Request) {

	err := func() (err error) {

		var ErrSomething = errors.New("something went wrong")

		// Parse form
		err = r.ParseForm()
		if err != nil {
			logging.Error(err)
			return err
		}

		// Backup
		err = session.WriteMany(w, r, map[string]string{
			"login-name":    r.PostForm.Get("name"),
			"login-email":   r.PostForm.Get("email"),
			"login-message": r.PostForm.Get("message"),
		})
		logging.Error(err)

		// Form validation
		if r.PostForm.Get("name") == "" {
			return errors.New("please fill in your name")
		}
		if r.PostForm.Get("email") == "" {
			return errors.New("please fill in your email")
		}
		if r.PostForm.Get("message") == "" {
			return errors.New("please fill in a message")
		}

		// Recaptcha
		err = recaptcha.CheckFromRequest(r)
		if err != nil {

			if err == recaptcha.ErrNotChecked {
				return errors.New("please check the captcha")
			}

			logging.Error(err)
			return ErrSomething
		}

		// Send
		message := mail.NewSingleEmail(
			mail.NewEmail(r.PostForm.Get("name"), r.PostForm.Get("email")),
			"Game DB Contact Form",
			mail.NewEmail(viper.GetString("ADMIN_NAME"), viper.GetString("ADMIN_EMAIL")),
			r.PostForm.Get("message"),
			r.PostForm.Get("message"),
		)
		client := sendgrid.NewSendClient(viper.GetString("SENDGRID"))

		_, err = client.Send(message)
		if err != nil {
			logging.Error(err)
			return ErrSomething
		}

		// Remove backup
		err = session.WriteMany(w, r, map[string]string{
			"login-name":    "",
			"login-email":   "",
			"login-message": "",
		})
		logging.Error(err)

		return nil
	}()

	// Redirect
	if err != nil {
		err = session.SetGoodFlash(w, r, err.Error())
		logging.Error(err)
		http.Redirect(w, r, "/contact", 302)
	} else {
		err = session.SetGoodFlash(w, r, "Message sent!")
		logging.Error(err)
		http.Redirect(w, r, "/contact", 302)
	}
}
