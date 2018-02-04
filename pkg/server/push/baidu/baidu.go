// Package baidu provides Baidu Push client implementation, and a push.Pusher
// implementation for skygear-server.
//
// Currently only push notification to Android is supported.
package baidu

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/skyversion"
)

var timeNow = func() time.Time { return time.Now().UTC() }

// DeviceType specifies the kind of device to send a notification.
// Currently only Android is supported. This type exists for documentation purpose.
type DeviceType int

// Types of device that can be sent a notification via baidu push. Note that
// DeviceTypeUnspecified is not a valid value in a push request.
const (
	DeviceTypeUnspecified DeviceType = 0
	DeviceTypeAndroid     DeviceType = 3
	DeviceTypeIOS         DeviceType = 4
)

// PushAllDeviceRequest represents a push request to all devices registered
// to an account. It exists for testing purpose and no skygear's API is
// using it.
type PushAllDeviceRequest struct {
	MsgType      int
	Msg          AndroidNotificationMsg
	MsgExpires   int
	DeployStatus int
	SendTime     int
}

// PushSingleDeviceRequest represents a push request to a single device.
// See http://push.baidu.com/doc/restapi/restapi#-post-rest-3-0-push-single_device.
type PushSingleDeviceRequest struct {
	ChannelID    string
	MsgType      int
	Msg          AndroidNotificationMsg
	MsgExpires   int // in seconds
	DeployStatus int // only applicable in iOS
}

// NewPushSingleDeviceRequest returns a NewPushSingleDeviceRequest
func NewPushSingleDeviceRequest(channelID string, Msg AndroidNotificationMsg) PushSingleDeviceRequest {
	return PushSingleDeviceRequest{
		ChannelID: channelID,
		Msg:       Msg,
		MsgType:   1,
	}
}

// AndroidNotificationMsg represents a notification of Android devices.
type AndroidNotificationMsg struct {
	Title                  string                 `json:"title"`
	Description            string                 `json:"description"`
	NotificationBuilderID  int                    `json:"notification_builder_id,omitempty"`
	NotificationBasicStyle int                    `json:"notification_basic_style,omitempty"`
	OpenType               int                    `json:"open_type,omitempty"`
	URL                    string                 `json:"url,omitempty"`
	PkgContent             string                 `json:"pkg_content,omitempty"`
	CustomContent          map[string]interface{} `json:"custom_content,omitempty"`
}

// NewAndroidNotificationMsg returns a new AndroidNotificationMsg.
func NewAndroidNotificationMsg(title, description string) AndroidNotificationMsg {
	return AndroidNotificationMsg{
		Title:       title,
		Description: description,
	}
}

// PushResponse is the response returned from a successful push request (e.g.
// PushAllDeviceRequest & PushSingleDeviceRequest)
type PushResponse struct {
	RequestID      int                `json:"request_id"`
	ResponseParams PushResponseParams `json:"response_params"`
}

// PushResponseParams is the params returned from a successful push request.
type PushResponseParams struct {
	MsgID    string `json:"msg_id"`
	SendTime int    `json:"send_time"`
}

// An Error is an erroneous response returned by baidu push server.
type Error struct {
	// Request ID that leads to this error.
	RequestID int `json:"request_id"`

	// See http://push.baidu.com/doc/restapi/error_code for a complete
	// list of error codes.
	Code int `json:"error_code"`

	// A human readable message about the error in Simplified Chinese.
	Msg string `json:"error_msg"`
}

func (e Error) Error() string {
	return fmt.Sprintf(`%v: %v`, e.Code, e.Msg)
}

// A Client is a baidu push client.
type Client struct {
	httpClient *http.Client
	baseURL    *url.URL
	apiKey     string
	secretKey  string
}

// NewClient returns a new Client given a baseURL, apiKey and secretyKey.
func NewClient(baseURL string, apiKey string, secretKey string) *Client {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}

	return &Client{
		httpClient: &http.Client{},
		baseURL:    parsed,
		apiKey:     apiKey,
		secretKey:  secretKey,
	}
}

func newClientWithURL(baseURL *url.URL, apiKey string, secretKey string) *Client {
	return &Client{
		httpClient: &http.Client{},
		baseURL:    baseURL,
		apiKey:     apiKey,
		secretKey:  secretKey,
	}
}

