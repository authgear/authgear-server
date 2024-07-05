package botprotection

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// https://developers.cloudflare.com/turnstile/get-started/server-side-validation/
const (
	CloudflareTurnstileVerifyEndpoint string = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
)

type CloudflareClient struct {
	HTTPClient     *http.Client
	Credentials    *config.BotProtectionProviderCredentials
	VerifyEndpoint string
}

func NewCloudflareClient(c *config.BotProtectionProviderCredentials, e *config.EnvironmentConfig) *CloudflareClient {
	if c == nil {
		return nil
	}
	verifyEndpoint := CloudflareTurnstileVerifyEndpoint
	if e.End2EndBotProtection.CloudflareEndpoint != "" {
		verifyEndpoint = e.End2EndBotProtection.CloudflareEndpoint
	}
	return &CloudflareClient{
		HTTPClient:     httputil.NewExternalClient(60 * time.Second),
		Credentials:    c,
		VerifyEndpoint: verifyEndpoint,
	}
}

func (c *CloudflareClient) Verify(token string, remoteip string) (*CloudflareTurnstileResponse, error) {
	formValues := url.Values{}
	formValues.Add("secret", c.Credentials.SecretKey)
	formValues.Add("response", token)

	if remoteip != "" {
		formValues.Add("remoteip", remoteip)
	}

	resp, err := c.HTTPClient.PostForm(c.VerifyEndpoint, formValues)

	if err != nil {
		return nil, errors.Join(ErrVerificationServiceUnavailable, err)
	}
	defer resp.Body.Close()

	httpBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Join(ErrVerificationServiceUnavailable, fmt.Errorf("failed to read response body: %w", err))
	}

	respBody := &CloudflareTurnstileResponse{}
	err = json.Unmarshal(httpBodyBytes, &respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err) // internal server error
	}

	if respBody.Success == nil {
		return nil, fmt.Errorf("unexpected response body: %v", string(httpBodyBytes)) // internal server error
	}

	if *respBody.Success {
		return respBody, nil
	}

	// failed
	if len(respBody.ErrorCodes) == 0 {
		return nil, ErrVerificationFailed // fail without error codes if empty
	}

	for _, suErrCode := range CloudFlareTurnstileServiceUnavailableErrorCodes {
		for _, errCode := range respBody.ErrorCodes {
			if errCode == suErrCode {
				return nil, errors.Join(ErrVerificationServiceUnavailable, respBody)
			}
		}
	}

	return nil, errors.Join(ErrVerificationFailed, respBody)
}
