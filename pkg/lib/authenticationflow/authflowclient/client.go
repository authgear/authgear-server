package authflowclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

var httpClient = &http.Client{}

func init() {
	httpClient.Timeout = 30 * time.Second
}

type Client struct {
	Context       context.Context
	HTTPClient    *http.Client
	LocalEndpoint *url.URL
	HTTPHost      httputil.HTTPHost
}

func NewClient(ctx context.Context, mainListenAddr config.MainListenAddr, httpHost httputil.HTTPHost) *Client {
	// Always use http because we are going to call ourselves locally.
	localEndpointString := fmt.Sprintf("http://%v", mainListenAddr)
	localEndpointURL, err := url.Parse(localEndpointString)
	if err != nil {
		panic(err)
	}

	// Only the port is important, the host is always the loopback address.
	localEndpointURL.Host = fmt.Sprintf("127.0.0.1:%v", localEndpointURL.Port())

	return &Client{
		Context:       ctx,
		HTTPClient:    httpClient,
		LocalEndpoint: localEndpointURL,
		HTTPHost:      httpHost,
	}
}

func (c *Client) Create(flowReference FlowReference, urlQuery string) (*FlowResponse, error) {
	endpoint := c.LocalEndpoint.JoinPath("/api/v1/authentication_flows")
	endpoint.RawQuery = urlQuery

	req, err := c.makeRequest(nil, endpoint, flowReference)
	if err != nil {
		return nil, err
	}

	return c.doRequest(nil, req)
}

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
