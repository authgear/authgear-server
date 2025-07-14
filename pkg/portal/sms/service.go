package sms

import (
	"context"
	"fmt"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
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
	DenoEndpoint config.DenoEndpoint
	Logger       Logger
}

const TEST_OTP = "000000"
const TEST_APP_NAME = "Test"

func makeTestSMSBody(appName string, code string) string {
	return fmt.Sprintf("[%s] Your one-time password is %s", appName, code)
}

func (s *Service) sendByTwilio(
	ctx context.Context,
	app *model.App,
	to string,
	cfg model.SMSProviderConfigurationTwilioInput,
) error {
	twilioClient := twilio.NewTwilioClient(&config.TwilioCredentials{
		CredentialType_WriteOnly: &cfg.CredentialType,
		AccountSID:               cfg.AccountSID,
		AuthToken:                cfg.AuthToken,
		APIKeySID:                cfg.APIKeySID,
		APIKeySecret:             cfg.APIKeySecret,
		MessagingServiceSID:      cfg.MessagingServiceSID,
		From:                     cfg.From,
	})

	translationService := NewTranslationService(app)
	sender, err := translationService.GetSenderForTestSMS(ctx)
	if err != nil {
		return err
	}

	return twilioClient.Send(ctx, smsapi.SendOptions{
		Sender: sender,
		To:     to,
		Body:   makeTestSMSBody(TEST_APP_NAME, TEST_OTP),
		TemplateVariables: &smsapi.TemplateVariables{
			AppName: TEST_APP_NAME,
			Code:    TEST_OTP,
		},
	})
}

func (s *Service) sendByWebhook(
	ctx context.Context,
	secret *config.WebhookKeyMaterials,
	to string,
	cfg model.SMSProviderConfigurationWebhookInput,
) error {
	webHookImpl := &hook.WebHookImpl{
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

	err = webhook.Call(ctx, url, custom.SendOptions{
		To:   to,
		Body: makeTestSMSBody(TEST_APP_NAME, TEST_OTP),
		TemplateVariables: &smsapi.TemplateVariables{
			AppName: TEST_APP_NAME,
			Code:    TEST_OTP,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) sendByDeno(
	ctx context.Context,
	app *model.App,
	to string,
	cfg model.SMSProviderConfigurationDenoInput,
) error {

	deno := custom.NewSMSDenoHookForTest(s.DenoEndpoint, &config.CustomSMSProviderConfig{
		// URL is not important here, we execute the script with a string
		URL:     "",
		Timeout: (*config.DurationSeconds)(cfg.Timeout),
	})

	err := deno.Test(ctx, cfg.Script, custom.SendOptions{
		To:   to,
		Body: makeTestSMSBody(TEST_APP_NAME, TEST_OTP),
		TemplateVariables: &smsapi.TemplateVariables{
			AppName: TEST_APP_NAME,
			Code:    TEST_OTP,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) SendTestSMS(
	ctx context.Context,
	app *model.App,
	to string,
	webhookSecretLoader func(ctx context.Context) (*config.WebhookKeyMaterials, error),
	input model.SMSProviderConfigurationInput) error {
	if input.Twilio != nil {
		return s.sendByTwilio(ctx, app, to, *input.Twilio)

	} else if input.Webhook != nil {
		webhookSecret, err := webhookSecretLoader(ctx)
		if err != nil {
			return err
		}
		return s.sendByWebhook(ctx, webhookSecret, to, *input.Webhook)

	} else if input.Deno != nil {
		return s.sendByDeno(ctx, app, to, *input.Deno)
	}
	return apierrors.NewInvalid("no provider config given")
}
