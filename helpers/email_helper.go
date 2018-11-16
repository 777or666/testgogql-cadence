package helpers

import (
	"bytes"
	"html/template"
	"net/smtp"
)

type EmailRequest struct {
	To      []string
	Subject string
	Body    string
	Config  *EmailConfig
}

type EmailRequestData struct {
	Message      string
	WorkflowData WorkflowInput
}

func (r *EmailRequest) ParseEmailTemplate(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}
	r.Body = buf.String()
	return nil
}

func (r *EmailRequest) SendEmail() (bool, error) {

	// Google
	//	auth := smtp.PlainAuth(
	//			r.Config.Emailidentity,
	//			r.Config.Emailusername,
	//			r.Config.Emailpassword,
	//			r.Config.Emailhost)

	// АКСИТЕХ
	auth := unencryptedAuth{
		smtp.PlainAuth(
			r.Config.Emailidentity,
			r.Config.Emailusername,
			r.Config.Emailpassword,
			r.Config.Emailhost),
	}

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: " + r.Subject + "\n"
	msg := []byte(subject + mime + "\n" + r.Body)

	err := smtp.SendMail(
		r.Config.Emailhost+":"+r.Config.Emailport,
		auth,
		r.Config.Emailfrom,
		r.To,
		msg,
	)

	if err != nil {
		return false, err
	}
	return true, nil
}

type unencryptedAuth struct {
	smtp.Auth
}

func (a unencryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	s := *server
	s.TLS = true
	return a.Auth.Start(&s)
}
