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

func EmailSenderActivity(ctx context.Context, emailrequest *helpers.EmailRequest) (string, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("****EMAIL*** => ")

	_, err := emailrequest.SendEmail()

	if err != nil {
		logger.Info(err.Error())
		return "", errors.New(err.Error())
	}

	return "Сообщение отправлено на е-маил", nil
}
