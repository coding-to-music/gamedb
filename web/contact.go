package web

import (
	"errors"
	"net/http"
	"os"

	"github.com/Jleagle/recaptcha-go"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/session"
)

func ContactHandler(w http.ResponseWriter, r *http.Request) {

	t := contactTemplate{}
	t.Fill(w, r, "Contact")
	t.RecaptchaPublic = os.Getenv("STEAM_RECAPTCHA_PUBLIC")

	returnTemplate(w, r, "contact", t)
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
		if err := r.ParseForm(); err != nil {
			logger.Error(err)
			return err
		}

		// Backup
		session.WriteMany(w, r, map[string]string{
			"login-name":    r.PostForm.Get("name"),
			"login-email":   r.PostForm.Get("email"),
			"login-message": r.PostForm.Get("message"),
		})

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
			} else {
				logger.Error(err)
				return ErrSomething
			}
		}

		// Send
		message := mail.NewSingleEmail(
			mail.NewEmail(r.PostForm.Get("name"), r.PostForm.Get("email")),
			"Steam Authority Contact Form",
			mail.NewEmail(os.Getenv("STEAM_ADMIN_NAME"), os.Getenv("STEAM_ADMIN_EMAIL")),
			r.PostForm.Get("message"),
			r.PostForm.Get("message"),
		)
		client := sendgrid.NewSendClient(os.Getenv("STEAM_SENDGRID"))

		_, err = client.Send(message)
		if err != nil {
			logger.Error(err)
			return ErrSomething
		}

		// Remove backup
		session.WriteMany(w, r, map[string]string{
			"login-name":    "",
			"login-email":   "",
			"login-message": "",
		})

		return nil
	}()

	// Redirect
	if err != nil {
		session.SetGoodFlash(w, r, err.Error())
		http.Redirect(w, r, "/contact", 302)
	} else {
		session.SetGoodFlash(w, r, "Message sent!")
		http.Redirect(w, r, "/contact", 302)
	}

	return
}
