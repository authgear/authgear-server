package verification

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
)

type OTPMessageSender interface {
	SendEmail(email string, opts otp.SendOptions) error
	SendSMS(phone string, opts otp.SendOptions) error
}

type CodeSender struct {
	OTPMessageSender OTPMessageSender
}

func (s *CodeSender) SendCode(code *Code) (err error) {
	opts := otp.SendOptions{
		OTP:         code.Code,
		URL:         "", // TODO(verification): Support verification link in future.
		MessageType: otp.MessageTypeVerification,
	}

	switch model.LoginIDKeyType(code.LoginIDType) {
	case model.LoginIDKeyTypeEmail:
		err = s.OTPMessageSender.SendEmail(code.LoginID, opts)
	case model.LoginIDKeyTypePhone:
		err = s.OTPMessageSender.SendSMS(code.LoginID, opts)
	default:
		panic("verification: unsupported login ID type: " + code.LoginIDType)
	}
	if err != nil {
		return
	}

	return
}
