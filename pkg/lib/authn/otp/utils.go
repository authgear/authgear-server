package otp

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

func selectByChannel[T any](channel model.AuthenticatorOOBChannel, email T, sms T, whatsapp T) T {
	switch channel {
	case model.AuthenticatorOOBChannelEmail:
		return email
	case model.AuthenticatorOOBChannelSMS:
		return sms
	case model.AuthenticatorOOBChannelWhatsapp:
		return whatsapp
	}
	panic("invalid channel: " + channel)
}
