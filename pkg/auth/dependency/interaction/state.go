package interaction

// State represent the derived current state of interaction.
type State struct {
	Steps []StepState
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
	AvailableAuthenticators []AuthenticatorSpec
}
