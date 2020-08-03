package otp

type OOBSendResult struct {
	Channel      string
	CodeLength   int
	SendCooldown int
}
