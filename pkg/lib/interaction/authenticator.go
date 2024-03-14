package interaction

type AuthenticatorUpdateReason string

const (
	AuthenticatorUpdateReasonPolicy            AuthenticatorUpdateReason = "policy"
	AuthenticatorUpdateReasonExpiryForceChange AuthenticatorUpdateReason = "expiry_force_change"
)
