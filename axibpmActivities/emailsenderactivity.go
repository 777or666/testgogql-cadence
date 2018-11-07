package axibpmActivities

import (
	"context"

	"errors"
	"net/smtp"

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

func EmailSenderActivity(ctx context.Context, addressees []string) (string, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("axibpmActivities: EmailSenderActivity начинаю отправку е-маил")

	//	auth := smtp.PlainAuth(
	//		"",
	//		"belka@axitech.ru",
	//		"778523",
	//		"mail.axitech.ru",
	//	)

	auth := unencryptedAuth{
		smtp.PlainAuth(
			"",
			"m.kravetz@axitech.ru",
			"!1",
			"mail.axitech.ru",
		),
	}

	err := smtp.SendMail(
		"mail.axitech.ru:25",
		auth,
		"m.kravetz@axitech.ru",
		[]string{"kravetsmihail@mail.ru"},
		[]byte("AXI-BPM. ТЕСТВОЕ ПИСЬМО"),
	)
	if err != nil {
		logger.Info(err.Error())
		return "", errors.New(err.Error())
	}

	return "Сообщение отправлено на е-маил", nil
}
