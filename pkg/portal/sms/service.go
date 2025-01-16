package sms

import (
	"context"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/custom"
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

const SMS_BODY = "[Test] Authgear sms"

func (s *Service) SendByTwilio(
	ctx context.Context,
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
		Body:   SMS_BODY,
	})
}

func (s *Service) SendByWebhook(
	ctx context.Context,
	app *model.App,
	to string,
	cfg model.SMSProviderConfigurationWebhookInput,
) error {

	webhook := custom.NewSMSWebHook(&config.CustomSMSProviderConfig{
		URL:     cfg.URL,
		Timeout: (*config.DurationSeconds)(cfg.Timeout),
	})

	url, err := url.Parse(cfg.URL)
	if err != nil {
		return err
	}

	_, err = webhook.Call(ctx, url, custom.SendOptions{
		To:   to,
		Body: SMS_BODY,
	})
	if err != nil {
		return err
	}
	return nil
}
