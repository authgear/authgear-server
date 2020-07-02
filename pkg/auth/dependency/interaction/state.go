package interaction

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
)

// State represent the derived current state of interaction.
// Steps contains the steps that have been executed and the next step to be executed.
// The following invariants hold:
// The last item is the next step to be executed.
// Any other preceding items are steps that have been executed in this interaction.
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
