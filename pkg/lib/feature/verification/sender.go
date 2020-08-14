package verification

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type OTPMessageSender interface {
	SendEmail(email string, opts otp.SendOptions, message config.EmailMessageConfig) error
	SendSMS(phone string, opts otp.SendOptions, message config.SMSMessageConfig) error
}

type WebAppURLProvider interface {
	VerifyIdentityURL(code string, webStateID string) *url.URL
}

type CodeSender struct {
	Config           *config.VerificationConfig
	OTPMessageSender OTPMessageSender
	WebAppURLs       WebAppURLProvider
}

func (s *CodeSender) SendCode(code *Code, webStateID string) (*otp.CodeSendResult, error) {
	opts := otp.SendOptions{
		OTP:         code.Code,
		URL:         s.WebAppURLs.VerifyIdentityURL(code.Code, webStateID).String(),
		MessageType: otp.MessageTypeVerification,
	}

	var err error
	switch config.LoginIDKeyType(code.LoginIDType) {
	case config.LoginIDKeyTypeEmail:
		err = s.OTPMessageSender.SendEmail(code.LoginID, opts, s.Config.Email.Message)
	case config.LoginIDKeyTypePhone:
		err = s.OTPMessageSender.SendSMS(code.LoginID, opts, s.Config.SMS.Message)
	default:
		panic("verification: unsupported login ID type: " + code.LoginIDType)
	}
	if err != nil {
		return nil, err
	}

	return code.SendResult(), nil
}
