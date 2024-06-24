package authenticationflow

type BotProtectionVerificationStatus string

const (
	// initial status
	BotProtectionVerificationStatusDefault BotProtectionVerificationStatus = ""

	// terminal status
	BotProtectionVerificationStatusFailed             BotProtectionVerificationStatus = "failed"
	BotProtectionVerificationStatusSuccess            BotProtectionVerificationStatus = "success"
	BotProtectionVerificationStatusServiceUnavailable BotProtectionVerificationStatus = "service_unavailable"
)
