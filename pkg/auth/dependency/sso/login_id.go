package sso

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type LoginIDNormalizerFactory interface {
	NormalizerWithLoginIDKey(loginIDKey string) loginid.Normalizer
	NormalizerWithLoginIDType(loginIDKeyType config.LoginIDKeyType) loginid.Normalizer
}
