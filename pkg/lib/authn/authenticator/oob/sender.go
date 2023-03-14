package oob

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
)

type OTPMessageSender interface {
	CanSendEmail(email string) error
	SendEmail(email string, opts otp.SendOptions) error
	CanSendSMS(phone string) error
	SendSMS(phone string, opts otp.SendOptions) error
}

type CodeSender struct {
	OTPMessageSender OTPMessageSender
}

func (s *CodeSender) CanSendCode(
	channel model.AuthenticatorOOBChannel,
	target string,
) error {
	switch channel {
	case model.AuthenticatorOOBChannelEmail:
		return s.OTPMessageSender.CanSendEmail(target)
	case model.AuthenticatorOOBChannelSMS:
		return s.OTPMessageSender.CanSendSMS(target)
	default:
		panic("oob: unknown channel type: " + channel)
	}
}

func (s *CodeSender) SendCode(
	channel model.AuthenticatorOOBChannel,
	target string,
	code string,
	messageType otp.MessageType,
	otpMode otp.OTPMode,
) (err error) {
	opts := otp.SendOptions{
		OTP:         code,
		URL:         "", // TODO(interaction): Include login link in email.
		MessageType: messageType,
		OTPMode:     otpMode,
	}
	switch channel {
	case model.AuthenticatorOOBChannelEmail:
		err = s.OTPMessageSender.SendEmail(target, opts)
	case model.AuthenticatorOOBChannelSMS:
		err = s.OTPMessageSender.SendSMS(target, opts)
	default:
		panic("oob: unknown channel type: " + channel)
	}

	if err != nil {
		return
	}

	return
}
