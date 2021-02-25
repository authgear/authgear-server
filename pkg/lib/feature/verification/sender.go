package verification

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type OTPMessageSender interface {
	SendEmail(email string, opts otp.SendOptions) error
	SendSMS(phone string, opts otp.SendOptions) error
}

type WebAppURLProvider interface {
	VerifyIdentityURL(code string, webStateID string) *url.URL
}

type CodeSender struct {
	OTPMessageSender OTPMessageSender
	WebAppURLs       WebAppURLProvider
}

func (s *CodeSender) SendCode(code *Code) error {
	opts := otp.SendOptions{
		OTP:         code.Code,
		URL:         s.WebAppURLs.VerifyIdentityURL(code.Code, code.ID).String(),
		MessageType: otp.MessageTypeVerification,
	}

	var err error
	switch config.LoginIDKeyType(code.LoginIDType) {
	case config.LoginIDKeyTypeEmail:
		err = s.OTPMessageSender.SendEmail(code.LoginID, opts)
	case config.LoginIDKeyTypePhone:
		err = s.OTPMessageSender.SendSMS(code.LoginID, opts)
	default:
		panic("verification: unsupported login ID type: " + code.LoginIDType)
	}
	if err != nil {
		return err
	}

	return nil
}
