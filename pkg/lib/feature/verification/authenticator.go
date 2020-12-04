package verification

type AuthenticatorStatus string

const (
	AuthenticatorStatusUnverified AuthenticatorStatus = "unverified"
	AuthenticatorStatusVerified   AuthenticatorStatus = "verified"
)
