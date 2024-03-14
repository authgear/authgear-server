package declarative

type PasswordChangeReason string

const (
	PasswordChangeReasonPolicy PasswordChangeReason = "policy"
	// nolint: gosec // This is not a credential
	PasswordChangeReasonExpiryForceChange PasswordChangeReason = "expiry_force_change"
)
