package web

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Jleagle/go-helpers/logger"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func ContactHandler(w http.ResponseWriter, r *http.Request) {

	template := contactTemplate{}
	template.Fill(r, "Contact")
	template.RecaptchaPublic = os.Getenv("STEAM_RECAPTCHA_PUBLIC")

	returnTemplate(w, r, "contact", template)
}

type contactTemplate struct {
	GlobalTemplate
	RecaptchaPublic string
	Messages        []string
	Success         bool
}

func PostContactHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	template := new(contactTemplate)
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
	if r.PostForm.Get("g-recaptcha-response") != "" {

		form := url.Values{}
		form.Add("secret", os.Getenv("STEAM_RECAPTCHA_PRIVATE"))
		form.Add("response", r.PostForm.Get("g-recaptcha-response"))
		form.Add("remoteip", r.RemoteAddr)

		req, err := http.NewRequest("POST", "https://www.google.com/recaptcha/api/siteverify", bytes.NewBufferString(form.Encode()))
		if err != nil {
			logger.Error(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			logger.Error(err)
		}
		defer resp.Body.Close()

		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Error(err)
		}

		var str recaptchaResponse
		err = json.Unmarshal(respBytes, &str)
		if err != nil {
			logger.Error(err)
		}

		if !str.Success {
			template.Messages = append(template.Messages, "Please check the captcha.")
		}

	} else {
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

type recaptchaResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}
