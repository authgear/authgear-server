package otp

type CodeSendResult struct {
	Channel      string
	CodeLength   int
	SendCooldown int
}
