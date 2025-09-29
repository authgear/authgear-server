package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"gopkg.in/gomail.v2"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/otelutil"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var SenderLogger = slogutil.NewLogger("messaging")

type EventService interface {
	DispatchEventImmediately(ctx context.Context, payload event.NonBlockingPayload) error
}

type MailSender interface {
	PrepareMessage(opts mail.SendOptions) (*gomail.Message, error)
	Send(*gomail.Message) error
}

type SMSSender interface {
	Send(ctx context.Context, client smsapi.Client, opts sms.SendOptions) error
	ResolveClient() (smsapi.Client, error)
}

type WhatsappSender interface {
	SendAuthenticationOTP(ctx context.Context, opts *whatsapp.SendAuthenticationOTPOptions) (*whatsapp.SendAuthenticationOTPResult, error)
}

type Sender struct {
	Limits         Limits
	Events         EventService
	MailSender     MailSender
	SMSSender      SMSSender
	WhatsappSender WhatsappSender
	Database       *appdb.Handle

	DevMode config.DevMode

	MessagingFeatureConfig *config.MessagingFeatureConfig

	FeatureTestModeEmailSuppressed config.FeatureTestModeEmailSuppressed
	TestModeEmailConfig            *config.TestModeEmailConfig

	FeatureTestModeSMSSuppressed config.FeatureTestModeSMSSuppressed
	TestModeSMSConfig            *config.TestModeSMSConfig

	FeatureTestModeWhatsappSuppressed config.FeatureTestModeWhatsappSuppressed
	TestModeWhatsappConfig            *config.TestModeWhatsappConfig
}

type SendWhatsappResult struct {
	MessageID     string
	MessageStatus whatsapp.WhatsappMessageStatus
}

func (s *Sender) SendEmailInNewGoroutine(ctx context.Context, msgType translation.MessageType, opts *mail.SendOptions) error {
	err := s.Limits.checkEmail(ctx, opts.Recipient)
	if err != nil {
		return err
	}

	if s.FeatureTestModeEmailSuppressed {
		return s.testModeSendEmail(ctx, msgType, opts)
	}

	if s.TestModeEmailConfig.Enabled {
		if r, ok := s.TestModeEmailConfig.MatchTarget(opts.Recipient); ok && r.Suppressed {
			return s.testModeSendEmail(ctx, msgType, opts)
		}
	}

	if s.DevMode {
		return s.devModeSendEmail(ctx, msgType, opts)
	}

	message, err := s.MailSender.PrepareMessage(*opts)
	if err != nil {
		return err
	}

	sendInTx := func(ctx context.Context) error {
		logger := SenderLogger.GetLogger(ctx)
		err := s.MailSender.Send(message)
		if err != nil {
			// Log the send error immediately.
			logger.WithError(err).With(
				slog.String("email", mail.MaskAddress(opts.Recipient)),
			).Error(ctx, "failed to send email")

			otelutil.IntCounterAddOne(
				ctx,
				otelauthgear.CounterEmailRequestCount,
				otelauthgear.WithStatusError(),
			)

			dispatchErr := s.DispatchEventImmediatelyWithTx(ctx, &nonblocking.EmailErrorEventPayload{
				Description: s.errorToDescription(err),
			})
			if dispatchErr != nil {
				logger.WithError(dispatchErr).Error(ctx, "failed to emit event", slog.String("event", string(nonblocking.EmailError)))
			}
			return err
		}

		otelutil.IntCounterAddOne(
			ctx,
			otelauthgear.CounterEmailRequestCount,
			otelauthgear.WithStatusOk(),
		)

		dispatchErr := s.DispatchEventImmediatelyWithTx(ctx, &nonblocking.EmailSentEventPayload{
			Sender:    opts.Sender,
			Recipient: opts.Recipient,
			Type:      string(msgType),
		})
		if dispatchErr != nil {
			logger.WithError(dispatchErr).Error(ctx, "failed to emit event", slog.String("event", string(nonblocking.EmailSent)))
		}
		return nil
	}

	// Detach the deadline so that the context is not canceled along with the request.
	ctxWithoutCancel := context.WithoutCancel(ctx)
	go func(ctx context.Context) {
		// Always use a new transaction to send in async routine
		// No need to handle the error as sendInTx is assumed to have handle it by logging.
		_ = s.Database.ReadOnly(ctx, func(ctx context.Context) error {
			return sendInTx(ctx)
		})
	}(ctxWithoutCancel)

	return nil
}

