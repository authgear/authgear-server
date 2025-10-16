package e2eclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/pkce"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
)

type ProjectHostRewriteTransport struct {
	HTTPHost     httputil.HTTPHost
	MainEndpoint *url.URL
}

func (t *ProjectHostRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == string(t.HTTPHost) {
		// Send the request to main endpoint if it is to the project specific host
		req.Host = string(t.HTTPHost)
		req.URL.Host = t.MainEndpoint.Host
	}

	if req.Host == string(t.MainEndpoint.Host) {
		// Set the host header so that authgear server can map the request to the correct project
		req.Host = string(t.HTTPHost)
	}

	return http.DefaultTransport.RoundTrip(req)
}

type Client struct {
	Context          context.Context
	CookieJar        http.CookieJar
	HTTPClient       *http.Client
	NoRedirectClient *http.Client
	OAuthClient      *http.Client
	MainEndpoint     *url.URL
	AdminEndpoint    *url.URL
	HTTPHost         httputil.HTTPHost

	SAMLClient *SAMLClient
}

func NewClient(ctx context.Context, mainListenAddr string, adminListenAddr string, httpHost httputil.HTTPHost) *Client {
	// Always use http because we are going to call ourselves locally.
	mainEndpointString := fmt.Sprintf("http://%v", mainListenAddr)
	mainEndpointURL, err := url.Parse(mainEndpointString)
	if err != nil {
		panic(err)
	}

	adminEndpointString := fmt.Sprintf("http://%v", adminListenAddr)
	adminEndpointURL, err := url.Parse(adminEndpointString)
	if err != nil {
		panic(err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	customJar := &JarWorkingAroundGolangIssue38988{
		Jar:           jar,
		CorrectedHost: string(httpHost),
	}
	var transport = &ProjectHostRewriteTransport{
		MainEndpoint: mainEndpointURL,
		HTTPHost:     httpHost,
	}

	var httpClient = &http.Client{
		Jar:       customJar,
		Transport: transport,
	}
	var noRedirectClient = &http.Client{
		Jar: customJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: transport,
	}
	var oauthClient = &http.Client{}

	// Use go test -timeout instead of setting timeout here.
	httpClient.Timeout = 0
	oauthClient.Timeout = 0

	// Intercept HTTP requests to the OAuth server.
	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		panic(err)
	}
	caCert, err := os.ReadFile("../../ssl/ca.crt")
	if err != nil {
		panic(err)
	}
	caCertPool.AppendCertsFromPEM(caCert)

	proxyUrl, err := url.Parse("http://localhost:8080")
	if err != nil {
		panic(err)
	}

	oauthClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			// TLS 1.2 is minimum version by default
			MinVersion: tls.VersionTLS12,
			RootCAs:    caCertPool,
		},
		Proxy: http.ProxyURL(proxyUrl),
	}

	// Disable redirect following to extract OAuth callback code.
	oauthClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	samlClient := &SAMLClient{
		Context:    ctx,
		HTTPClient: noRedirectClient,
		HTTPHost:   httpHost,
	}

	return &Client{
		Context:          ctx,
		CookieJar:        customJar,
		HTTPClient:       httpClient,
		NoRedirectClient: noRedirectClient,
		OAuthClient:      oauthClient,
		MainEndpoint:     mainEndpointURL,
		AdminEndpoint:    adminEndpointURL,
		HTTPHost:         httpHost,
		SAMLClient:       samlClient,
	}
}

// CreateFlow creates a new authentication flow.
func (c *Client) CreateFlow(input map[string]any) (*FlowResponse, error) {
	endpoint := c.MainEndpoint.JoinPath("/api/v1/authentication_flows")

	req, err := c.makeRequest(nil, endpoint, input)
	if err != nil {
		return nil, err
	}

	return c.doFlowRequest(nil, req)
}

// GetFlowState retrieves the flow state.
func (c *Client) GetFlowState(stateToken string) (*FlowResponse, error) {
	endpoint := c.MainEndpoint.JoinPath("/api/v1/authentication_flows/states")

	body := map[string]interface{}{
		"state_token": stateToken,
	}

	req, err := c.makeRequest(nil, endpoint, body)
	if err != nil {
		return nil, err
	}

	return c.doFlowRequest(nil, req)
}

// GenerateTOTPCode generates a TOTP code for the given secret.
func (c *Client) GenerateTOTPCode(secret string) (string, error) {
	totp, err := secretcode.NewTOTPFromSecret(secret)
	if err != nil {
		return "", err
	}

	code, err := totp.GenerateCode(time.Now().UTC())
	if err != nil {
		return "", err
	}

	return code, nil
}

// OAuthRedirect follows the OAuth redirect until the URL matches the given prefix.
func (c *Client) OAuthRedirect(url string, redirectUntil string) (finalURL string, err error) {
	for {
		req, err := http.NewRequestWithContext(c.Context, "GET", url, nil)
		if err != nil {
			return "", err
		}

		resp, err := c.OAuthClient.Do(req)
		if err != nil {
			return "", err
		}

		if resp.StatusCode != http.StatusFound {
			return "", fmt.Errorf("unexpected status code at %s: %v", req.URL.String(), resp.StatusCode)
		}

		location := resp.Header.Get("Location")
		if strings.HasPrefix(location, redirectUntil) {
			return location, nil
		}

		url = location
	}
}

