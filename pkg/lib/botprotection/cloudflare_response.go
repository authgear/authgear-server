package botprotection

import (
	"strings"

	"github.com/authgear/authgear-server/pkg/util/slice"
)

// raw API response from cloudflare turnstile
type CloudflareTurnstileResponse struct {
	Success *bool `json:"success,omitempty,omitempty"`

	// non-empty if Success == false, empty if Success == true
	ErrorCodes []CloudflareTurnstileErrorCode `json:"error-codes,omitempty"`

	// specific to Success == true
	ChallengeTs string `json:"challenge_ts,omitempty"` // ISO timestamp for the time the challenge was solved.
	Hostname    string `json:"hostname,omitempty"`     // hostname for which the challenge was served.
	Action      string `json:"action,omitempty"`       // customer widget identifier passed to the widget on the client side.
	CData       string `json:"cdata,omitempty"`        // customer data passed to the widget on the client side.
}

// returns a comma separated string of error codes
func (e *CloudflareTurnstileResponse) Error() string {
	errorCodeStrings := slice.Map(e.ErrorCodes, func(c CloudflareTurnstileErrorCode) string {
		return string(c)
	})

	return strings.Join(errorCodeStrings, ",")
}

type CloudflareTurnstileErrorCode string

const (
	CloudflareTurnstileErrorCodeMissingInputSecret   CloudflareTurnstileErrorCode = "missing-input-secret"
	CloudflareTurnstileErrorCodeInvalidInputSecret   CloudflareTurnstileErrorCode = "invalid-input-secret"
	CloudflareTurnstileErrorCodeMissingInputResponse CloudflareTurnstileErrorCode = "missing-input-response"
	CloudflareTurnstileErrorCodeInvalidInputResponse CloudflareTurnstileErrorCode = "invalid-input-response"
	CloudflareTurnstileErrorCodeInvalidWidgetId      CloudflareTurnstileErrorCode = "invalid-widget-id"
	// nolint: gosec
	CloudflareTurnstileErrorCodeInvalidParsedSecret CloudflareTurnstileErrorCode = "invalid-parsed-secret"
	CloudflareTurnstileErrorCodeBadRequest          CloudflareTurnstileErrorCode = "bad-request"
	CloudflareTurnstileErrorCodeTimeoutOrDuplicate  CloudflareTurnstileErrorCode = "timeout-or-duplicate"
	CloudflareTurnstileErrorCodeInternalError       CloudflareTurnstileErrorCode = "internal-error"
)

var CloudFlareTurnstileServiceUnavailableErrorCodes = [...]CloudflareTurnstileErrorCode{
	CloudflareTurnstileErrorCodeInternalError,
}
