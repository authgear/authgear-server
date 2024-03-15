package service

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
)

type VerifyOptions struct {
	OOBChannel        *model.AuthenticatorOOBChannel
	UseSubmittedValue bool
}

type VerifyResult struct {
	Password *password.VerifyResult
	Passkey  bool
}
