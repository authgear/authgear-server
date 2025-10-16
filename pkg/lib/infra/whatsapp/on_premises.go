package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	nethttputil "net/http/httputil"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var WhatsappOnPremisesClientLogger = slogutil.NewLogger("whatsapp-on-premises-client")

type OnPremisesClient struct {
	HTTPClient  HTTPClient
	Endpoint    *url.URL
	Credentials *config.WhatsappOnPremisesCredentials
	TokenStore  *TokenStore
}

func NewWhatsappOnPremisesClient(
	credentials *config.WhatsappOnPremisesCredentials,
	tokenStore *TokenStore,
	httpClient HTTPClient,
) *OnPremisesClient {
	if credentials == nil {
		return nil
	}
	endpoint, err := url.Parse(credentials.APIEndpoint)
	if err != nil {
		panic(err)
	}
	return &OnPremisesClient{
		HTTPClient:  httpClient,
		Endpoint:    endpoint,
		Credentials: credentials,
		TokenStore:  tokenStore,
	}
}

func (c *OnPremisesClient) SendTemplate(
	ctx context.Context,
	to string,
	templateConfig *config.WhatsappOnPremisesOTPTemplateConfig,
	templateLanguage string,
	templateComponents []onPremisesTemplateComponent,
) error {
	token, err := c.TokenStore.Get(ctx, c.Credentials.APIEndpoint, c.Credentials.Username)
	if err != nil {
		return err
	}
	refreshToken := func() error {
		token, err = c.login(ctx)
		if err != nil {
			return err
		}
		return c.TokenStore.Set(ctx, token)
	}
	if token == nil {
		err := refreshToken()
		if err != nil {
			return err
		}
	}
	var send func(retryOnUnauthorized bool) error
	send = func(retryOnUnauthorized bool) error {
		err = c.sendTemplate(ctx, token.Token, to, templateConfig.Name, templateLanguage, templateComponents, templateConfig.Namespace)
		if err != nil {
			if retryOnUnauthorized && errors.Is(err, ErrUnauthorized) {
				err := refreshToken()
				if err != nil {
					return err
				}
				return send(false)
			} else {
				return err
			}
		}
		return nil
	}
	return send(true)
}

func (c *OnPremisesClient) GetOTPTemplate() *config.WhatsappOnPremisesOTPTemplateConfig {
	return &c.Credentials.Templates.OTP
}

func (c *OnPremisesClient) sendTemplate(
	ctx context.Context,
	authToken string,
	to string,
	templateName string,
	templateLanguage string,
	templateComponents []onPremisesTemplateComponent,
	namespace string) error {
	url := c.Endpoint.JoinPath("/v1/messages")
	body := &onPremisesSendTemplateRequest{
		RecipientType: "individual",
		To:            to,
		Type:          "template",
		Template: &onPremisesTemplate{
			Name: templateName,
			Language: &onPremisesTemplateLanguage{
				Policy: "deterministic",
				Code:   templateLanguage,
			},
			Components: templateComponents,
			Namespace:  &namespace,
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// 2xx means success
		// https://developers.facebook.com/docs/whatsapp/on-premises/errors#http
		return nil
	}

	whatsappAPIErr := &WhatsappAPIError{
		HTTPStatusCode: resp.StatusCode,
	}

	dumpedResponse, dumpResponseErr := nethttputil.DumpResponse(resp, true)
	if dumpResponseErr != nil {
		logger := WhatsappOnPremisesClientLogger.GetLogger(ctx)
		logger.WithError(dumpResponseErr).Warn(ctx, "failed to dump response")
	} else {
		whatsappAPIErr.DumpedResponse = dumpedResponse
	}
	// The dump error is not part of the api error, ignore it

	errResp, err := c.tryParseErrorResponse(ctx, resp)
	if err != nil {
		return errors.Join(err, whatsappAPIErr)
	} else {
		whatsappAPIErr.OnPremisesResponse = errResp
	}

	if resp.StatusCode == 401 {
		return errors.Join(ErrUnauthorized, whatsappAPIErr)
	}

	if errResp != nil {
		if firstErrorCode, ok := errResp.FirstErrorCode(); ok {
			switch firstErrorCode {
			case onPremisesErrorCodeInvalidUser:
				return errors.Join(ErrInvalidWhatsappUser, whatsappAPIErr)
			}
		}
	}

	return whatsappAPIErr
}

func (c *OnPremisesClient) tryParseErrorResponse(ctx context.Context, resp *http.Response) (*WhatsappOnPremisesAPIErrorResponse, error) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		// If we failed to read the response body, it is an error
		return nil, err
	}

	var errResp WhatsappOnPremisesAPIErrorResponse
	parseErr := json.Unmarshal(respBody, &errResp)
	// The api could return other errors in format we don't understand, so non-nil parseErr is expected.
	// Just return nil in this case.
	if parseErr != nil {
		logger := WhatsappOnPremisesClientLogger.GetLogger(ctx)
		logger.WithError(parseErr).Warn(ctx, "failed to parse error response")
		return nil, nil
	}
	return &errResp, nil
}

func (c *OnPremisesClient) login(ctx context.Context) (*onPremisesUserToken, error) {
	url := c.Endpoint.JoinPath("/v1/users/login")
	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Credentials.Username, c.Credentials.Password)

	whatsappAPIErr := &WhatsappAPIError{}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Join(err, whatsappAPIErr)
	}
	defer resp.Body.Close()
	whatsappAPIErr.HTTPStatusCode = resp.StatusCode

	dumpedResponse, err := nethttputil.DumpResponse(resp, true)
	if err != nil {
		return nil, errors.Join(err, whatsappAPIErr)
	}
	whatsappAPIErr.DumpedResponse = dumpedResponse

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// It was observed that when status code is 401, the body is empty.
		// Thus parse error should be ignored.
		errResp, parseErr := c.tryParseErrorResponse(ctx, resp)
		if parseErr == nil {
			whatsappAPIErr.OnPremisesResponse = errResp
		}
		return nil, whatsappAPIErr
	}

	loginHTTPResponseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Join(err, whatsappAPIErr)
	}

	var loginResponse onPremisesLoginResponse
	err = json.Unmarshal(loginHTTPResponseBytes, &loginResponse)
	if err != nil {
		return nil, errors.Join(err, whatsappAPIErr)
	}

	if len(loginResponse.Users) < 1 {
		return nil, errors.Join(ErrUnexpectedLoginResponse, whatsappAPIErr)
	}

	return &onPremisesUserToken{
		Endpoint: c.Credentials.APIEndpoint,
		Username: c.Credentials.Username,
		Token:    loginResponse.Users[0].Token,
		ExpireAt: time.Time(loginResponse.Users[0].ExpiresAfter),
	}, nil
}
