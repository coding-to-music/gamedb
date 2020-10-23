package email

import (
	"errors"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"jaytaylor.com/html2text"
)

type sendgridProvider struct {
}

func (sendgridProvider) Send(toEmail, replyToName, replyToEmail, subject string, template interface{}) (err error) {

	if config.C.SendGridAPIKey == "" {
		return errors.New("missing environment variables")
	}

	html, err := getBodyFromTemplate(template)
	if err != nil {
		return err
	}

	text, err := html2text.FromString(html, html2text.Options{PrettyTables: true})
	if err != nil {
		return err
	}

	_, err = sendgrid.NewSendClient(config.C.SendGridAPIKey).
		Send(mail.NewSingleEmail(mail.NewEmail(replyToName, replyToEmail), subject, mail.NewEmail("", toEmail), text, html))

	return err
}
