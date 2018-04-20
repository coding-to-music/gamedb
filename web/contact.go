package web

import (
	"net/http"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/recaptcha"
)

func ContactHandler(w http.ResponseWriter, r *http.Request) {

	template := contactTemplate{}
	template.Fill(r, "Contact")
	template.RecaptchaPublic = os.Getenv("STEAM_RECAPTCHA_PUBLIC")

	returnTemplate(w, r, "contact", template)
}

func PostContactHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	template := contactTemplate{}
	template.Fill(r, "Contact")
	template.RecaptchaPublic = os.Getenv("STEAM_RECAPTCHA_PUBLIC")

	// Form validation
	if err := r.ParseForm(); err != nil {
		logger.Error(err)
		template.Messages = append(template.Messages, err.Error())
	}

	if r.PostForm.Get("name") == "" {
		template.Messages = append(template.Messages, "Please fill in your name.")
	}
	if r.PostForm.Get("email") == "" {
		template.Messages = append(template.Messages, "Please fill in your email.")
	}
	if r.PostForm.Get("message") == "" {
		template.Messages = append(template.Messages, "Please fill in a message.")
	}

	// Recaptcha
	var success bool
	var err error

	response := r.PostForm.Get("g-recaptcha-response")
	if response != "" {

		success, err = recaptcha.Check(os.Getenv("STEAM_RECAPTCHA_PRIVATE"), response, r.RemoteAddr)
		if err != nil {
			if err != recaptcha.ErrInvalidInputs {
				logger.Error(err)
			}
		}
	}

	if !success {
		template.Messages = append(template.Messages, "Please check the captcha.")
	}

	// Send
	if len(template.Messages) == 0 {

		message := mail.NewSingleEmail(
			mail.NewEmail(r.PostForm.Get("name"), r.PostForm.Get("email")),
			"Steam Authority Contact Form",
			mail.NewEmail(os.Getenv("STEAM_ADMIN_NAME"), os.Getenv("STEAM_ADMIN_EMAIL")),
			r.PostForm.Get("message"),
			r.PostForm.Get("message"),
		)
		client := sendgrid.NewSendClient(os.Getenv("STEAM_SENDGRID"))

		_, err := client.Send(message)
		if err != nil {
			template.Success = false
			template.Messages = append(template.Messages, "Something went wrong")
			logger.Error(err)
		} else {
			template.Success = true
			template.Messages = append(template.Messages, "Message sent.")
		}
	}

	returnTemplate(w, r, "contact", template)
}

type contactTemplate struct {
	GlobalTemplate
	RecaptchaPublic string
	Messages        []string
	Success         bool
}
