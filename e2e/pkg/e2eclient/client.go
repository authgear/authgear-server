package e2eclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type Client struct {
	Context       context.Context
	HTTPClient    *http.Client
	OAuthClient   *http.Client
	LocalEndpoint *url.URL
	HTTPHost      httputil.HTTPHost
}

func NewClient(ctx context.Context, mainListenAddr string, httpHost httputil.HTTPHost) *Client {
	// Always use http because we are going to call ourselves locally.
	localEndpointString := fmt.Sprintf("http://%v", mainListenAddr)
	localEndpointURL, err := url.Parse(localEndpointString)
	if err != nil {
		panic(err)
	}

	// Only the port is important, the host is always the loopback address.
	localEndpointURL.Host = fmt.Sprintf("127.0.0.1:%v", localEndpointURL.Port())

	// Prepare HTTP clients.
	var httpClient = &http.Client{}
	var oauthClient = &http.Client{}

	httpClient.Timeout = 60 * time.Second
	oauthClient.Timeout = 60 * time.Second

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

	return &Client{
		Context:       ctx,
		HTTPClient:    httpClient,
		OAuthClient:   oauthClient,
		LocalEndpoint: localEndpointURL,
		HTTPHost:      httpHost,
	}
}

// Create creates a new authentication flow.
func (c *Client) Create(flowReference FlowReference, urlQuery string) (*FlowResponse, error) {
	endpoint := c.LocalEndpoint.JoinPath("/api/v1/authentication_flows")
	endpoint.RawQuery = urlQuery

	req, err := c.makeRequest(nil, endpoint, flowReference)
	if err != nil {
		return nil, err
	}

	return c.doRequest(nil, req)
}

// Get retrieves the flow state.
func (c *Client) Get(stateToken string) (*FlowResponse, error) {
	endpoint := c.LocalEndpoint.JoinPath("/api/v1/authentication_flows/states")

	body := map[string]interface{}{
		"state_token": stateToken,
	}

	req, err := c.makeRequest(nil, endpoint, body)
	if err != nil {
		return nil, err
	}

	return c.doRequest(nil, req)
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

// Input submits the input to the flow.
func (c *Client) Input(w http.ResponseWriter, r *http.Request, stateToken string, input map[string]interface{}) (*FlowResponse, error) {
	endpoint := c.LocalEndpoint.JoinPath("/api/v1/authentication_flows/states/input")

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

	return c.doRequest(w, req)
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

func (c *Client) doRequest(maybeResponseWriter http.ResponseWriter, r *http.Request) (*FlowResponse, error) {
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
