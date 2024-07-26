package service

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
)

type VerifyOptions struct {
	OOBChannel        *model.AuthenticatorOOBChannel
	UseSubmittedValue bool
	Form              otp.Form
}

type VerifyResult struct {
	Password *password.VerifyResult
	Passkey  bool
}
