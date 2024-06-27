package botprotection

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// https://developers.google.com/recaptcha/docs/verify
const (
	Recaptchav2VerifyEndpoint string = "https://www.google.com/recaptcha/api/siteverify"
)

type Recaptchav2VerifyResponse struct {
	Success *bool `json:"success,omitempty"`

	// specific to Success == false
	ErrorCodes []Recaptchav2VerifyResponseErrorCode `json:"error-codes"`

	// specific to Success == true
	ChallengeTs string `json:"challenge_ts"` // timestamp of the challenge load (ISO format yyyy-MM-dd'T'HH:mm:ssZZ)
	Hostname    string `json:"hostname"`     // hostname for which the challenge was served.
}

type Recaptchav2VerifyResponseErrorCode string

const (
	Recaptchav2VerifyResponseErrorCodeMissingInputSecret   Recaptchav2VerifyResponseErrorCode = "missing-input-secret"
	Recaptchav2VerifyResponseErrorCodeInvalidInputSecret   Recaptchav2VerifyResponseErrorCode = "invalid-input-secret"
	Recaptchav2VerifyResponseErrorCodeMissingInputResponse Recaptchav2VerifyResponseErrorCode = "missing-input-response"
	Recaptchav2VerifyResponseErrorCodeInvalidInputResponse Recaptchav2VerifyResponseErrorCode = "invalid-input-response"
	Recaptchav2VerifyResponseErrorCodeBadRequest           Recaptchav2VerifyResponseErrorCode = "bad-request"
	Recaptchav2VerifyResponseErrorCodeTimeoutOrDuplicate   Recaptchav2VerifyResponseErrorCode = "timeout-or-duplicate"
)

type RecaptchaV2Client struct {
	HTTPClient  *http.Client
	Credentials *config.BotProtectionProviderCredentials
}

func NewRecaptchaV2Client(c *config.BotProtectionProviderCredentials) *RecaptchaV2Client {
	if c == nil {
		return nil
	}
	return &RecaptchaV2Client{
		HTTPClient:  httputil.NewExternalClient(60 * time.Second),
		Credentials: c,
	}
}

func (c *RecaptchaV2Client) Verify(token string, remoteip string) (*Recaptchav2VerifyResponse, error) {
	formValues := url.Values{}
	formValues.Add("secret", c.Credentials.SecretKey)
	formValues.Add("response", token)

	if remoteip != "" {
		formValues.Add("remoteip", remoteip)
	}

	resp, err := c.HTTPClient.PostForm(Recaptchav2VerifyEndpoint, formValues)

	if err != nil {
		return nil, err
	}

	respBody := &Recaptchav2VerifyResponse{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)

	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func (c *RecaptchaV2Client) IsServiceUnavailable(resp *Recaptchav2VerifyResponse) bool {
	if resp == nil {
		return true
	}

	for _, ec := range resp.ErrorCodes {
		switch ec {
		case Recaptchav2VerifyResponseErrorCodeTimeoutOrDuplicate:
			return true
		default:
			return false
		}
	}
	// empty error-codes[] slice
	return false
}
