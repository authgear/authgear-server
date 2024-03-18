package interaction

type AuthenticatorUpdateReason string

const (
	AuthenticatorUpdateReasonPolicy AuthenticatorUpdateReason = "policy"
	AuthenticatorUpdateReasonExpiry AuthenticatorUpdateReason = "expiry"
)
