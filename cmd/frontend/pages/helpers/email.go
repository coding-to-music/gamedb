package helpers

import (
	"errors"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"jaytaylor.com/html2text"
)

var sendGridClient *sendgrid.Client

// Lazy load client, fixes race condition with config
func getSendGridClient() *sendgrid.Client {

	if sendGridClient == nil {
		sendGridClient = sendgrid.NewSendClient(config.C.SendGridAPIKey)
	}

	return sendGridClient
}

func SendEmail(to, from *mail.Email, subject, html string) (resp *rest.Response, err error) {

	text, err := html2text.FromString(html, html2text.Options{PrettyTables: true})
	if err != nil {
		return resp, err
	}

	if config.C.SendGridAPIKey == "" {
		return nil, errors.New("missing environment variables")
	} else {
		return getSendGridClient().Send(mail.NewSingleEmail(from, subject, to, text, html))
	}
}
