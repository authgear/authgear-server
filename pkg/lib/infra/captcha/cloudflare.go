// Legacy cloudfare captcha
package captcha

import (
	"encoding/json"
	"net/http"
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
	HTTPClient  *http.Client
	Credentials *config.CaptchaCloudflareCredentials
}

func NewCloudflareClient(c *config.CaptchaCloudflareCredentials) *CloudflareClient {
	if c == nil {
		return nil
	}
	return &CloudflareClient{
		HTTPClient:  http.DefaultClient,
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
