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

type RecaptchaV2Client struct {
	HTTPClient     *http.Client
	Credentials    *config.BotProtectionProviderCredentials
	VerifyEndpoint string
}

func NewRecaptchaV2Client(c *config.BotProtectionProviderCredentials, e *config.EnvironmentConfig) *RecaptchaV2Client {
	if c == nil {
		return nil
	}
	return &RecaptchaV2Client{
		HTTPClient:     httputil.NewExternalClient(60 * time.Second),
		VerifyEndpoint: e.BotProtectionConfig.RecaptchaV2Endpoint,
		Credentials:    c,
	}
}

func (c *RecaptchaV2Client) Verify(token string, remoteip string) (*RecaptchaV2Response, error) {
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

	respBody := &RecaptchaV2Response{}
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

	return nil, errors.Join(ErrVerificationFailed, respBody)
}
