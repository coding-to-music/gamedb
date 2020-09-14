package email_providers

type EmailSender interface {
	Send(toName, toEmail, fromName, fromEmail, subject, html string) error
}

func GetSender() EmailSender {
	return mailjetProvider{}
}
