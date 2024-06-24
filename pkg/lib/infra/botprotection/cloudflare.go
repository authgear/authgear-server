package botprotection

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// https://developers.cloudflare.com/turnstile/get-started/server-side-validation/
const (
	CloudflareVerifyEndpoint string = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
)

type CloudflareVerifyResponse struct {
	Success *bool `json:"success,omitempty"`

	// non-empty if Success == false, empty if Success == true
	ErrorCodes []CloudflareVerifyResponseErrorCode `json:"error-codes"`

	// specific to Success == true
	ChallengeTs string `json:"challenge_ts"` // ISO timestamp for the time the challenge was solved.
	Hostname    string `json:"hostname"`     // hostname for which the challenge was served.
	Action      string `json:"action"`       // customer widget identifier passed to the widget on the client side.
	Cdata       string `json:"cdata"`        // customer data passed to the widget on the client side.

}

type CloudflareVerifyResponseErrorCode string

const (
	CloudflareVerifyResponseErrorCodeMissingInputSecret   CloudflareVerifyResponseErrorCode = "missing-input-secret"
	CloudflareVerifyResponseErrorCodeInvalidInputSecret   CloudflareVerifyResponseErrorCode = "invalid-input-secret"
	CloudflareVerifyResponseErrorCodeMissingInputResponse CloudflareVerifyResponseErrorCode = "missing-input-response"
	CloudflareVerifyResponseErrorCodeInvalidInputResponse CloudflareVerifyResponseErrorCode = "invalid-input-response"
	CloudflareVerifyResponseErrorCodeInvalidWidgetId      CloudflareVerifyResponseErrorCode = "invalid-widget-id"
	// nolint: gosec
	CloudflareVerifyResponseErrorCodeInvalidParsedSecret CloudflareVerifyResponseErrorCode = "invalid-parsed-secret"
	CloudflareVerifyResponseErrorCodeBadRequest          CloudflareVerifyResponseErrorCode = "bad-request"
	CloudflareVerifyResponseErrorCodeTimeoutOrDuplicate  CloudflareVerifyResponseErrorCode = "timeout-or-duplicate"
	CloudflareVerifyResponseErrorCodeInternalError       CloudflareVerifyResponseErrorCode = "internal-error"
)

type CloudflareClient struct {
	HTTPClient  *http.Client
	Credentials *config.BotProtectionProviderCredentials
}

func NewCloudflareClient(c *config.BotProtectionProviderCredentials) *CloudflareClient {
	if c == nil {
		return nil
	}
	return &CloudflareClient{
		HTTPClient:  httputil.NewExternalClient(60 * time.Second),
		Credentials: c,
	}
}

func (c *CloudflareClient) Verify(token string, remoteip string) (*CloudflareVerifyResponse, error) {
	formValues := url.Values{}
	formValues.Add("secret", c.Credentials.SecretKey)
	formValues.Add("response", token)

	if remoteip != "" {
		formValues.Add("remoteip", remoteip)
	}

	resp, err := c.HTTPClient.PostForm(CloudflareVerifyEndpoint, formValues)

	if err != nil {
		return nil, err
	}

	respBody := &CloudflareVerifyResponse{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)

	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func (c *CloudflareClient) IsServiceUnavailable(resp *CloudflareVerifyResponse) bool {
	if resp == nil {
		return true
	}

	for _, ec := range resp.ErrorCodes {
		switch ec {
		case CloudflareVerifyResponseErrorCodeInternalError:
			fallthrough
		case CloudflareVerifyResponseErrorCodeTimeoutOrDuplicate:
			return true
		default:
			return false
		}
	}
	// empty error-codes[] slice
	return false
}
