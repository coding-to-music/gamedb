package email

import (
	"bytes"
	"errors"
	"html/template"

	"github.com/gamedb/gamedb/pkg/log"
)

type EmailProvider interface {
	Send(toEmail, replyToName, replyToEmail, subject string, template interface{}) error
}

func GetProvider() EmailProvider {
	return mailjetProvider{}
}

var templatex *template.Template

func Init() {

	var err error
	templatex, err = template.ParseGlob("./helpers/email/templates/*.gohtml")
	if err != nil {
		log.ErrS(err)
		return
	}
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

	buf := bytes.Buffer{}
	err = templatex.ExecuteTemplate(&buf, string(t), data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