func (s *Sender) testModeSendEmail(ctx context.Context, msgType translation.MessageType, opts *mail.SendOptions) error {
	logger := SenderLogger.GetLogger(ctx)

	logger.With(
		slog.String("message_type", string(msgType)),
		slog.String("recipient", opts.Recipient),
		slog.String("body", opts.TextBody),
		slog.String("sender", opts.Sender),
		slog.String("subject", opts.Subject),
		slog.String("reply_to", opts.ReplyTo),
	).Info(ctx, "email is suppressed by test mode")

	desc := fmt.Sprintf("email (%v) to %v is suppressed by test mode.", msgType, opts.Recipient)
	return s.DispatchEventImmediatelyWithTx(ctx, &nonblocking.EmailSuppressedEventPayload{
		Description: desc,
	})
}

func (s *Sender) devModeSendEmail(ctx context.Context, msgType translation.MessageType, opts *mail.SendOptions) error {
	logger := SenderLogger.GetLogger(ctx)

	logger.With(
		slog.String("message_type", string(msgType)),
		slog.String("recipient", opts.Recipient),
		slog.String("body", opts.TextBody),
		slog.String("sender", opts.Sender),
		slog.String("subject", opts.Subject),
		slog.String("reply_to", opts.ReplyTo),
	).Info(ctx, "email is suppressed by development mode")

	desc := fmt.Sprintf("email (%v) to %v is suppressed by development mode", msgType, opts.Recipient)
	return s.DispatchEventImmediatelyWithTx(ctx, &nonblocking.EmailSuppressedEventPayload{
		Description: desc,
	})
}

func (s *Sender) SendSMSInNewGoroutine(ctx context.Context, msgType translation.MessageType, opts *sms.SendOptions) error {
	return s.sendSMS(ctx, msgType, opts, true)
}

func (s *Sender) SendSMSImmediately(ctx context.Context, msgType translation.MessageType, opts *sms.SendOptions) error {
	return s.sendSMS(ctx, msgType, opts, false)
}

func (s *Sender) sendSMS(ctx context.Context, msgType translation.MessageType, opts *sms.SendOptions, isAsync bool) error {
	logger := SenderLogger.GetLogger(ctx)

	err := s.Limits.checkSMS(ctx, opts.To)
	if err != nil {
		return err
	}

	if s.FeatureTestModeSMSSuppressed {
		return s.testModeSendSMS(ctx, msgType, opts)
	}

	if s.TestModeSMSConfig.Enabled {
		if r, ok := s.TestModeSMSConfig.MatchTarget(opts.To); ok && r.Suppressed {
			return s.testModeSendSMS(ctx, msgType, opts)
		}
	}

	if s.DevMode {
		return s.devModeSendSMS(ctx, msgType, opts)
	}

	client, err := s.SMSSender.ResolveClient()
	if err != nil {
		return err
	}

	sendInTx := func(ctx context.Context) error {
		logger := SenderLogger.GetLogger(ctx)

		err = s.SMSSender.Send(ctx, client, *opts)
		if err != nil {
			// Log the send error immediately.
			// TODO: Handle expected errors https://linear.app/authgear/issue/DEV-1139
			logger.WithError(err).With(
				slog.String("phone", phone.Mask(opts.To)),
			).Error(ctx, "failed to send SMS")

			var smsapiErr *smsapi.SendError
			if errors.As(err, &smsapiErr) && smsapiErr.APIErrorKind != nil {
				otelutil.IntCounterAddOne(
					ctx,
					otelauthgear.CounterSMSRequestCount,
					otelauthgear.WithStatusError(),
					otelauthgear.WithAPIErrorReason(smsapiErr.APIErrorKind.Reason),
				)
			} else {
				otelutil.IntCounterAddOne(
					ctx,
					otelauthgear.CounterSMSRequestCount,
					otelauthgear.WithStatusError(),
				)
			}

			dispatchErr := s.DispatchEventImmediatelyWithTx(ctx, &nonblocking.SMSErrorEventPayload{
				Description: s.errorToDescription(err),
			})
			if dispatchErr != nil {
				logger.WithError(dispatchErr).Error(ctx, "failed to emit event", slog.String("event", string(nonblocking.SMSError)))
			}
			return err
		}
		return nil
	}

	if isAsync {
		// Detach the deadline so that the context is not canceled along with the request.
		ctxWithoutCancel := context.WithoutCancel(ctx)
		go func(ctx context.Context) {
			// Always use a new transaction to send in async routine
			// No need to handle the error as sendInTx is assumed to have handle it by logging.
			_ = s.Database.ReadOnly(ctx, func(ctx context.Context) error {
				return sendInTx(ctx)
			})
		}(ctxWithoutCancel)
	} else {
		err = sendInTx(ctx)
		if err != nil {
			return err
		}
	}

	otelutil.IntCounterAddOne(
		ctx,
		otelauthgear.CounterSMSRequestCount,
		otelauthgear.WithStatusOk(),
	)

	dispatchErr := s.DispatchEventImmediatelyWithTx(ctx, &nonblocking.SMSSentEventPayload{
		Sender:              opts.Sender,
		Recipient:           opts.To,
		Type:                string(msgType),
		IsNotCountedInUsage: *s.MessagingFeatureConfig.SMSUsageCountDisabled,
	})
	if dispatchErr != nil {
		logger.WithError(dispatchErr).Error(ctx, "failed to emit event", slog.String("event", string(nonblocking.SMSSent)))
	}

	return nil
}

