package duration

import (
	"time"
)

const (
	// UserInteraction is a duration that normal user interaction should finish within.
	// https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html#step-3-send-a-token-over-a-side-channel
	UserInteraction = 20 * time.Minute
	// Short is a duration for short living things.
	Short = 5 * time.Minute
	// PerMinute is 1 minute.
	PerMinute = 1 * time.Minute
	// ClockSkew is the duration of acceptable clock skew.
	ClockSkew = 5 * time.Minute
)
