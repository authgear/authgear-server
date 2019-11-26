package sso

import (
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type Provider interface {
	EncodeState(state State) (encodedState string, err error)
	DecodeState(encodedState string) (*State, error)

	EncodeSkygearAuthorizationCode(SkygearAuthorizationCode) (code string, err error)
	DecodeSkygearAuthorizationCode(code string) (*SkygearAuthorizationCode, error)

	IsAllowedOnUserDuplicate(a model.OnUserDuplicate) bool
	IsValidCallbackURL(string) bool

	IsExternalAccessTokenFlowEnabled() bool
}
