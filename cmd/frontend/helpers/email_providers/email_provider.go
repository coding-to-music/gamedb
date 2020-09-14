package email_providers

type EmailSender interface {
	Send(toName, toEmail, replyToName, replyToEmail, subject, html string) error
}

func GetSender() EmailSender {
	return mailjetProvider{}
}
