package interaction

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
)

type Step string

const (
	StepAuthenticatePrimary         Step = "authenticate.primary"
	StepAuthenticateSecondary       Step = "authenticate.secondary"
	StepSetupPrimaryAuthenticator   Step = "setup.primary"
	StepSetupSecondaryAuthenticator Step = "setup.secondary"
	StepCommit                      Step = "commit"
)

type StepState struct {
	Step                    Step
	AvailableAuthenticators []authenticator.Spec
}
