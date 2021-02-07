package otp

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/tasks"
	"github.com/authgear/authgear-server/pkg/lib/translation"
)

type SendOptions struct {
	OTP         string
	URL         string
	MessageType MessageType
}

type EndpointsProvider interface {
	BaseURL() *url.URL
}

type TranslationService interface {
	EmailMessageData(msg *translation.MessageSpec, args interface{}) (*translation.EmailMessageData, error)
	SMSMessageData(msg *translation.MessageSpec, args interface{}) (*translation.SMSMessageData, error)
}

type RateLimiter interface {
	TakeToken(bucket ratelimit.Bucket) error
}

type MessageSender struct {
	Translation TranslationService
	Endpoints   EndpointsProvider
	RateLimiter RateLimiter
	TaskQueue   task.Queue
}

func (s *MessageSender) makeData(opts SendOptions) (*MessageTemplateContext, error) {
	ctx := &MessageTemplateContext{
		// To be filled by caller
		Email: "",
		Phone: "",
		Code:  opts.OTP,
		URL:   opts.URL,
		Host:  s.Endpoints.BaseURL().Host,
	}

	return ctx, nil
}

func (s *MessageSender) SendEmail(email string, opts SendOptions) error {
	data, err := s.makeData(opts)
	if err != nil {
		return err
	}
	data.Email = email

	var spec *translation.MessageSpec
	switch opts.MessageType {
	case MessageTypeVerification:
		spec = messageVerification
	case MessageTypeSetupPrimaryOOB:
		spec = messageSetupPrimaryOOB
	case MessageTypeSetupSecondaryOOB:
		spec = messageSetupSecondaryOOB
	case MessageTypeAuthenticatePrimaryOOB:
		spec = messageAuthenticatePrimaryOOB
	case MessageTypeAuthenticateSecondaryOOB:
		spec = messageAuthenticateSecondaryOOB
	default:
		panic("otp: unknown message type: " + opts.MessageType)
	}

	msg, err := s.Translation.EmailMessageData(spec, data)
	if err != nil {
		return err
	}

	err = s.RateLimiter.TakeToken(mail.RateLimitBucket(email))
	if err != nil {
		return err
	}

	s.TaskQueue.Enqueue(&tasks.SendMessagesParam{
		EmailMessages: []mail.SendOptions{{
			Sender:    msg.Sender,
			ReplyTo:   msg.ReplyTo,
			Subject:   msg.Subject,
			Recipient: email,
			TextBody:  msg.TextBody,
			HTMLBody:  msg.HTMLBody,
		}},
	})

	return nil
}

func (s *MessageSender) SendSMS(phone string, opts SendOptions) (err error) {
	data, err := s.makeData(opts)
	if err != nil {
		return err
	}
	data.Phone = phone

	var spec *translation.MessageSpec
	switch opts.MessageType {
	case MessageTypeVerification:
		spec = messageVerification
	case MessageTypeSetupPrimaryOOB:
		spec = messageSetupPrimaryOOB
	case MessageTypeSetupSecondaryOOB:
		spec = messageSetupSecondaryOOB
	case MessageTypeAuthenticatePrimaryOOB:
		spec = messageAuthenticatePrimaryOOB
	case MessageTypeAuthenticateSecondaryOOB:
		spec = messageAuthenticateSecondaryOOB
	default:
		panic("otp: unknown message type: " + opts.MessageType)
	}

	msg, err := s.Translation.SMSMessageData(spec, data)
	if err != nil {
		return err
	}

	err = s.RateLimiter.TakeToken(sms.RateLimitBucket(phone))
	if err != nil {
		return err
	}

	s.TaskQueue.Enqueue(&tasks.SendMessagesParam{
		SMSMessages: []sms.SendOptions{{
			Sender: msg.Sender,
			To:     phone,
			Body:   msg.Body,
		}},
	})

	return
}