func (s *Sender) testModeSendSMS(ctx context.Context, msgType translation.MessageType, opts *sms.SendOptions) error {
	logger := SenderLogger.GetLogger(ctx)
	logger.With(
		slog.String("message_type", string(msgType)),
		slog.String("recipient", opts.To),
		slog.String("sender", opts.Sender),
		slog.String("body", opts.Body),
		slog.String("app_id", opts.AppID),
		slog.String("template_name", opts.TemplateName),
		slog.String("language_tag", opts.LanguageTag),
		slog.Any("template_variables", opts.TemplateVariables),
	).Info(ctx, "SMS is suppressed in test mode")

	desc := fmt.Sprintf("SMS (%v) to %v is suppressed by test mode.", msgType, opts.To)
	return s.DispatchEventImmediatelyWithTx(ctx, &nonblocking.SMSSuppressedEventPayload{
		Description: desc,
	})
}

func (s *Sender) devModeSendSMS(ctx context.Context, msgType translation.MessageType, opts *sms.SendOptions) error {
	logger := SenderLogger.GetLogger(ctx)
	logger.With(
		slog.String("message_type", string(msgType)),
		slog.String("recipient", opts.To),
		slog.String("sender", opts.Sender),
		slog.String("body", opts.Body),
		slog.String("app_id", opts.AppID),
		slog.String("template_name", opts.TemplateName),
		slog.String("language_tag", opts.LanguageTag),
		slog.Any("template_variables", opts.TemplateVariables),
	).Info(ctx, "SMS is suppressed in development mode")

	desc := fmt.Sprintf("SMS (%v) to %v is suppressed by development mode.", msgType, opts.To)
	return s.DispatchEventImmediatelyWithTx(ctx, &nonblocking.SMSSuppressedEventPayload{
		Description: desc,
	})
}

func (s *Sender) SendWhatsappImmediately(ctx context.Context, msgType translation.MessageType, opts *whatsapp.SendAuthenticationOTPOptions) (*SendWhatsappResult, error) {
	logger := SenderLogger.GetLogger(ctx)
	err := s.Limits.checkWhatsapp(ctx, opts.To)
	if err != nil {
		return nil, err
	}

	if s.FeatureTestModeWhatsappSuppressed {
		return s.testModeSendWhatsapp(ctx, msgType, opts)
	}

	if s.TestModeWhatsappConfig.Enabled {
		if r, ok := s.TestModeWhatsappConfig.MatchTarget(opts.To); ok && r.Suppressed {
			return s.testModeSendWhatsapp(ctx, msgType, opts)
		}
	}

	if s.DevMode {
		return s.devModeSendWhatsapp(ctx, msgType, opts)
	}

	// Send immediately.
	result, err := s.sendWhatsapp(ctx, opts)
	if err != nil {
		// Log the send error immediately.
		logger.WithError(err).With(
			slog.String("phone", phone.Mask(opts.To)),
		).Error(ctx, "failed to send Whatsapp")

		metricOptions := []otelutil.MetricOption{otelauthgear.WithStatusError()}
		var apiErr *whatsapp.WhatsappAPIError
		if ok := errors.As(err, &apiErr); ok {
			metricOptions = append(metricOptions, otelauthgear.WithWhatsappAPIType(apiErr.APIType))
			metricOptions = append(metricOptions, otelauthgear.WithHTTPStatusCode(apiErr.HTTPStatusCode))
			errorCode, ok := apiErr.GetErrorCode()
			if ok {
				metricOptions = append(metricOptions, otelauthgear.WithWhatsappAPIErrorCode(errorCode))
			}
		}

		otelutil.IntCounterAddOne(
			ctx,
			otelauthgear.CounterWhatsappRequestCount,
			metricOptions...,
		)

		dispatchErr := s.DispatchEventImmediatelyWithTx(ctx, &nonblocking.WhatsappErrorEventPayload{
			Description: s.errorToDescription(err),
		})
		if dispatchErr != nil {
			logger.WithError(dispatchErr).Error(ctx, "failed to emit event", slog.String("event", string(nonblocking.WhatsappError)))
		}

		return nil, err
	}

	otelutil.IntCounterAddOne(
		ctx,
		otelauthgear.CounterWhatsappRequestCount,
		otelauthgear.WithStatusOk(),
	)

	dispatchErr := s.DispatchEventImmediatelyWithTx(ctx, &nonblocking.WhatsappSentEventPayload{
		Recipient:           opts.To,
		Type:                string(msgType),
		IsNotCountedInUsage: *s.MessagingFeatureConfig.WhatsappUsageCountDisabled,
	})
	if dispatchErr != nil {
		logger.WithError(dispatchErr).Error(ctx, "failed to emit %v event", slog.String("event", string(nonblocking.WhatsappSent)))
	}

	return result, nil
}

