package interaction

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator"
)

// State represent the derived current state of interaction.
type State struct {
	Steps []StepState
}

func (s *State) CurrentStep() *StepState {
	if len(s.Steps) == 0 {
		panic("interaction: attempted to get current step when empty")
	}
	return &s.Steps[len(s.Steps)-1]
}

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
