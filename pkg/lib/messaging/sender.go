package messaging

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("messaging")}
}

type EventService interface {
	DispatchEventImmediately(ctx context.Context, payload event.NonBlockingPayload) error
}

type Sender struct {
	Limits                 Limits
	TaskQueue              task.Queue
	Events                 EventService
	Whatsapp               *whatsapp.Service
	MessagingFeatureConfig *config.MessagingFeatureConfig
}

func (s *Sender) SendEmailInNewGoroutine(ctx context.Context, msgType translation.MessageType, opts *mail.SendOptions) error {
	err := s.Limits.checkEmail(ctx, opts.Recipient)
	if err != nil {
		return err
	}

	// FIXME(messaging):send the email in a new goroutine.
	return nil
}

func (s *Sender) SendSMSInNewGoroutine(ctx context.Context, msgType translation.MessageType, opts *sms.SendOptions) error {
	err := s.Limits.checkSMS(ctx, opts.To)
	if err != nil {
		return err
	}

	// FIXME(messaging): respect s.MessagingFeatureConfig.SMSUsageCountDisabled
	// FIXME(messaging): send the SMS in a new goroutine.
	return nil
}

func (s *Sender) SendWhatsappImmediately(ctx context.Context, msgType translation.MessageType, opts *whatsapp.SendTemplateOptions) error {
	err := s.Limits.checkWhatsapp(ctx, opts.To)
	if err != nil {
		return err
	}

	// FIXME(messaging): respect s.MessagingFeatureConfig.WhatsappUsageCountDisabled
	// FIXME(messaging): send the Whatsapp immediately.
	return nil
}
