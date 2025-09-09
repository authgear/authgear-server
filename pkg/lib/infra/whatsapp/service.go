package whatsapp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var logger = slogutil.NewLogger("whatsapp-service")

type Service struct {
	Clock                 clock.Clock
	WhatsappConfig        *config.WhatsappConfig
	LocalizationConfig    *config.LocalizationConfig
	GlobalWhatsappAPIType config.GlobalWhatsappAPIType
	OnPremisesClient      *OnPremisesClient
	CloudAPIClient        *CloudAPIClient
	MessageStore          *MessageStore
}

func (s *Service) resolveTemplateLanguage(ctx context.Context, supportedLanguages []string) string {
	if len(supportedLanguages) < 1 {
		panic("whatsapp: template has no supported language")
	}
	preferredLanguageTags := intl.GetPreferredLanguageTags(ctx)
	configSupportedLanguageTags := intl.Supported(
		s.LocalizationConfig.SupportedLanguages,
		intl.Fallback(*s.LocalizationConfig.FallbackLanguage),
	)
	// First, resolve once based on supported language in config
	// This is to avoid inconsistency of ui language and whatsapp message language
	_, resolvedTag := intl.BestMatch(preferredLanguageTags, configSupportedLanguageTags)
	supportedLanguageTags := intl.Supported(supportedLanguages, intl.Fallback(supportedLanguages[0]))

	// Then, resolve to a language supported by the whatsapp template
	idx, _ := intl.BestMatch([]string{resolvedTag.String()}, supportedLanguageTags)
	return supportedLanguageTags[idx]
}

func (s *Service) makeAuthenticationTemplateComponents(code string) []onPremisesTemplateComponent {
	// See https://developers.facebook.com/docs/whatsapp/api/messages/message-templates/authentication-message-templates

	var component []onPremisesTemplateComponent = []onPremisesTemplateComponent{}

	body := onPremisesNewTemplateComponent(onPremisesTemplateComponentTypeBody)
	// The body is just the code.
	bodyParam := onPremisesNewTemplateComponentTextParameter(code)
	body.Parameters = append(body.Parameters, *bodyParam)
	component = append(component, *body)

	button := onPremisesNewTemplateButtonComponent(onPremisesTemplateComponentSubTypeURL, 0)
	// The button copies the code.
	buttonParam := onPremisesNewTemplateComponentTextParameter(code)
	button.Parameters = append(button.Parameters, *buttonParam)
	component = append(component, *button)

	return component
}

func (s *Service) prepareOTPComponents(template *config.WhatsappOnPremisesOTPTemplateConfig, code string) []onPremisesTemplateComponent {
	switch template.Type {
	case config.WhatsappOnPremisesTemplateTypeAuthentication:
		return s.makeAuthenticationTemplateComponents(code)
	default:
		panic("whatsapp: unknown template type")
	}
}

func (s *Service) SendAuthenticationOTP(ctx context.Context, opts *SendAuthenticationOTPOptions) error {
	switch s.WhatsappConfig.GetAPIType(s.GlobalWhatsappAPIType) {
	case config.WhatsappAPITypeOnPremises:
		if s.OnPremisesClient == nil {
			return ErrNoAvailableWhatsappClient
		}

		otpTemplate := s.OnPremisesClient.GetOTPTemplate()
		lang := s.resolveTemplateLanguage(ctx, otpTemplate.Languages)
		components := s.prepareOTPComponents(otpTemplate, opts.OTP)

		return s.OnPremisesClient.SendTemplate(
			ctx,
			opts.To,
			otpTemplate,
			lang,
			components)
	case config.WhatsappAPITypeCloudAPI:
		if s.CloudAPIClient == nil {
			return ErrNoAvailableWhatsappClient
		}

		configuredLanguages := s.CloudAPIClient.GetLanguages()
		lang := s.resolveTemplateLanguage(ctx, configuredLanguages)
		messageID, err := s.CloudAPIClient.SendAuthenticationOTP(
			ctx,
			opts,
			lang,
		)
		if err != nil {
			return err
		}
		success := make(chan bool, 1)
		// Wait for 5 seconds for the message status
		go s.waitUntilSent(ctx, success, messageID, 5*time.Second)

		isSuccess := <-success
		if !isSuccess {
			// Historically, InvalidWhatsappUser means message failed to deliver
			// Therefore, we return this error when we failed to get a sent status within 5 seconds
			return ErrInvalidWhatsappUser
		}
		return nil

	default:
		panic(fmt.Errorf("whatsapp: unknown api type"))
	}
}

func (s *Service) waitUntilSent(ctx context.Context, success chan bool, messageID string, timeout time.Duration) {
	logger := logger.GetLogger(ctx)
	start := s.Clock.NowUTC()
	for {
		time.Sleep(500 * time.Millisecond)
		logger.Info(ctx, "waiting for message status update...", slog.String("message_id", messageID))
		timeElasped := s.Clock.NowUTC().Sub(start)
		if timeElasped > timeout {
			logger.Error(ctx, "failed to wait for whatsapp message status: timeout")
			success <- false
			return
		}
		status, err := s.MessageStore.GetMessageStatus(ctx, messageID)
		if err != nil {
			logger.WithError(err).Error(ctx, "failed to get message status")
			success <- false
			return
		}
		switch status {
		case WhatsappMessageStatusFailed:
			success <- false
			return
		case WhatsappMessageStatusDelivered, WhatsappMessageStatusRead, WhatsappMessageStatusSent:
			success <- true
			return
		case WhatsappMessageStatusAccepted, "":
			// Unknown yet
			continue
		default:
			// Unknown status
			success <- false
			logger.WithError(err).With(
				slog.String("status", string(status)),
			).Error(ctx, "unexpected whatsapp message status")
			return
		}
	}
}

func (s *Service) UpdateMessageStatus(ctx context.Context, messageID string, status WhatsappMessageStatus) error {
	return s.MessageStore.UpdateMessageStatus(ctx, messageID, status)
}
