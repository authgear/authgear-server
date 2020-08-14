package oob

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type OTPMessageSender interface {
	SendEmail(email string, opts otp.SendOptions, message config.EmailMessageConfig) error
	SendSMS(phone string, opts otp.SendOptions, message config.SMSMessageConfig) error
}

type CodeSender struct {
	Config           *config.AuthenticatorOOBConfig
	OTPMessageSender OTPMessageSender
}

func (s *CodeSender) SendCode(
	channel authn.AuthenticatorOOBChannel,
	target string,
	code string,
	messageType otp.MessageType,
) (result *otp.CodeSendResult, err error) {
	opts := otp.SendOptions{
		OTP:         code,
		URL:         "", // FIXME: send a login link to email?
		MessageType: messageType,
	}
	switch channel {
	case authn.AuthenticatorOOBChannelEmail:
		err = s.OTPMessageSender.SendEmail(target, opts, s.Config.Email.Message)
	case authn.AuthenticatorOOBChannelSMS:
		err = s.OTPMessageSender.SendSMS(target, opts, s.Config.SMS.Message)
	default:
		panic("oob: unknown channel type: " + channel)
	}

	if err != nil {
		return
	}

	result = &otp.CodeSendResult{
		Target:       target,
		Channel:      string(channel),
		CodeLength:   len(code),
		SendCooldown: OOBOTPSendCooldownSeconds,
	}
	return
}
