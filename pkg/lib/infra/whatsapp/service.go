package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/otelutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var logger = slogutil.NewLogger("whatsapp-service")

//go:generate go tool mockgen -source=service.go -destination=service_mock_test.go -package whatsapp_test

type ServiceCloudAPIClient interface {
	GetLanguages() []string
	SendAuthenticationOTP(ctx context.Context, opts *SendAuthenticationOTPOptions, lang string) (result *CloudAPISendAuthenticationOTPResult, err error)
}

type ServiceMessageStore interface {
	GetMessageStatus(ctx context.Context, messageID string) (*WhatsappMessageStatusData, error)
	UpdateMessageStatus(ctx context.Context, messageID string, status *WhatsappMessageStatusData) error
	SetMessageStatusIfNotExist(ctx context.Context, messageID string, status *WhatsappMessageStatusData) (bool, error)
}

type SendAuthenticationOTPResult struct {
	MessageID     string
	MessageStatus WhatsappMessageStatus
}

type GetMessageStatusResult struct {
	Status   WhatsappMessageStatus
	APIError *apierrors.APIError
}

const (
	// A special message id to identify suppressed message
	suppressedMessageID string = "_suppressed-message-id"
	// A special message id to identify unknown message id
	unknownMessageID string = "_unknown-message-id"
)

