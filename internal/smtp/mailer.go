package smtp

import (
	"bytes"
	"embed"
	"github.com/go-mail/mail/v2"
	"html/template"
	"time"
)

var templateFS embed.FS

type Mailer struct {
	dialer *mail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return Mailer{
		sender: sender,
		dialer: dialer,
	}
}

func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
	// use ParseFS() method to parse required template from the embedded file system
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// Execute named template "subject", passing the dynamic data and storing
	// the result in buffer
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// Follow the same pattern to execute the "plainBody" template and store the result
	// in the plainBody variable
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}

	return nil
}
