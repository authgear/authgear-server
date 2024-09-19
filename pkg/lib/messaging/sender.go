package messaging

import (
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
	DispatchEventImmediately(payload event.NonBlockingPayload) error
}

type Sender struct {
	Limits                 Limits
	TaskQueue              task.Queue
	Events                 EventService
	Whatsapp               *whatsapp.Service
	MessagingFeatureConfig *config.MessagingFeatureConfig
}

func (s *Sender) PrepareEmail(email string, msgType translation.MessageType) (*EmailMessage, error) {
	msg, err := s.Limits.checkEmail(email)
	if err != nil {
		return nil, err
	}

	return &EmailMessage{
		message:     *msg,
		taskQueue:   s.TaskQueue,
		events:      s.Events,
		SendOptions: mail.SendOptions{Recipient: email},
		Type:        msgType,
	}, nil
}

func (s *Sender) PrepareSMS(phoneNumber string, msgType translation.MessageType) (*SMSMessage, error) {
	msg, err := s.Limits.checkSMS(phoneNumber)
	if err != nil {
		return nil, err
	}

	return &SMSMessage{
		message:      *msg,
		taskQueue:    s.TaskQueue,
		events:       s.Events,
		SendOptions:  sms.SendOptions{To: phoneNumber},
		Type:         msgType,
		IsNotCounted: s.MessagingFeatureConfig.SMSUsageCountDisabled,
	}, nil
}

func (s *Sender) PrepareWhatsapp(phoneNumber string, msgType translation.MessageType) (*WhatsappMessage, error) {
	msg, err := s.Limits.checkWhatsapp(phoneNumber)
	if err != nil {
		return nil, err
	}

	return &WhatsappMessage{
		message:      *msg,
		taskQueue:    s.TaskQueue,
		events:       s.Events,
		Options:      whatsapp.SendTemplateOptions{To: phoneNumber},
		Type:         msgType,
		IsNotCounted: s.MessagingFeatureConfig.WhatsappUsageCountDisabled,
	}, nil
}