func (c *Client) SetupOAuth() (output map[string]any, err error) {
	u := c.MainEndpoint.JoinPath("/oauth2/authorize")

	values := make(url.Values)
	values.Set("client_id", "e2e")
	values.Set("redirect_uri", "http://localhost:4000")
	values.Set("response_type", "code")
	values.Set("code_challenge_method", "S256")

	codeVerifier := pkce.GenerateS256Verifier()
	values.Set("code_challenge", codeVerifier.Challenge())
	values.Set("scope", strings.Join([]string{
		"openid",
		"offline_access",
		"https://authgear.com/scopes/full-access",
	}, " "))
	values.Set("x_sso_enabled", "false")

	u.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(c.Context, "GET", u.String(), nil)
	if err != nil {
		return
	}

	resp, err := c.NoRedirectClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")

	redirectURI, err := url.Parse(location)
	if err != nil {
		return
	}

	output = map[string]any{
		"query":         redirectURI.RawQuery,
		"code_verifier": codeVerifier.CodeVerifier,
	}
	return
}

type OAuthExchangeCodeOptions struct {
	CodeVerifier string
	RedirectURI  string
}

type OAuthExchangeCodeResult struct {
	IDToken map[string]any `json:"id_token"`
}

func (c *Client) OAuthExchangeCode(opts OAuthExchangeCodeOptions) (result *OAuthExchangeCodeResult, err error) {
	// We first need to visit the RedirectURI and extract the authorization code.

	req, err := http.NewRequestWithContext(c.Context, "GET", opts.RedirectURI, nil)
	if err != nil {
		return
	}

	resp, err := c.NoRedirectClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	redirectURI, err := url.Parse(resp.Header.Get("Location"))
	if err != nil {
		return
	}

	code := redirectURI.Query().Get("code")

	u := c.MainEndpoint.JoinPath("/oauth2/token")

	values := make(url.Values)
	values.Set("grant_type", "authorization_code")
	values.Set("client_id", "e2e")
	values.Set("code", code)
	values.Set("redirect_uri", "http://localhost:4000")
	values.Set("code_verifier", opts.CodeVerifier)

	tokenReq, err := http.NewRequestWithContext(
		c.Context,
		"POST",
		u.String(),
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenResp, err := c.NoRedirectClient.Do(tokenReq)
	if err != nil {
		return
	}
	defer tokenResp.Body.Close()

	var tokenRespBody map[string]any
	err = json.NewDecoder(tokenResp.Body).Decode(&tokenRespBody)
	if err != nil {
		return
	}

	idTokenStr := tokenRespBody["id_token"].(string)

	idToken, err := jwt.ParseInsecure([]byte(idTokenStr))
	if err != nil {
		return
	}

	idTokenMap, err := idToken.AsMap(c.Context)
	if err != nil {
		return
	}

	result = &OAuthExchangeCodeResult{
		IDToken: idTokenMap,
	}
	return
}

// InputFlow submits the input to the flow.
func (c *Client) InputFlow(w http.ResponseWriter, r *http.Request, stateToken string, input map[string]interface{}) (*FlowResponse, error) {
	endpoint := c.MainEndpoint.JoinPath("/api/v1/authentication_flows/states/input")

	body := map[string]interface{}{
		"input": input,
	}
	if stateToken != "" {
		body["state_token"] = stateToken
	}

	req, err := c.makeRequest(r, endpoint, body)
	if err != nil {
		return nil, err
	}

	return c.doFlowRequest(w, req)
}

func (c *Client) InjectSession(idpSessionID string, idpSessionToken string) {
	encodedToken := idpsession.E2EEncodeToken(idpSessionID, idpSessionToken)
	urlWithProjectHost := c.MainEndpoint.JoinPath("")
	urlWithProjectHost.Host = string(c.HTTPHost)
	c.CookieJar.SetCookies(urlWithProjectHost, []*http.Cookie{
		{Name: "session", Value: encodedToken},
	})
}

func (c *Client) SendSAMLRequest(
	path string,
	samlElementName string,
	samlElementXML string,
	binding SAMLBinding,
	relayState string,
	fn func(r *http.Response) error) error {
	destination := c.MainEndpoint.JoinPath(path)
	switch binding {
	case SAMLBindingHTTPPost:
		return c.SAMLClient.SendSAMLRequestWithHTTPPost(
			samlElementName,
			samlElementXML,
			destination,
			relayState,
			fn,
		)
	case SAMLBindingHTTPRedirect:
		return c.SAMLClient.SendSAMLRequestWithHTTPRedirect(
			samlElementName,
			samlElementXML,
			destination,
			relayState,
			fn,
		)
	default:
		return fmt.Errorf("unknown saml binding %s", binding)
	}
}

func (c *Client) MakeHTTPRequest(
	method string,
	toURL string,
	headers map[string]string,
	body string,
	fn func(r *http.Response) error) error {
	var buf *bytes.Buffer = bytes.NewBuffer([]byte(body))

	if body != "" {
		buf = bytes.NewBuffer([]byte(body))
	}

	req, err := http.NewRequestWithContext(c.Context, method, toURL, buf)
	if err != nil {
		return err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	req.Header.Set("Content-Length", strconv.Itoa(buf.Len()))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return fn(resp)
}

func (c *Client) makeRequest(maybeOriginalRequest *http.Request, endpoint *url.URL, body interface{}) (*http.Request, error) {
	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(c.Context, "POST", endpoint.String(), &buf)
	if err != nil {
		return nil, err
	}

	req.Host = string(c.HTTPHost)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(buf.Len()))

	if maybeOriginalRequest != nil {
		for _, c := range maybeOriginalRequest.Cookies() {
			req.AddCookie(c)
		}
	}

	return req, nil
}

func (c *Client) doFlowRequest(maybeResponseWriter http.ResponseWriter, r *http.Request) (*FlowResponse, error) {
	resp, err := c.HTTPClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Forward cookies.
	if maybeResponseWriter != nil {
		for _, c := range resp.Cookies() {
			httputil.UpdateCookie(maybeResponseWriter, c)
		}
	}

	var httpResponse HTTPResponse
	err = json.NewDecoder(resp.Body).Decode(&httpResponse)
	if err != nil {
		return nil, err
	}

	if httpResponse.Error != nil {
		return nil, httpResponse.Error
	}

	return httpResponse.Result, nil
}

type GraphQLAPIRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Extensions struct {
			Reason string `json:"reason"`
		} `json:"extensions"`
	} `json:"errors"`
}

