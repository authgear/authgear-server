package interaction

// State represent the derived current state of interaction.
type State struct {
	RequiredAction          StepAction
	AvailableAuthenticators []AuthenticatorSpec
}

type StepAction string

const (
	StepActionAuthenticatePrimary         StepAction = "authenticate.primary"
	StepActionAuthenticateSecondary       StepAction = "authenticate.secondary"
	StepActionSetupPrimaryAuthenticator   StepAction = "setup.primary"
	StepActionSetupSecondaryAuthenticator StepAction = "setup.secondary"
	StepActionCommit                      StepAction = "commit"
)
