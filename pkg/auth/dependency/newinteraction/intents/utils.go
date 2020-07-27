package intents

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func needPrimaryAuthn(t authn.IdentityType) bool {
	switch t {
	case authn.IdentityTypeLoginID:
		return true
	case authn.IdentityTypeAnonymous, authn.IdentityTypeOAuth:
		return false
	default:
		panic("interaction: unknown identity type" + t)
	}
}

func firstAuthenticationStage(t authn.IdentityType) newinteraction.AuthenticationStage {
	if needPrimaryAuthn(t) {
		return newinteraction.AuthenticationStagePrimary
	}
	return newinteraction.AuthenticationStageSecondary
}
