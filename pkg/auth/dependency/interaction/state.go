package interaction

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
)

type Step string

const (
	StepAuthenticatePrimary         Step = "authenticate.primary"
	StepAuthenticateSecondary       Step = "authenticate.secondary"
	StepSetupPrimaryAuthenticator   Step = "setup.primary"
	StepSetupSecondaryAuthenticator Step = "setup.secondary"
	StepCommit                      Step = "commit"
	StepOAuth                       Step = "oauth"
)

type StepState struct {
	Step                          Step
	AvailableAuthenticators       []authenticator.Spec
	Identity                      identity.Spec
	OAuthAction                   OAuthAction
	OAuthNonce                    string
	OAuthProviderAuthorizationURL string
	OAuthUserID                   string
}
