package email_providers

import (
	"errors"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"jaytaylor.com/html2text"
)

type sendgridProvider struct {
}

func (sendgridProvider) Send(toName, toEmail, fromName, fromEmail, subject, html string) (err error) {

	if config.C.SendGridAPIKey == "" {
		return errors.New("missing environment variables")
	}

	text, err := html2text.FromString(html, html2text.Options{PrettyTables: true})
	if err != nil {
		return err
	}

	_, err = sendgrid.NewSendClient(config.C.SendGridAPIKey).
		Send(mail.NewSingleEmail(mail.NewEmail(fromName, fromEmail), subject, mail.NewEmail(toName, toEmail), text, html))

	return err
}
