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

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type HTTPClient struct {
	*http.Client
}

func NewHTTPClient() HTTPClient {
	return HTTPClient{
		httputil.NewExternalClient(5 * time.Second),
	}
}

type OnPremisesClient struct {
	HTTPClient  HTTPClient
	Endpoint    *url.URL
	Credentials *config.WhatsappOnPremisesCredentials
	TokenStore  *TokenStore
}

func NewWhatsappOnPremisesClient(
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

	// Try to read and parse the body, but it is ok if it failed
	respBody, _ := io.ReadAll(resp.Body)
	var errResp *WhatsappAPIErrorResponse
	if len(respBody) > 0 {
		errResp, _ = c.tryParseErrorResponse(respBody)
	}

	var finalErr error = &WhatsappAPIError{
		APIType:          config.WhatsappAPITypeOnPremises,
		HTTPStatusCode:   resp.StatusCode,
		ResponseBodyText: string(respBody),
		ParsedResponse:   errResp,
	}

	if resp.StatusCode == 401 {
		finalErr = errors.Join(ErrUnauthorized, finalErr)
	}

	if errResp.Errors != nil && len(*errResp.Errors) > 0 {
		switch (*errResp.Errors)[0].Code {
		case errorCodeInvalidUser:
			finalErr = errors.Join(ErrInvalidWhatsappUser, finalErr)
		}
	}

	return finalErr
}

func (c *OnPremisesClient) tryParseErrorResponse(body []byte) (*WhatsappAPIErrorResponse, error) {
	var errResp WhatsappAPIErrorResponse
	err := json.Unmarshal(body, &errResp)
	if err == nil {
		return &errResp, nil
	}
	return nil, err
}

func (c *OnPremisesClient) login(ctx context.Context) (*UserToken, error) {
	url := c.Endpoint.JoinPath("/v1/users/login")
	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Credentials.Username, c.Credentials.Password)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("whatsapp: unexpected response status %d", resp.StatusCode)
	}

	loginHTTPResponseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var loginResponse LoginResponse
	err = json.Unmarshal(loginHTTPResponseBytes, &loginResponse)
	if err != nil {
		return nil, err
	}

	if len(loginResponse.Users) < 1 {
		return nil, ErrUnexpectedLoginResponse
	}

	return &UserToken{
		Endpoint: c.Credentials.APIEndpoint,
		Username: c.Credentials.Username,
		Token:    loginResponse.Users[0].Token,
		ExpireAt: time.Time(loginResponse.Users[0].ExpiresAfter),
	}, nil
}
