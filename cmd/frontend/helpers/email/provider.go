package email

import (
	"bytes"
	"errors"
	"html/template"
	"os"
)

type EmailProvider interface {
	Send(toEmail, replyToName, replyToEmail, subject string, template interface{}) error
}

func GetProvider() EmailProvider {
	return mailjetProvider{}
}

func getBodyFromTemplate(data interface{}) (body string, err error) {

	var t Template

	switch data.(type) {
	case ContactTemplate:
		t = TemplateContact
	case Forgot1Template:
		t = TemplateForgot1
	case Forgot2Template:
		t = TemplateForgot2
	case SignupTemplate:
		t = TemplateSignup
	case VerifyTemplate:
		t = TemplateVerify
	default:
		return "", errors.New("invalid email template")
	}

	ex, err := os.Getwd()
	if err != nil {
		return "", err
	}

	base := ex + "/helpers/email/templates/"

	tmpl, err := template.ParseFiles(
		base+"_header.gohtml",
		base+"_footer.gohtml",
		base+string(t)+".gohtml")
	if err != nil {
		return "", err
	}

	buf := bytes.Buffer{}

	err = tmpl.ExecuteTemplate(&buf, string(t), data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
