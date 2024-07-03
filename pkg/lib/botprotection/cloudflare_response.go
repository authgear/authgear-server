package botprotection

import (
	"strings"

	"github.com/authgear/authgear-server/pkg/util/slice"
)

// raw API response from cloudflare turnstile, which is not visible to consumer of this API client
type CloudflareTurnstileRawResponse struct {
	Success *bool `json:"success,omitempty"`

	// non-empty if Success == false, empty if Success == true
	ErrorCodes []CloudflareTurnstileErrorCode `json:"error-codes"`

	// specific to Success == true
	ChallengeTs string `json:"challenge_ts"` // ISO timestamp for the time the challenge was solved.
	Hostname    string `json:"hostname"`     // hostname for which the challenge was served.
	Action      string `json:"action"`       // customer widget identifier passed to the widget on the client side.
	CData       string `json:"cdata"`        // customer data passed to the widget on the client side.
}

// parsed success response
type CloudflareTurnstileSuccessResponse struct {
	ChallengeTs string `json:"challenge_ts"` // ISO timestamp for the time the challenge was solved.
	Hostname    string `json:"hostname"`     // hostname for which the challenge was served.
	Action      string `json:"action"`       // customer widget identifier passed to the widget on the client side.
	CData       string `json:"cdata"`        // customer data passed to the widget on the client side.
}

// parsed error response
type CloudflareTurnstileErrorResponse struct {
	ErrorCodes []CloudflareTurnstileErrorCode `json:"error-codes"`
}

// returns a comma separated string of error codes
func (e *CloudflareTurnstileErrorResponse) Error() string {
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
