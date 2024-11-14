package messaging

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("messaging")}
}

type EventService interface {
	DispatchEventImmediately(ctx context.Context, payload event.NonBlockingPayload) error
}

type MailSender interface {
	Send(opts mail.SendOptions) error
}

type SMSSender interface {
	Send(ctx context.Context, opts sms.SendOptions) error
}

type WhatsappSender interface {
	SendAuthenticationOTP(Ctx context.Context, opts *whatsapp.SendAuthenticationOTPOptions) error
}

type Sender struct {
	Logger         Logger
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

func (s *Sender) SendEmailInNewGoroutine(ctx context.Context, msgType translation.MessageType, opts *mail.SendOptions) error {
	err := s.Limits.checkEmail(ctx, opts.Recipient)
	if err != nil {
		return err
	}

	if s.FeatureTestModeEmailSuppressed {
		s.testModeSendEmail(opts)
		return nil
	}

	if s.TestModeEmailConfig.Enabled {
		if r, ok := s.TestModeEmailConfig.MatchTarget(opts.Recipient); ok && r.Suppressed {
			s.testModeSendEmail(opts)
			return nil
		}
	}

	if s.DevMode {
		s.devModeSendEmail(opts)
		return nil
	}

	go func() {
		// Detach the deadline so that the context is not canceled along with the request.
		ctx = context.WithoutCancel(ctx)

		err := s.MailSender.Send(*opts)
		if err != nil {
			s.Logger.WithError(err).WithFields(logrus.Fields{
				"email": mail.MaskAddress(opts.Recipient),
			}).Error("failed to send email")
			return
		}

		err = s.Database.WithTx(ctx, func(ctx context.Context) error {
			return s.Events.DispatchEventImmediately(ctx, &nonblocking.EmailSentEventPayload{
				Sender:    opts.Sender,
				Recipient: opts.Recipient,
				Type:      string(msgType),
			})
		})
		if err != nil {
			s.Logger.Error("failed to emit email.sent event")
		}
	}()

	return nil
}

func (s *Sender) testModeSendEmail(opts *mail.SendOptions) {
	s.Logger.
		WithField("recipient", opts.Recipient).
		WithField("body", opts.TextBody).
		WithField("sender", opts.Sender).
		WithField("subject", opts.Subject).
		WithField("reply_to", opts.ReplyTo).
		Warn("sending email is suppressed by test mode")
}

func (s *Sender) devModeSendEmail(opts *mail.SendOptions) {
	s.Logger.
		WithField("recipient", opts.Recipient).
		WithField("body", opts.TextBody).
		WithField("sender", opts.Sender).
		WithField("subject", opts.Subject).
		WithField("reply_to", opts.ReplyTo).
		Warn("skip sending email in development mode")
}

func (s *Sender) SendSMSInNewGoroutine(ctx context.Context, msgType translation.MessageType, opts *sms.SendOptions) error {
	err := s.Limits.checkSMS(ctx, opts.To)
	if err != nil {
		return err
	}

	if s.FeatureTestModeSMSSuppressed {
		s.testModeSendSMS(opts)
		return nil
	}

	if s.TestModeSMSConfig.Enabled {
		if r, ok := s.TestModeSMSConfig.MatchTarget(opts.To); ok && r.Suppressed {
			s.testModeSendSMS(opts)
			return nil
		}
	}

	if s.DevMode {
		s.devModeSendSMS(opts)
		return nil
	}

	go func() {
		// Detach the deadline so that the context is not canceled along with the request.
		ctx = context.WithoutCancel(ctx)

		err := s.SMSSender.Send(ctx, *opts)
		if err != nil {
			s.Logger.WithError(err).WithFields(logrus.Fields{
				"phone": phone.Mask(opts.To),
			}).Error("failed to send SMS")
			return
		}

		err = s.Database.WithTx(ctx, func(ctx context.Context) error {
			return s.Events.DispatchEventImmediately(ctx, &nonblocking.SMSSentEventPayload{
				Sender:              opts.Sender,
				Recipient:           opts.To,
				Type:                string(msgType),
				IsNotCountedInUsage: s.MessagingFeatureConfig.SMSUsageCountDisabled,
			})
		})
		if err != nil {
			s.Logger.Error("failed to emit sms.sent event")
		}
	}()

	return nil
}

func (s *Sender) testModeSendSMS(opts *sms.SendOptions) {
	s.Logger.
		WithField("recipient", opts.To).
		WithField("sender", opts.Sender).
		WithField("body", opts.Body).
		WithField("app_id", opts.AppID).
		WithField("template_name", opts.TemplateName).
		WithField("language_tag", opts.LanguageTag).
		WithField("template_variables", opts.TemplateVariables).
		Warn("sending SMS is suppressed in test mode")
}

func (s *Sender) devModeSendSMS(opts *sms.SendOptions) {
	s.Logger.
		WithField("recipient", opts.To).
		WithField("sender", opts.Sender).
		WithField("body", opts.Body).
		WithField("app_id", opts.AppID).
		WithField("template_name", opts.TemplateName).
		WithField("language_tag", opts.LanguageTag).
		WithField("template_variables", opts.TemplateVariables).
		Warn("skip sending SMS in development mode")
}

func (s *Sender) SendWhatsappImmediately(ctx context.Context, msgType translation.MessageType, opts *whatsapp.SendAuthenticationOTPOptions) error {
	err := s.Limits.checkWhatsapp(ctx, opts.To)
	if err != nil {
		return err
	}

	if s.FeatureTestModeWhatsappSuppressed {
		s.testModeSendWhatsapp(opts)
		return nil
	}

	if s.TestModeWhatsappConfig.Enabled {
		if r, ok := s.TestModeWhatsappConfig.MatchTarget(opts.To); ok && r.Suppressed {
			s.testModeSendWhatsapp(opts)
			return nil
		}
	}

	if s.DevMode {
		s.devModeSendWhatsapp(opts)
		return nil
	}

	// Send immediately.
	err = s.WhatsappSender.SendAuthenticationOTP(ctx, opts)
	if err != nil {
		return err
	}

	err = s.Events.DispatchEventImmediately(ctx, &nonblocking.WhatsappSentEventPayload{
		Recipient:           opts.To,
		Type:                string(msgType),
		IsNotCountedInUsage: s.MessagingFeatureConfig.WhatsappUsageCountDisabled,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Sender) testModeSendWhatsapp(opts *whatsapp.SendAuthenticationOTPOptions) {
	s.Logger.
		WithField("recipient", opts.To).
		WithField("otp", opts.OTP).
		Warn("sending whatsapp is suppressed in test mode")
}

func (s *Sender) devModeSendWhatsapp(opts *whatsapp.SendAuthenticationOTPOptions) {
	s.Logger.
		WithField("recipient", opts.To).
		WithField("otp", opts.OTP).
		Warn("skip sending whatsapp in development mode")
}
