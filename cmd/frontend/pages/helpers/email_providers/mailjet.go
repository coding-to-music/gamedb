package email_providers

import (
	"errors"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/mailjet/mailjet-apiv3-go"
	"jaytaylor.com/html2text"
)

type mailjetProvider struct {
}

func (mailjetProvider) Send(toName, toEmail, fromName, fromEmail, subject, html string) (err error) {

	if config.C.MailjetPublic == "" || config.C.MailjetPrivate == "" {
		return errors.New("missing environment variables")
	}

	text, err := html2text.FromString(html, html2text.Options{PrettyTables: true})
	if err != nil {
		return err
	}

	messages := []mailjet.InfoMessagesV31{
		{
			To:       &mailjet.RecipientsV31{mailjet.RecipientV31{Name: toName, Email: toEmail}},
			From:     &mailjet.RecipientV31{Name: "Game DB", Email: "no-reply@gamedb.online"},
			ReplyTo:  &mailjet.RecipientV31{Name: fromName, Email: fromEmail},
			Subject:  subject,
			HTMLPart: html,
			TextPart: text,
			CustomID: "",
		},
	}

	client := mailjet.NewMailjetClient(config.C.MailjetPublic, config.C.MailjetPrivate)
	_, err = client.SendMailV31(&mailjet.MessagesV31{Info: messages})
	return err
}
