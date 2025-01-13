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
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type HTTPClient struct {
	*http.Client
}

func NewHTTPClient() HTTPClient {
	return HTTPClient{
		httputil.NewExternalClient(5 * time.Second),
	}
}

type WhatsappOnPremisesClientLogger struct{ *log.Logger }

func NewWhatsappOnPremisesClientLogger(lf *log.Factory) WhatsappOnPremisesClientLogger {
	return WhatsappOnPremisesClientLogger{lf.New("whatsapp-on-premises-client")}
}

type OnPremisesClient struct {
	Logger      WhatsappOnPremisesClientLogger
	HTTPClient  HTTPClient
	Endpoint    *url.URL
	Credentials *config.WhatsappOnPremisesCredentials
	TokenStore  *TokenStore
}

func NewWhatsappOnPremisesClient(
	lf *log.Factory,
	cfg *config.WhatsappConfig,
	credentials *config.WhatsappOnPremisesCredentials,
	tokenStore *TokenStore,
	httpClient HTTPClient,
) *OnPremisesClient {
	if cfg.APIType != config.WhatsappAPITypeOnPremises || credentials == nil {
		return nil
	}
	endpoint, err := url.Parse(credentials.APIEndpoint)
	if err != nil {
		panic(err)
	}
	return &OnPremisesClient{
		Logger:      WhatsappOnPremisesClientLogger{lf.New("whatsapp-on-premises-client")},
		HTTPClient:  httpClient,
		Endpoint:    endpoint,
		Credentials: credentials,
		TokenStore:  tokenStore,
	}
}

func (c *OnPremisesClient) SendTemplate(
	ctx context.Context,
	to string,
	templateName string,
	templateLanguage string,
	templateComponents []TemplateComponent,
	namespace string) error {
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
		err = c.sendTemplate(ctx, token.Token, to, templateName, templateLanguage, templateComponents, namespace)
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

func (c *OnPremisesClient) GetOTPTemplate() *config.WhatsappTemplateConfig {
	return &c.Credentials.Templates.OTP
}

func (c *OnPremisesClient) sendTemplate(
	ctx context.Context,
	authToken string,
	to string,
	templateName string,
	templateLanguage string,
	templateComponents []TemplateComponent,
	namespace string) error {
	url := c.Endpoint.JoinPath("/v1/messages")
	body := &SendTemplateRequest{
		RecipientType: "individual",
		To:            to,
		Type:          "template",
		Template: &Template{
			Name: templateName,
			Language: &TemplateLanguage{
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
		APIType:        config.WhatsappAPITypeOnPremises,
		HTTPStatusCode: resp.StatusCode,
	}

	dumpedResponse, dumpResponseErr := nethttputil.DumpResponse(resp, true)
	if dumpResponseErr != nil {
		c.Logger.WithError(dumpResponseErr).Warn("failed to dump response")
	} else {
		whatsappAPIErr.DumpedResponse = dumpedResponse
	}
	// The dump error is not part of the api error, ignore it

	errResp, err := c.tryParseErrorResponse(resp)
	if err != nil {
		return errors.Join(err, whatsappAPIErr)
	} else {
		whatsappAPIErr.ParsedResponse = errResp
	}

	if resp.StatusCode == 401 {
		return errors.Join(ErrUnauthorized, whatsappAPIErr)
	}

	if errResp != nil {
		if firstErrorCode, ok := errResp.FirstErrorCode(); ok {
			switch firstErrorCode {
			case errorCodeInvalidUser:
				return errors.Join(ErrInvalidWhatsappUser, whatsappAPIErr)
			}
		}
	}

	return whatsappAPIErr
}

func (c *OnPremisesClient) tryParseErrorResponse(resp *http.Response) (*WhatsappAPIErrorResponse, error) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		// If we failed to read the response body, it is an error
		return nil, err
	}

	var errResp WhatsappAPIErrorResponse
	parseErr := json.Unmarshal(respBody, &errResp)
	// The api could return other errors in format we don't understand, so non-nil parseErr is expected.
	// Just return nil in this case.
	if parseErr != nil {
		c.Logger.WithError(parseErr).Warn("failed to parse error response")
		return nil, nil
	}
	return &errResp, nil
}

func (c *OnPremisesClient) login(ctx context.Context) (*UserToken, error) {
	url := c.Endpoint.JoinPath("/v1/users/login")
	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Credentials.Username, c.Credentials.Password)

	whatsappAPIErr := &WhatsappAPIError{
		APIType: config.WhatsappAPITypeOnPremises,
	}

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
		errResp, parseErr := c.tryParseErrorResponse(resp)
		if parseErr == nil {
			whatsappAPIErr.ParsedResponse = errResp
		}
		return nil, whatsappAPIErr
	}

	loginHTTPResponseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Join(err, whatsappAPIErr)
	}

	var loginResponse LoginResponse
	err = json.Unmarshal(loginHTTPResponseBytes, &loginResponse)
	if err != nil {
		return nil, errors.Join(err, whatsappAPIErr)
	}

	if len(loginResponse.Users) < 1 {
		return nil, errors.Join(ErrUnexpectedLoginResponse, whatsappAPIErr)
	}

	return &UserToken{
		Endpoint: c.Credentials.APIEndpoint,
		Username: c.Credentials.Username,
		Token:    loginResponse.Users[0].Token,
		ExpireAt: time.Time(loginResponse.Users[0].ExpiresAfter),
	}, nil
}
