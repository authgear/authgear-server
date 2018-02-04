package baidu

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func makeSignRequestFunc(sign string) func(method string, urlString string, values url.Values, secretKey string) string {
	return func(string, string, url.Values, string) string {
		return sign
	}
}

func newClient(httpClient *http.Client, baseURL string, apiKey string, secretKey string) *Client {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    parsed,
		apiKey:     apiKey,
		secretKey:  secretKey,
	}
}

// test taken from http://push.baidu.com/issue/view/1068
func TestSignRequestFunc(t *testing.T) {
	values := urlValuesFromMap(map[string]string{
		"apikey":      "tf9G7QicWAh2aEGFlrI1Ohyb",
		"timestamp":   "1447401600",
		"device_type": "3",
		"msg":         `{"title":"hello","description":"test"}`,
		"msg_expires": "3600",
		"msg_type":    "1",
	})

	signed := signRequestFunc("POST", "http://api.tuisong.baidu.com/rest/3.0/push/all", values, "qQ1h43yGQYG6H6mYgSPizDuCmtbMj6Ld")
	if signed != "f462b3fdac8b92d28bc52fc47c54f56e" {
		t.Errorf(`want signed == "f462b3fdac8b92d28bc52fc47c54f56e", got %v`, signed)
	}
}

func TestPushSingleDevice(t *testing.T) {
	timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
	signRequest = makeSignRequestFunc("1234567890")
	defer func() {
		timeNow = func() time.Time { return time.Now().UTC() }
		signRequest = signRequestFunc
	}()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/push/single_device"
		if r.URL.Path != expectedPath {
			t.Errorf("got path = %v; want %v", r.URL.Path, expectedPath)
		}

		expectedBody := urlValuesFromMap(map[string]string{
			"apikey":      "fake_api_key",
			"channel_id":  "fake_channel_id",
			"device_type": "3",
			"msg":         `{"title":"Title","description":"Description"}`,
			"msg_type":    "1",
			"sign":        "1234567890",
			"timestamp":   "1136214245",
		}).Encode()
		body := string(mustReadAll(r.Body))
		if body != expectedBody {
			t.Errorf("got unexpected request body = %v; want %v", body, expectedBody)
		}

		fmt.Fprintln(w, `{"request_id":123456789,"response_params":{"msg_id":"24234532453245","send_time":1427174155}}`)
	}))
	defer ts.Close()

	req := NewPushSingleDeviceRequest("fake_channel_id", NewAndroidNotificationMsg("Title", "Description"))
	client := newClient(ts.Client(), ts.URL, "fake_api_key", "fake_secret_key")
	resp, err := client.PushSingleDevice(req)
	if err != nil {
		t.Errorf("want err == nil, got %v", err)
	}

	expectedResp := PushResponse{
		RequestID: 123456789,
		ResponseParams: PushResponseParams{
			MsgID:    "24234532453245",
			SendTime: 1427174155,
		},
	}
	if *resp != expectedResp {
		t.Errorf("want resp = %#v, got %#v", expectedResp, resp)
	}
}

func TestPushSingleDeviceError(t *testing.T) {
	timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
	signRequest = makeSignRequestFunc("1234567890")
	defer func() {
		timeNow = func() time.Time { return time.Now().UTC() }
		signRequest = signRequestFunc
	}()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		fmt.Fprintln(w, `{"request_id":123456789,"error_code":30602,"error_msg":"Request Params Not Valid, parameter apikey format invalid"}`)
	}))
	defer ts.Close()

	req := NewPushSingleDeviceRequest("fake_channel_id", NewAndroidNotificationMsg("Title", "Description"))
	client := newClient(ts.Client(), ts.URL, "fake_api_key", "fake_secret_key")
	_, err := client.PushSingleDevice(req)
	if err == nil {
		t.Errorf("want err != nil, got nil")
	}

	expectedErr := Error{
		RequestID: 123456789,
		Code:      30602,
		Msg:       "Request Params Not Valid, parameter apikey format invalid",
	}
	if err != expectedErr {
		t.Errorf("want err = %#v, got %#v", err, expectedErr)
	}
}

func mustReadAll(r io.Reader) []byte {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}

	return b
}

func urlValuesFromMap(m map[string]string) url.Values {
	values := url.Values{}
	for k, v := range m {
		values.Set(k, v)
	}

	return values
}
