package authenticationflow

type BotProtectionVerificationResult struct {
	Outcome  BotProtectionVerificationOutcome `json:"outcome,omitempty"`
	Response interface{}                      `json:"response,omitempty"`
}

type BotProtectionVerificationOutcome string

const (
	BotProtectionVerificationOutcomeFailed             = "failed"
	BotProtectionVerificationOutcomeVerified           = "verified"
	BotProtectionVerificationOutcomeServiceUnavailable = "service_unavailable"
)
