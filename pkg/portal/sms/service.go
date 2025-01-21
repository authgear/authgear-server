package sms

import (
	"context"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
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
	LoggerFactory *log.Factory
	DenoEndpoint  config.DenoEndpoint
	Logger        Logger
}

const SMS_BODY = "[Test] Authgear sms"

func (s *Service) SendByTwilio(
	ctx context.Context,
	app *model.App,
	to string,
	cfg model.SMSProviderConfigurationTwilioInput,
) error {
	messagingServiceSID := ""
	if cfg.MessagingServiceSID != nil {
		messagingServiceSID = *cfg.MessagingServiceSID
	}
	twilioClient := twilio.NewTwilioClient(&config.TwilioCredentials{
		AccountSID:          cfg.AccountSID,
		AuthToken:           cfg.AuthToken,
		MessagingServiceSID: messagingServiceSID,
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
	secret *config.WebhookKeyMaterials,
	to string,
	cfg model.SMSProviderConfigurationWebhookInput,
) error {
	webHookImpl := &hook.WebHookImpl{
		Logger: hook.NewWebHookLogger(s.LoggerFactory),
		Secret: secret,
	}
	webhook := custom.NewSMSWebHook(webHookImpl, &config.CustomSMSProviderConfig{
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

func (s *Service) SendByDeno(
	ctx context.Context,
	app *model.App,
	to string,
	cfg model.SMSProviderConfigurationDenoInput,
) error {

	deno := custom.NewSMSDenoHook(s.LoggerFactory, s.DenoEndpoint, &config.CustomSMSProviderConfig{
		// URL is not important here, we execute the script with a string
		URL:     "",
		Timeout: (*config.DurationSeconds)(cfg.Timeout),
	})

	err := deno.Test(ctx, cfg.Script, custom.SendOptions{
		To:   to,
		Body: SMS_BODY,
	})
	if err != nil {
		return err
	}
	return nil
}
