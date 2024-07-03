package botprotection

import (
	"encoding/json"
	"errors"
	"fmt"
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

func (c *RecaptchaV2Client) Verify(token string, remoteip string) (*RecaptchaV2SuccessResponse, error) {
	formValues := url.Values{}
	formValues.Add("secret", c.Credentials.SecretKey)
	formValues.Add("response", token)

	if remoteip != "" {
		formValues.Add("remoteip", remoteip)
	}

	resp, err := c.HTTPClient.PostForm(Recaptchav2VerifyEndpoint, formValues)

	if err != nil {
		return nil, errors.Join(err, ErrVerificationFailed)
	}

	respBody := &RecaptchaV2RawResponse{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)

	if err != nil || respBody.Success == nil {
		err := errors.Join(
			fmt.Errorf("unrecognised response body from recaptchav2"),
			err,
			ErrVerificationFailed,
		)
		return nil, err
	}

	if *respBody.Success {
		return &RecaptchaV2SuccessResponse{
			ChallengeTs: respBody.ChallengeTs,
			Hostname:    respBody.Hostname,
		}, nil
	}

	// failed
	failedResp := &RecaptchaV2ErrorResponse{
		ErrorCodes: respBody.ErrorCodes,
	}
	return nil, errors.Join(
		errors.New(failedResp.Error()),
		ErrVerificationFailed,
	)
}
