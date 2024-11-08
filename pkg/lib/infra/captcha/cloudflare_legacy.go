// Legacy cloudfare captcha
package captcha

import (
	"encoding/json"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

const (
	CloudflareVerifyEndpoint string = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
)

type CloudflareVerificationResponse struct {
	Success bool `json:"success"`
}

type CloudflareClient struct {
	HTTPClient  HTTPClient
	Credentials *config.Deprecated_CaptchaCloudflareCredentials
}

func NewCloudflareClient(c *config.Deprecated_CaptchaCloudflareCredentials, httpClient HTTPClient) *CloudflareClient {
	if c == nil {
		return nil
	}
	return &CloudflareClient{
		HTTPClient:  httpClient,
		Credentials: c,
	}
}

func (c *CloudflareClient) Verify(token string, remoteip string) (*CloudflareVerificationResponse, error) {
	formValues := url.Values{}
	formValues.Add("secret", c.Credentials.Secret)
	formValues.Add("response", token)

	if remoteip != "" {
		formValues.Add("remoteip", remoteip)
	}

	resp, err := c.HTTPClient.PostForm(CloudflareVerifyEndpoint, formValues)

	if err != nil {
		return nil, err
	}

	respBody := &CloudflareVerificationResponse{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)

	if err != nil {
		return nil, err
	}

	return respBody, nil
}
