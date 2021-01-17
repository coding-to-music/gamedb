package email

import (
	"errors"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/mailjet/mailjet-apiv3-go"
	"jaytaylor.com/html2text"
)

type mailjetProvider struct {
}

func (mailjetProvider) Send(toEmail, replyToName, replyToEmail, subject string, template emailTemplate) (err error) {

	if config.C.MailjetPublic == "" || config.C.MailjetPrivate == "" {
		return errors.New("missing mailjet environment variables")
	}

	html, err := renderTemplate(template)
	if err != nil {
		return err
	}

	text, err := html2text.FromString(html, html2text.Options{PrettyTables: true})
	if err != nil {
		return err
	}

	message := mailjet.InfoMessagesV31{
		To:       &mailjet.RecipientsV31{mailjet.RecipientV31{Email: toEmail}},
		From:     &mailjet.RecipientV31{Name: "Global Steam", Email: "no-reply@globalsteam.online"}, // Must be from verified domain
		Subject:  subject,
		HTMLPart: html,
		TextPart: text,
		CustomID: "",
	}

	if replyToName != "" && replyToEmail != "" {
		message.ReplyTo = &mailjet.RecipientV31{Name: replyToName, Email: replyToEmail}
	}

	client := mailjet.NewMailjetClient(config.C.MailjetPublic, config.C.MailjetPrivate)
	_, err = client.SendMailV31(&mailjet.MessagesV31{Info: []mailjet.InfoMessagesV31{message}})
	return err
}
