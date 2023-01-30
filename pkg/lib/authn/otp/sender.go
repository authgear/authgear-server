package otp

import (
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
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
	OTPMode     OTPMode
}

type EndpointsProvider interface {
	BaseURL() *url.URL
	MagicLinkVerificationEndpointURL() *url.URL
}

type TranslationService interface {
	EmailMessageData(msg *translation.MessageSpec, args interface{}) (*translation.EmailMessageData, error)
	SMSMessageData(msg *translation.MessageSpec, args interface{}) (*translation.SMSMessageData, error)
}

type HardSMSBucketer interface {
	Bucket() ratelimit.Bucket
}

type RateLimiter interface {
	CheckToken(bucket ratelimit.Bucket) (pass bool, resetDuration time.Duration, err error)
	TakeToken(bucket ratelimit.Bucket) error
	ClearBucket(bucket ratelimit.Bucket) error
}

type EventService interface {
	DispatchEvent(payload event.Payload) error
}

type MessageSender struct {
	Translation TranslationService
	Endpoints   EndpointsProvider
	TaskQueue   task.Queue
	Events      EventService

	RateLimiter     RateLimiter
	HardSMSBucketer HardSMSBucketer
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

	if opts.OTPMode == OTPModeMagicLink {
		url := s.Endpoints.MagicLinkVerificationEndpointURL()
		query := url.Query()
		query.Set("token", data.Code)
		query.Set("target", data.Email)
		url.RawQuery = query.Encode()

		data.URL = url.String()
	}

	var spec *translation.MessageSpec
	var emailType nonblocking.MessageType
	switch opts.MessageType {
	case MessageTypeVerification:
		spec = messageVerification
		emailType = nonblocking.MessageTypeVerification
	case MessageTypeSetupPrimaryOOB:
		spec = messageSetupPrimaryOOB
		emailType = nonblocking.MessageTypeSetupPrimaryOOB
	case MessageTypeSetupSecondaryOOB:
		if opts.OTPMode == OTPModeMagicLink {
			spec = messageSetupSecondaryMagicLink
		} else {
			spec = messageSetupSecondaryOOB
		}
		emailType = nonblocking.MessageTypeSetupSecondaryOOB
	case MessageTypeAuthenticatePrimaryOOB:
		spec = messageAuthenticatePrimaryOOB
		emailType = nonblocking.MessageTypeAuthenticatePrimaryOOB
	case MessageTypeAuthenticateSecondaryOOB:
		if opts.OTPMode == OTPModeMagicLink {
			spec = messageAuthenticateSecondaryMagicLink
		} else {
			spec = messageAuthenticateSecondaryOOB
		}
		emailType = nonblocking.MessageTypeAuthenticateSecondaryOOB
	default:
		panic("otp: unknown message type: " + opts.MessageType)
	}

	msg, err := s.Translation.EmailMessageData(spec, data)
	if err != nil {
		return err
	}

	err = s.RateLimiter.TakeToken(mail.AntiSpamBucket(email))
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

	err = s.Events.DispatchEvent(&nonblocking.EmailSentEventPayload{
		Sender:    msg.Sender,
		Recipient: email,
		Type:      emailType,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *MessageSender) SendSMS(phone string, opts SendOptions) (err error) {
	data, err := s.makeData(opts)
	if err != nil {
		return err
	}
	data.Phone = phone

	var spec *translation.MessageSpec
	var smsType nonblocking.MessageType
	switch opts.MessageType {
	case MessageTypeVerification:
		spec = messageVerification
		smsType = nonblocking.MessageTypeVerification
	case MessageTypeSetupPrimaryOOB:
		spec = messageSetupPrimaryOOB
		smsType = nonblocking.MessageTypeSetupPrimaryOOB
	case MessageTypeSetupSecondaryOOB:
		spec = messageSetupSecondaryOOB
		smsType = nonblocking.MessageTypeSetupSecondaryOOB
	case MessageTypeAuthenticatePrimaryOOB:
		spec = messageAuthenticatePrimaryOOB
		smsType = nonblocking.MessageTypeAuthenticatePrimaryOOB
	case MessageTypeAuthenticateSecondaryOOB:
		spec = messageAuthenticateSecondaryOOB
		smsType = nonblocking.MessageTypeAuthenticateSecondaryOOB
	default:
		panic("otp: unknown message type: " + opts.MessageType)
	}

	msg, err := s.Translation.SMSMessageData(spec, data)
	if err != nil {
		return err
	}

	err = s.RateLimiter.TakeToken(sms.AntiSpamBucket(phone))
	if err != nil {
		return err
	}

	err = s.RateLimiter.TakeToken(s.HardSMSBucketer.Bucket())
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

	err = s.Events.DispatchEvent(&nonblocking.SMSSentEventPayload{
		Sender:    msg.Sender,
		Recipient: phone,
		Type:      smsType,
	})
	if err != nil {
		return err
	}

	return
}
