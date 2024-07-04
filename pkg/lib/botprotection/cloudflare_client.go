package botprotection

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// https://developers.cloudflare.com/turnstile/get-started/server-side-validation/

type CloudflareClient struct {
	HTTPClient     *http.Client
	Credentials    *config.BotProtectionProviderCredentials
	VerifyEndpoint string
}

func NewCloudflareClient(c *config.BotProtectionProviderCredentials, e *config.EnvironmentConfig) *CloudflareClient {
	if c == nil {
		return nil
	}
	return &CloudflareClient{
		HTTPClient:     httputil.NewExternalClient(60 * time.Second),
		Credentials:    c,
		VerifyEndpoint: e.BotProtectionConfig.CloudflareEndpoint,
	}
}

func (c *CloudflareClient) Verify(token string, remoteip string) (*CloudflareTurnstileSuccessResponse, error) {
	formValues := url.Values{}
	formValues.Add("secret", c.Credentials.SecretKey)
	formValues.Add("response", token)

	if remoteip != "" {
		formValues.Add("remoteip", remoteip)
	}

	resp, err := c.HTTPClient.PostForm(c.VerifyEndpoint, formValues)

	if err != nil {
		return nil, errors.Join(err, ErrVerificationFailed)
	}

	respBody := &CloudflareTurnstileRawResponse{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)

	if err != nil || respBody.Success == nil {
		err := errors.Join(
			fmt.Errorf("unrecognised response body from cloudflare turnstile"),
			err,
			ErrVerificationFailed,
		)
		return nil, errors.Join(err, ErrVerificationFailed)
	}

	if *respBody.Success {
		return &CloudflareTurnstileSuccessResponse{
			ChallengeTs: respBody.ChallengeTs,
			Hostname:    respBody.Hostname,
			Action:      respBody.Action,
			CData:       respBody.CData,
		}, nil
	}

	// failed
	failedResp := &CloudflareTurnstileErrorResponse{
		ErrorCodes: respBody.ErrorCodes,
	}
	err = errors.New(failedResp.Error())
	for _, errCode := range CloudFlareTurnstileServiceUnavailableErrorCodes {
		if strings.Contains(failedResp.Error(), string(errCode)) {
			return nil, errors.Join(ErrVerificationServiceUnavailable, err)
		}
	}

	return nil, errors.Join(err, ErrVerificationFailed)
}
