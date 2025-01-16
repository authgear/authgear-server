package sms

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/twilio"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("portal-sms")}
}

type Service struct {
	Logger Logger
}

func (s *Service) SendByTwilio(ctx context.Context,
	app *model.App,
	to string,
	cfg model.SMSProviderConfigurationTwilioInput,
) error {

	twilioClient := twilio.NewTwilioClient(&config.TwilioCredentials{
		AccountSID:          cfg.AccountSID,
		AuthToken:           cfg.AuthToken,
		MessagingServiceSID: cfg.MessagingServiceSID,
	})

	translationService := NewTranslationService(app)
	sender, err := translationService.GetSenderForTestSMS(ctx)
	if err != nil {
		return err
	}

	return twilioClient.Send(ctx, smsapi.SendOptions{
		Sender: sender,
		To:     to,
		Body:   "[Test] Authgear sms",
	})
}
