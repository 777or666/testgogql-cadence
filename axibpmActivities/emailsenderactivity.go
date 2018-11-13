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

func EmailSenderActivity(ctx context.Context, addressees []string, emailbody string, emailconfig *helpers.EmailConfig) (string, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("axibpmActivities: EmailSenderActivity начинаю отправку от имени")

	// Google
	//	auth := smtp.PlainAuth(
	//			emailconfig.Emailidentity,
	//			emailconfig.Emailusername,
	//			emailconfig.Emailpassword,
	//			emailconfig.Emailhost)

	// АКСИТЕХ
	auth := unencryptedAuth{
		smtp.PlainAuth(
			emailconfig.Emailidentity,
			emailconfig.Emailusername,
			emailconfig.Emailpassword,
			emailconfig.Emailhost),
	}

	err := smtp.SendMail(
		emailconfig.Emailhost+":"+emailconfig.Emailport,
		auth,
		emailconfig.Emailfrom,
		addressees,
		[]byte(emailbody),
	)
	if err != nil {
		logger.Info(err.Error())
		return "", errors.New(err.Error())
	}

	return "Сообщение отправлено на е-маил", nil
}
