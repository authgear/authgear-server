package otp

type CodeSendResult struct {
	Target       string
	Channel      string
	CodeLength   int
	SendCooldown int
}
