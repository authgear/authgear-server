package sso

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
)

type LoginIDNormalizerFactory interface {
	NormalizerWithLoginIDType(loginIDKeyType config.LoginIDKeyType) loginid.Normalizer
}