func (s *Sender) sendWhatsapp(ctx context.Context, opts *whatsapp.SendAuthenticationOTPOptions) (*SendWhatsappResult, error) {
	senderResult, err := s.WhatsappSender.SendAuthenticationOTP(ctx, opts)
	if err != nil {
		return nil, err
	}
	result := SendWhatsappResult(*senderResult)
	return &result, nil
}

func (s *Sender) testModeSendWhatsapp(ctx context.Context, msgType translation.MessageType, opts *whatsapp.SendAuthenticationOTPOptions) (*SendWhatsappResult, error) {
	logger := SenderLogger.GetLogger(ctx)
	logger.With(
		slog.String("message_type", string(msgType)),
		slog.String("recipient", opts.To),
		slog.String("otp", opts.OTP),
	).Info(ctx, "Whatsapp is suppressed in test mode")

	desc := fmt.Sprintf("Whatsapp (%v) to %v is suppressed by test mode.", msgType, opts.To)
	err := s.DispatchEventImmediatelyWithTx(ctx, &nonblocking.WhatsappSuppressedEventPayload{
		Description: desc,
	})
	if err != nil {
		return nil, err
	}
	return &SendWhatsappResult{
		MessageID:     "",
		MessageStatus: whatsapp.WhatsappMessageStatusDelivered,
	}, nil
}

func (s *Sender) devModeSendWhatsapp(ctx context.Context, msgType translation.MessageType, opts *whatsapp.SendAuthenticationOTPOptions) (*SendWhatsappResult, error) {
	logger := SenderLogger.GetLogger(ctx)
	logger.With(
		slog.String("message_type", string(msgType)),
		slog.String("recipient", opts.To),
		slog.String("otp", opts.OTP),
	).Info(ctx, "Whatsapp is suppressed in development mode")

	desc := fmt.Sprintf("Whatsapp (%v) to %v is suppressed by development mode.", msgType, opts.To)
	err := s.DispatchEventImmediatelyWithTx(ctx, &nonblocking.WhatsappSuppressedEventPayload{
		Description: desc,
	})
	if err != nil {
		return nil, err
	}
	return &SendWhatsappResult{
		MessageID:     "",
		MessageStatus: whatsapp.WhatsappMessageStatusDelivered,
	}, nil
}

func (s *Sender) DispatchEventImmediatelyWithTx(ctx context.Context, payload event.NonBlockingPayload) error {
	if s.Database.IsInTx(ctx) {
		return s.Events.DispatchEventImmediately(ctx, payload)
	}

	return s.Database.ReadOnly(ctx, func(ctx context.Context) error {
		return s.Events.DispatchEventImmediately(ctx, payload)
	})
}

func (s *Sender) errorToDescription(err error) string {
	// APIError.Error() shows message only, but we want to show the full content of it.
	// Modifying APIError.Error is another big change that I do not want to deal with here.
	if apierrors.IsAPIError(err) {
		apiError := apierrors.AsAPIError(err)
		b, err := json.Marshal(apiError)
		if err != nil {
			panic(err)
		}
		return string(b)
	}

	return err.Error()
}
