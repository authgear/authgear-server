package service

import "github.com/authgear/authgear-server/pkg/api/model"

type VerifyOptions struct {
	OOBChannel *model.AuthenticatorOOBChannel
}