type Service struct {
	Clock                 clock.Clock
	WhatsappConfig        *config.WhatsappConfig
	LocalizationConfig    *config.LocalizationConfig
	GlobalWhatsappAPIType config.GlobalWhatsappAPIType
	OnPremisesClient      *OnPremisesClient
	CloudAPIClient        ServiceCloudAPIClient
	MessageStore          ServiceMessageStore
	Credentials           *config.WhatsappCloudAPICredentials
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

func (s *Service) GetAPIType() config.WhatsappAPIType {
	return s.WhatsappConfig.GetAPIType(s.GlobalWhatsappAPIType)
}

func (s *Service) SendAuthenticationOTP(ctx context.Context, opts *SendAuthenticationOTPOptions) (*SendAuthenticationOTPResult, error) {
	apiType := s.GetAPIType()
	metricOptions := []otelutil.MetricOption{otelauthgear.WithWhatsappAPIType(apiType)}
	switch apiType {
	case config.WhatsappAPITypeOnPremises:
		if s.OnPremisesClient == nil {
			return nil, ErrNoAvailableWhatsappClient
		}

		otpTemplate := s.OnPremisesClient.GetOTPTemplate()
		lang := s.resolveTemplateLanguage(ctx, otpTemplate.Languages)
		components := s.prepareOTPComponents(otpTemplate, opts.OTP)

		err := s.OnPremisesClient.SendTemplate(
			ctx,
			opts.To,
			otpTemplate,
			lang,
			components)
		if err != nil {
			metricOptions = s.transformFatalErrorIntoLabels(metricOptions, err)
			otelutil.IntCounterAddOne(
				ctx,
				otelauthgear.CounterWhatsappRequestCount,
				metricOptions...,
			)
			return nil, err
		}

		messageStatus := WhatsappMessageStatusDelivered
		metricOptions = append(metricOptions, otelauthgear.WithStatusOk())
		metricOptions = append(metricOptions, otelauthgear.WithWhatsappAPIMessageStatusAndTimeout(string(messageStatus), false)...)
		otelutil.IntCounterAddOne(
			ctx,
			otelauthgear.CounterWhatsappRequestCount,
			metricOptions...,
		)

		return &SendAuthenticationOTPResult{
			MessageID: unknownMessageID,
			// We don't know the actual status, so always return devlivered
			MessageStatus: messageStatus,
		}, nil
	case config.WhatsappAPITypeCloudAPI:
		if s.CloudAPIClient == nil {
			return nil, ErrNoAvailableWhatsappClient
		}

		configuredLanguages := s.CloudAPIClient.GetLanguages()
		lang := s.resolveTemplateLanguage(ctx, configuredLanguages)
		cloudAPISendResult, err := s.CloudAPIClient.SendAuthenticationOTP(
			ctx,
			opts,
			lang,
		)
		if err != nil {
			// If the error is reported here, then it is a fatal error like
			// invalid phone number ID, invalid access token, etc.
			// In this case, we increment the counter here.
			metricOptions = s.transformFatalErrorIntoLabels(metricOptions, err)
			otelutil.IntCounterAddOne(
				ctx,
				otelauthgear.CounterWhatsappRequestCount,
				metricOptions...,
			)
			return nil, err
		}
		result := SendAuthenticationOTPResult(*cloudAPISendResult)

		if s.Credentials.Webhook != nil {
			// If webhook is configured, then we defer the increment of the counter until we receive the webhook, or timeout.
			go func() {
				// Detach the deadline so that the context is not canceled along with the request.
				ctx := context.WithoutCancel(ctx)
				time.Sleep(s.WhatsappConfig.MessageSentCallbackTimeout.Duration())
				// Mark the message as failed after timeout
				keyWasSet, err := s.MessageStore.SetMessageStatusIfNotExist(ctx, result.MessageID, &WhatsappMessageStatusData{
					Status:    WhatsappMessageStatusFailed,
					IsTimeout: true,
				})
				if err != nil {
					metricOptions = s.transformFatalErrorIntoLabels(metricOptions, err)
					logger.GetLogger(ctx).
						WithError(err).
						Error(ctx, "failed to update whatsapp message status")
				} else if keyWasSet {
					messageStatus := WhatsappMessageStatusFailed
					isTimeout := true
					metricOptions = append(metricOptions, otelauthgear.WithStatusError())
					metricOptions = append(metricOptions, otelauthgear.WithWhatsappAPIMessageStatusAndTimeout(string(messageStatus), isTimeout)...)
				}
				otelutil.IntCounterAddOne(
					ctx,
					otelauthgear.CounterWhatsappRequestCount,
					metricOptions...,
				)
			}()
		} else {
			// If webhook is not configured, set it to delivered immediately.
			// And we also increment the counter with status=ok
			messageStatus := WhatsappMessageStatusDelivered
			isTimeout := false

			metricOptions = append(metricOptions, otelauthgear.WithStatusOk())
			metricOptions = append(metricOptions, otelauthgear.WithWhatsappAPIMessageStatusAndTimeout(string(messageStatus), isTimeout)...)

			otelutil.IntCounterAddOne(
				ctx,
				otelauthgear.CounterWhatsappRequestCount,
				metricOptions...,
			)

			_, err := s.MessageStore.SetMessageStatusIfNotExist(ctx, result.MessageID, &WhatsappMessageStatusData{
				Status:    messageStatus,
				IsTimeout: isTimeout,
			})
			if err != nil {
				return nil, err
			}
		}

		return &result, nil

	default:
		panic(fmt.Errorf("whatsapp: unknown api type"))
	}
}

func (s *Service) SendSuppressedAuthenticationOTP(ctx context.Context, opts *SendAuthenticationOTPOptions) (*SendAuthenticationOTPResult, error) {
	return &SendAuthenticationOTPResult{
		MessageID:     suppressedMessageID,
		MessageStatus: WhatsappMessageStatusSent,
	}, nil
}

func (s *Service) UpdateMessageStatus(ctx context.Context, messageID string, status WhatsappMessageStatus, errors []WhatsappStatusError) error {
	data := &WhatsappMessageStatusData{
		Status: status,
		Errors: errors,
	}
	apiType := s.GetAPIType()
	metricOptions := []otelutil.MetricOption{otelauthgear.WithWhatsappAPIType(apiType)}
	metricOptions = s.transformMessageStatusDataIntoLabels(metricOptions, data)
	otelutil.IntCounterAddOne(
		ctx,
		otelauthgear.CounterWhatsappRequestCount,
		metricOptions...,
	)
	return s.MessageStore.UpdateMessageStatus(ctx, messageID, data)
}

func (s *Service) GetMessageStatus(ctx context.Context, messageID string) (*GetMessageStatusResult, error) {
	switch messageID {
	case suppressedMessageID, unknownMessageID:
		// If the message was suppressed or not known, treat it as sent
		return &GetMessageStatusResult{
			Status:   WhatsappMessageStatusSent,
			APIError: nil,
		}, nil
	}
	data, err := s.MessageStore.GetMessageStatus(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	var apierr *apierrors.APIError
	if data.IsTimeout {
		apierr = apierrors.AsAPIErrorWithContext(ctx, ErrWhatsappMessageStatusCallbackTimeout)
	} else if len(data.Errors) > 0 {
		// See https://developers.facebook.com/docs/whatsapp/cloud-api/support/error-codes/
		// Message Undeliverable
		if data.Errors[0].Code == 131026 {
			apierr = apierrors.AsAPIErrorWithContext(ctx, ErrWhatsappUndeliverable)
		} else {
			logger.GetLogger(ctx).With(
				slog.Int("error_code", data.Errors[0].Code),
			).Error(ctx, "unexpected whatsapp status error")
			apierr = apierrors.AsAPIErrorWithContext(ctx, ErrUnexpectedWhatsappMessageStatusError)
		}
	}
	return &GetMessageStatusResult{
		Status:   data.Status,
		APIError: apierr,
	}, nil
}

func (s *Service) transformFatalErrorIntoLabels(metricOptions []otelutil.MetricOption, err error) []otelutil.MetricOption {
	messageStatus := WhatsappMessageStatusFailed
	isTimeout := false

	metricOptions = append(metricOptions, otelauthgear.WithStatusError())
	metricOptions = append(metricOptions, otelauthgear.WithWhatsappAPIMessageStatusAndTimeout(string(messageStatus), isTimeout)...)

	var apiErr *WhatsappAPIError
	if ok := errors.As(err, &apiErr); ok {
		metricOptions = append(metricOptions, otelauthgear.WithHTTPStatusCode(apiErr.HTTPStatusCode))
		if errorCode, ok := apiErr.GetErrorCode(); ok {
			metricOptions = append(metricOptions, otelauthgear.WithWhatsappAPIErrorCode(errorCode))
		}
		if errorSubcode, ok := apiErr.GetErrorSubcode(); ok {
			metricOptions = append(metricOptions, otelauthgear.WithWhatsappAPIErrorSubcode(errorSubcode))
		}
	}

	return metricOptions
}

func (s *Service) transformMessageStatusDataIntoLabels(metricOptions []otelutil.MetricOption, data *WhatsappMessageStatusData) []otelutil.MetricOption {
	if len(data.Errors) > 0 || data.IsTimeout {
		metricOptions = append(metricOptions, otelauthgear.WithStatusError())
	} else {
		metricOptions = append(metricOptions, otelauthgear.WithStatusOk())
	}

	metricOptions = append(metricOptions, otelauthgear.WithWhatsappAPIMessageStatusAndTimeout(string(data.Status), data.IsTimeout)...)

	if len(data.Errors) > 0 {
		metricOptions = append(metricOptions, otelauthgear.WithWhatsappAPIErrorCode(data.Errors[0].Code))
	}

	return metricOptions
}
