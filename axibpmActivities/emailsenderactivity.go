package axibpmActivities

import (
	"context"

	"errors"
	"net/smtp"

	"github.com/777or666/testgogql-cadence/helpers"
	"go.uber.org/cadence/activity"
)

type unencryptedAuth struct {
	smtp.Auth
}

func (a unencryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	s := *server
	s.TLS = true
	return a.Auth.Start(&s)
}

//func EmailSenderActivity(ctx context.Context, addressees []string, emailbody string, emailsubject string, emailconfig *helpers.EmailConfig) (string, error) {
func EmailSenderActivity(ctx context.Context, emailrequest *helpers.EmailRequest) (string, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("axibpmActivities: EmailSenderActivity начинаю отправку")

	//	// Google
	//	//	auth := smtp.PlainAuth(
	//	//			emailconfig.Emailidentity,
	//	//			emailconfig.Emailusername,
	//	//			emailconfig.Emailpassword,
	//	//			emailconfig.Emailhost)

	//	// АКСИТЕХ
	//	auth := unencryptedAuth{
	//		smtp.PlainAuth(
	//			emailrequest.Conf.Emailidentity,
	//			emailrequest.Conf.Emailusername,
	//			emailrequest.Conf.Emailpassword,
	//			emailrequest.Conf.Emailhost),
	//	}

	//	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	//	subject := "Subject: " +
	//		//emailrequest.subject +
	//		"TEST TEST" +
	//		"\n"
	//	msg := []byte(subject + mime + "\n" +
	//		//emailrequest.body
	//		"TEST TEST")

	//	err := smtp.SendMail(
	//		emailrequest.Conf.Emailhost+":"+emailrequest.Conf.Emailport,
	//		auth,
	//		emailrequest.Conf.Emailfrom,
	//		emailrequest.To,
	//		msg,
	//	)
	//	if err != nil {
	//		logger.Info(err.Error())
	//		return "", errors.New(err.Error())
	//	}

	logger.Info("****EMAIL*** => ")

	_, err := emailrequest.SendEmail()

	if err != nil {
		logger.Info(err.Error())
		return "", errors.New(err.Error())
	}

	return "Сообщение отправлено на е-маил", nil
}
