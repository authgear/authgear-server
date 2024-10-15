package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/opencvfr/openapi"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type OpenCVFRClient interface {
	Get(path string, query url.Values) (respBody []byte, err error)
	Post(path string, body io.Reader, expectedStatus int) (respBody []byte, err error)
	Patch(path string, body io.Reader) (respBody []byte, err error)
	Delete(path string, targetID string) (err error)
}

var _ OpenCVFRClient = &Client{}

type Client struct {
	HTTPClient                *http.Client
	OpenCVFREnvironmentConfig config.OpenCVFREnvironmentConfig
	Endpoint                  string
}

// https://sg.opencv.fr/docs#/
const (
	openCVFREndpoint string = "https://sg.opencv.fr"
	//nolint:gosec // linter mistook this as actual api key and raise security warning
	openCVFRHeaderAPIKey          string = "X-API-Key"
	openCVFRPostDefaultRespStatus int    = http.StatusCreated
)

func NewClient(envCfg *config.EnvironmentConfig) *Client {
	if envCfg == nil {
		panic("openvcv-fr: missing env config")
	}
	return &Client{
		HTTPClient:                httputil.NewExternalClient(60 * time.Second),
		OpenCVFREnvironmentConfig: envCfg.OpenCVFRConfig,
		Endpoint:                  openCVFREndpoint,
	}
}

func (c *Client) Get(path string, query url.Values) (respBody []byte, err error) {
	req, err := http.NewRequest(http.MethodGet, c.Endpoint+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct get request: %w", err)
	}

	c.prepareRequest(req)

	// set query params
	if query != nil {
		req.URL.RawQuery = query.Encode()
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get request: %w", err)
	}
	defer resp.Body.Close()

	respBody, apiErr := c.handleResponse(resp, http.StatusOK)
	if apiErr != nil {
		return nil, apiErr
	}

	return respBody, nil
}

func (c *Client) Post(path string, body io.Reader, expectedStatus int) (respBody []byte, err error) {
	if expectedStatus == 0 {
		expectedStatus = openCVFRPostDefaultRespStatus
	}
	req, err := http.NewRequest(http.MethodPost, c.Endpoint+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to construct post request: %w", err)
	}

	c.prepareRequest(req)
	req.Header.Set("Content-Type", "application/json") // All POST requests in https://sg.opencv.fr/docs#/ have application/json content-type

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute post request: %w", err)
	}
	defer resp.Body.Close()

	respBody, apiErr := c.handleResponse(resp, expectedStatus)
	if apiErr != nil {
		return nil, apiErr
	}

	return respBody, nil
}

func (c *Client) Patch(path string, body io.Reader) (respBody []byte, err error) {
	req, err := http.NewRequest(http.MethodPatch, c.Endpoint+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to construct patch request: %w", err)
	}

	c.prepareRequest(req)
	req.Header.Set("Content-Type", "application/json") // All PATCH requests in https://sg.opencv.fr/docs#/ have application/json content-type

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute patch request: %w", err)
	}
	defer resp.Body.Close()

	respBody, apiErr := c.handleResponse(resp, http.StatusOK)
	if apiErr != nil {
		return nil, apiErr
	}

	return respBody, nil
}

func (c *Client) Delete(path string, targetID string) (err error) {
	url := c.Endpoint + path + "/" + targetID
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to construct delete request: %w", err)
	}

	c.prepareRequest(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute delete request: %w", err)
	}
	defer resp.Body.Close()

	// Note respBody is not used here, since successful DELETE request response body is nil
	_, apiErr := c.handleResponse(resp, http.StatusOK)
	if apiErr != nil {
		return apiErr
	}

	return nil

}

// handleResponse globally handles response from opencv-fr by status code
// if        422, return  nil, OpenCVFRValidationError
// if unexpected, return  nil, OpenCVFRAPIError
// if   expected, return resp, nil
func (c *Client) handleResponse(resp *http.Response, expectedStatus int) (body []byte, err error) {
	// 422
	if resp.StatusCode == http.StatusUnprocessableEntity {
		return nil, c.handle422Response(resp)
	}

	// unexpected
	if resp.StatusCode != expectedStatus {
		return nil, c.handleUnexpectedResponse(resp)
	}

	// expected
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read resp body: %w", err)
	}
	return body, nil
}

// handle422Response handles 422 response from opencv-fr
// it returns OpenCVFRValidationError
func (c *Client) handle422Response(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read resp body: %w", err)
	}
	return &OpenCVFRValidationError{Details: string(body)}
}

// handleUnexpectedResponse handles api-err response from opencv-fr
// it returns OpenCVFRAPIError
func (c *Client) handleUnexpectedResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read resp body: %w", err)
	}
	var parsedBody *openapi.APIErrorResponse
	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		return fmt.Errorf("failed to parse unexpected resp body: %w", err)
	}

	apiErr := &OpenCVFRAPIError{
		HTTPStatusCode:  resp.StatusCode,
		OpenCVFRErrCode: parsedBody.Code,
		Message:         parsedBody.Message,
	}

	retryAfter := resp.Header.Get("retry-after")
	if retryAfter != "" {
		ra, err := strconv.Atoi(retryAfter)
		if err != nil {
			return fmt.Errorf("failed to parse retry-after header: %w", err)
		}
		apiErr.RetryAfter = ra
	}

	return apiErr
}

func (c *Client) prepareRequest(req *http.Request) {
	c.setReqApiKey(req)
}

func (c *Client) setReqApiKey(req *http.Request) {
	req.Header.Set(openCVFRHeaderAPIKey, c.OpenCVFREnvironmentConfig.APIKey)
}
