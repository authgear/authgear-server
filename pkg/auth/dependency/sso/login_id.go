package sso

import (
	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
)

type LoginIDNormalizerFactory interface {
	NormalizerWithLoginIDKey(loginIDKey string) loginid.Normalizer
	NormalizerWithLoginIDType(loginIDKeyType config.LoginIDKeyType) loginid.Normalizer
}
