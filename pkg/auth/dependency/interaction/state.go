package interaction

// State represent the derived current state of interaction.
type State struct {
	RequiredAction          StepAction
	AvailableAuthenticators []AuthenticatorSpec
}

type StepAction string

const (
	StepActionAuthenticatePrimary   StepAction = "authenticate.primary"
	StepActionAuthenticateSecondary StepAction = "authenticate.secondary"
	StepActionSetupAuthenticator    StepAction = "setup-authenticator"
	StepActionCompleted             StepAction = "completed"
)