// In e2e test, recommend to use `GraphQLAPIRaw` below to check the response in JSON string format over a GraphQLResponse Object
func (c *Client) GraphQLAPI(body GraphQLAPIRequest) (*GraphQLResponse, error) {
	endpoint := c.AdminEndpoint.JoinPath("/_api/admin/graphql")

	req, err := c.makeRequest(nil, endpoint, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var graphQLResponse GraphQLResponse
	err = json.NewDecoder(resp.Body).Decode(&graphQLResponse)
	if err != nil {
		return nil, err
	}

	return &graphQLResponse, nil
}

func (c *Client) GraphQLAPIRaw(body GraphQLAPIRequest) (string, error) {
	endpoint := c.AdminEndpoint.JoinPath("/_api/admin/graphql")

	req, err := c.makeRequest(nil, endpoint, body)
	if err != nil {
		return "", err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

type UserImportRequest struct {
	JSONDocument string
}

type UserImportResponseResult struct {
	ID          string                     `json:"id"`
	CreatedAt   *time.Time                 `json:"created_at"`
	CompletedAt *time.Time                 `json:"completed_at,omitzero"`
	Status      string                     `json:"status"`
	Summary     *UserImportResponseSummary `json:"summary,omitzero"`
	Details     []UserImportResponseDetail `json:"details,omitzero"`
}

type UserImportResponseSummary struct {
	Total    int `json:"total"`
	Inserted int `json:"inserted"`
	Updated  int `json:"updated"`
	Skipped  int `json:"skipped"`
	Failed   int `json:"failed"`
}

type UserImportResponseDetail struct {
	Index    int                         `json:"index"`
	Outcome  string                      `json:"outcome"`
	UserID   string                      `json:"user_id,omitzero"`
	Record   map[string]interface{}      `json:"record"`
	Warnings []UserImportResponseWarning `json:"warnings,omitzero"`
	Errors   []*apierrors.APIError       `json:"errors,omitzero"`
}

type UserImportResponseWarning struct {
	Message string `json:"message"`
}

type UserImportResponse struct {
	Result *UserImportResponseResult `json:"result,omitzero"`
	Error  *apierrors.APIError       `json:"error,omitempty"`
}

func (c *Client) CreateUserImport(body UserImportRequest) (*UserImportResponseResult, error) {
	endpoint := c.AdminEndpoint.JoinPath("/_api/admin/users/import")

	req, err := c.makeRequest(nil, endpoint, json.RawMessage(body.JSONDocument))
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userImportResponse UserImportResponse
	err = json.NewDecoder(resp.Body).Decode(&userImportResponse)
	if err != nil {
		return nil, err
	}

	if userImportResponse.Error != nil {
		return nil, userImportResponse.Error
	}

	return userImportResponse.Result, nil
}

func (c *Client) GetUserImport(id string) (*UserImportResponseResult, error) {
	endpoint := c.AdminEndpoint.JoinPath("/_api/admin/users/import").JoinPath(id)

	req, err := http.NewRequestWithContext(c.Context, "GET", endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Host = string(c.HTTPHost)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userImportResponse UserImportResponse
	err = json.NewDecoder(resp.Body).Decode(&userImportResponse)
	if err != nil {
		return nil, err
	}

	if userImportResponse.Error != nil {
		return nil, userImportResponse.Error
	}

	return userImportResponse.Result, nil
}