// PushAllDevice issues a "push/all" request to baidu push server.
func (c *Client) PushAllDevice(req PushAllDeviceRequest) (*PushResponse, error) {
	urlCopy := *c.baseURL
	urlCopy.Path = path.Join(c.baseURL.Path, "push/all")
	url := urlCopy.String()

	timestamp := timeNow().Unix()
	msgBytes, err := json.Marshal(req.Msg)
	if err != nil {
		return nil, fmt.Errorf("Couldn't encode push msg; err = %v", err)
	}

	params := map[string]string{
		"msg_type": strconv.Itoa(req.MsgType),
		"msg":      string(msgBytes),
	}
	if req.MsgExpires != 0 {
		params["msg_expires"] = strconv.Itoa(req.MsgExpires)
	}
	if req.DeployStatus != 0 {
		params["deploy_status"] = strconv.Itoa(req.DeployStatus)
	}
	if req.SendTime != 0 {
		params["send_time"] = strconv.Itoa(req.SendTime)
	}

	pushReq := request{
		method:     "POST",
		url:        url,
		apiKey:     c.apiKey,
		timestamp:  timestamp,
		expires:    0,
		deviceType: DeviceTypeAndroid,
		params:     params,
	}
	pushResp := PushResponse{}
	if err := c.do(&pushReq, &pushResp); err != nil {
		return nil, err
	}

	return &pushResp, nil
}

// PushSingleDevice issues a "push/single_device" request to baidu push server.
func (c *Client) PushSingleDevice(req PushSingleDeviceRequest) (*PushResponse, error) {
	urlCopy := *c.baseURL
	urlCopy.Path = path.Join(c.baseURL.Path, "push/single_device")
	url := urlCopy.String()

	timestamp := timeNow().Unix()
	msgBytes, err := json.Marshal(req.Msg)
	if err != nil {
		return nil, fmt.Errorf("Couldn't encode push msg; err = %v", err)
	}

	params := map[string]string{
		"channel_id": req.ChannelID,
		"msg_type":   strconv.Itoa(req.MsgType),
		"msg":        string(msgBytes),
	}
	if req.MsgExpires != 0 {
		params["msg_expires"] = strconv.Itoa(req.MsgExpires)
	}
	if req.DeployStatus != 0 {
		params["deploy_status"] = strconv.Itoa(req.DeployStatus)
	}

	pushReq := request{
		method:     "POST",
		url:        url,
		apiKey:     c.apiKey,
		timestamp:  timestamp,
		expires:    0,
		deviceType: DeviceTypeAndroid,
		params:     params,
	}
	pushResp := PushResponse{}
	if err := c.do(&pushReq, &pushResp); err != nil {
		return nil, err
	}

	return &pushResp, nil
}

func (c *Client) do(req *request, resp interface{}) error {
	httpReq := httpRequest(*req, c.secretKey)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("push/baidu: couldn't read resp body: err = %v", err)
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode > 299 {
		e := Error{}
		if err := json.Unmarshal(body, &e); err != nil {
			return fmt.Errorf("push/baidu: couldn't decode error resp body: err = %v", err)
		}

		return e
	}

	if err := json.Unmarshal(body, resp); err != nil {
		return fmt.Errorf("push/baidu: couldn't decode resp; err = %v", err)
	}

	return nil
}

// Private struct that fully represents a request to baidu push. It and
// Client.secretKey should completely defines a HTTP request to baidu push.
type request struct {
	method     string
	url        string
	apiKey     string
	timestamp  int64
	expires    int64
	deviceType DeviceType
	params     map[string]string
}

func httpRequest(r request, secretKey string) *http.Request {
	values := urlValues(r)
	signature := signRequest(r.method, r.url, values, secretKey)
	values.Set("sign", signature)

	body := strings.NewReader(values.Encode())
	req, err := http.NewRequest(r.method, r.url, body)
	if err != nil {
		// it must be programmatical error for NewRequest to err because
		// body is strings.Reader, which never err on Read
		panic(fmt.Sprintf("want http.NewRequest succeed, got err = %v", err))
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	req.Header.Set("User-Agent", uaString())

	return req
}

func urlValues(r request) url.Values {
	values := url.Values{}

	// common params
	values.Set("apikey", r.apiKey)
	values.Set("timestamp", strconv.FormatInt(r.timestamp, 10))
	if r.expires > 0 {
		values.Set("expires", strconv.FormatInt(r.expires, 10))
	}
	if r.deviceType != 0 {
		values.Set("device_type", strconv.Itoa(int(r.deviceType)))
	}

	// request-specific params
	for key, value := range r.params {
		values.Add(key, value)
	}

	return values
}

func uaString() string {
	return fmt.Sprintf("BCCS_SDK/3.0 (%v; %v) Go/%v (skygear-server %v)", runtime.GOOS, runtime.GOARCH, runtime.Version(), skyversion.Version())
}
