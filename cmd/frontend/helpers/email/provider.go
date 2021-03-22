package email

import (
	"bytes"
	"html/template"

	"github.com/gamedb/gamedb/pkg/log"
)

type EmailProvider interface {
	Send(toEmail, replyToName, replyToEmail, subject string, template emailTemplate) error
}

func GetProvider() EmailProvider {
	return mailjetProvider{}
}

var templatex *template.Template

func Init() {

	var err error
	templatex, err = template.ParseGlob("./templates/emails/*.gohtml")
	if err != nil {
		log.ErrS(err)
		return
	}
}

func renderTemplate(template emailTemplate) (body string, err error) {

	buf := bytes.Buffer{}
	err = templatex.ExecuteTemplate(&buf, template.filename(), template)
	return buf.String(), err
}
