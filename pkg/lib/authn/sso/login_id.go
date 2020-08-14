package sso

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type LoginIDNormalizerFactory interface {
	NormalizerWithLoginIDType(loginIDKeyType config.LoginIDKeyType) loginid.Normalizer
}
