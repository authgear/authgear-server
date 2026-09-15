package custom

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	utilhttputil "github.com/authgear/authgear-server/pkg/util/httputil"
)

type mockWebHook struct {
}

var _ hook.WebHook = &mockWebHook{}

// PerformNoResponse implements hook.WebHook.
func (m *mockWebHook) PerformNoResponse(ctx context.Context, client *http.Client, request *http.Request) error {
	panic("not implemented")
}

// PerformWithResponse implements hook.WebHook.
func (m *mockWebHook) PerformWithResponse(ctx context.Context, client *http.Client, request *http.Request) (resp *http.Response, err error) {
	panic("not implemented")
}

// PrepareRequest implements hook.WebHook.
func (m *mockWebHook) PrepareRequest(ctx context.Context, u *url.URL, body any) (*http.Request, error) {
	return &http.Request{}, nil
}

// SupportURL implements hook.WebHook.
func (m *mockWebHook) SupportURL(u *url.URL) bool {
	return true
}

type mockWebHookClient struct {
	ResponseStatusCode int
	ResponseBody       io.ReadCloser
}

// Do implements HookHTTPClient.
func (m *mockWebHookClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: m.ResponseStatusCode, Body: m.ResponseBody}, nil
}

var _ HookHTTPClient = &mockWebHookClient{}

// EnvHookHTTPClient is a distinct interface from HookHTTPClient, so it needs
// its own mock.
type mockEnvWebHookClient struct {
	ResponseStatusCode int
	ResponseBody       io.ReadCloser
}

// Do implements EnvHookHTTPClient.
func (m *mockEnvWebHookClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: m.ResponseStatusCode, Body: m.ResponseBody}, nil
}

var _ EnvHookHTTPClient = &mockEnvWebHookClient{}

type mockDenoHook struct {
	Output any
}

// RunSync implements DenoHook.
func (m *mockDenoHook) RunSync(ctx context.Context, client hook.DenoClient, u *url.URL, input any) (out any, err error) {
	return m.Output, nil
}

// SupportURL implements DenoHook.
func (m *mockDenoHook) SupportURL(u *url.URL) bool {
	return true
}

var _ DenoHook = &mockDenoHook{}

type mockDenoHookClient struct{}

func (m *mockDenoHookClient) Run(ctx context.Context, script string, input any) (out any, err error) {
	return nil, nil
}

var _ HookDenoClient = &mockDenoHookClient{}

func TestCustomClient(t *testing.T) {
	Convey("webhook is compatible with old clients", t, func() {
		// Originally we only check the status code of the webhook response
		// And do not care about the response body
		// We don't want to break this behavior

		Convey("empty response is ok", func() {
			var webhook hook.WebHook = &mockWebHook{}

			smsWebHook := &SMSWebHook{
				WebHook: webhook,
				Client: &mockWebHookClient{
					ResponseStatusCode: 200,
					ResponseBody:       io.NopCloser(strings.NewReader("")),
				},
			}
			ctx := context.Background()
			u := &url.URL{}
			err := smsWebHook.Call(ctx, u, SendOptions{})

			So(err, ShouldBeNil)
		})

		Convey("response not compatible with current response schema is ok", func() {
			webhook := &mockWebHook{}

			smsWebHook := &SMSWebHook{
				WebHook: webhook,
				Client: &mockWebHookClient{
					ResponseStatusCode: 200,
					ResponseBody:       io.NopCloser(strings.NewReader(`{"code": 1}`)),
				},
			}
			ctx := context.Background()
			u := &url.URL{}
			err := smsWebHook.Call(ctx, u, SendOptions{})

			So(err, ShouldBeNil)
		})
	})

	Convey("denohook is compatible with old clients", t, func() {
		// Originally we only check the status code of the webhook response
		// And do not care about the response body
		// We don't want to break this behavior

		Convey("null output is ok", func() {
			var denohook DenoHook = &mockDenoHook{
				Output: nil,
			}

			smsDenoHook := &SMSDenoHook{
				DenoHook: denohook,
				Client:   &mockDenoHookClient{},
			}
			ctx := context.Background()
			url := &url.URL{}
			err := smsDenoHook.Call(ctx, url, SendOptions{})

			So(err, ShouldBeNil)
		})

		Convey("incompatible output is ok", func() {
			var denohook DenoHook = &mockDenoHook{
				Output: map[string]any{"code": 1},
			}

			smsDenoHook := &SMSDenoHook{
				DenoHook: denohook,
				Client:   &mockDenoHookClient{},
			}
			ctx := context.Background()
			url := &url.URL{}
			err := smsDenoHook.Call(ctx, url, SendOptions{})

			So(err, ShouldBeNil)
		})
	})

	Convey("env webhook is compatible with old clients", t, func() {
		// EnvSMSWebHook shares callWebHook with SMSWebHook, so it must answer
		// these exactly as SMSWebHook does above.

		Convey("empty response is ok", func() {
			var webhook hook.WebHook = &mockWebHook{}

			envSMSWebHook := &EnvSMSWebHook{
				WebHook: webhook,
				Client: &mockEnvWebHookClient{
					ResponseStatusCode: 200,
					ResponseBody:       io.NopCloser(strings.NewReader("")),
				},
			}
			ctx := context.Background()
			u := &url.URL{}
			err := envSMSWebHook.Call(ctx, u, SendOptions{})

			So(err, ShouldBeNil)
		})

		Convey("response not compatible with current response schema is ok", func() {
			webhook := &mockWebHook{}

			envSMSWebHook := &EnvSMSWebHook{
				WebHook: webhook,
				Client: &mockEnvWebHookClient{
					ResponseStatusCode: 200,
					ResponseBody:       io.NopCloser(strings.NewReader(`{"code": 1}`)),
				},
			}
			ctx := context.Background()
			u := &url.URL{}
			err := envSMSWebHook.Call(ctx, u, SendOptions{})

			So(err, ShouldBeNil)
		})
	})
}

func TestNewEnvSMSHookTimeout(t *testing.T) {
	Convey("NewEnvSMSHookTimeout", t, func() {
		cases := []struct {
			Timeout  string
			Expected time.Duration
		}{
			// A zero duration on http.Client.Timeout means no timeout at
			// all, so none of these may produce one.
			{"", DefaultSMSHookTimeout},
			{"0", DefaultSMSHookTimeout},
			{"-5", DefaultSMSHookTimeout},
			{"abc", DefaultSMSHookTimeout},
			{"30", 30 * time.Second},
		}

		for _, c := range cases {
			actual := NewEnvSMSHookTimeout(config.SMSGatewayEnvironmentCustomSMSProviderConfig{
				Timeout: c.Timeout,
			})
			So(actual.Timeout, ShouldEqual, c.Expected)
			So(actual.Timeout, ShouldNotEqual, time.Duration(0))
		}
	})
}

func TestWebHookClientAddressPolicy(t *testing.T) {
	Convey("only the project-configured webhook enforces the address policy", t, func() {
		// httptest listens on 127.0.0.1, which is loopback and therefore not
		// publicly routable.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		}))
		defer server.Close()

		newRequest := func() *http.Request {
			req, err := http.NewRequest("POST", server.URL, strings.NewReader("{}"))
			So(err, ShouldBeNil)
			return req
		}

		Convey("the authgear.secrets.yaml client refuses a loopback address", func() {
			client := NewHookHTTPClient(SMSHookTimeout{Timeout: 5 * time.Second}, &config.HTTPFeatureConfig{})

			_, err := client.Do(newRequest())

			So(err, ShouldNotBeNil)
			So(errors.Is(err, utilhttputil.ErrBlockedAddress), ShouldBeTrue)
		})

		Convey("the SMS_GATEWAY_CUSTOM_URL client reaches it", func() {
			client := NewEnvHookHTTPClient(EnvSMSHookTimeout{Timeout: 5 * time.Second})

			resp, err := client.Do(newRequest())

			So(err, ShouldBeNil)
			defer resp.Body.Close()
			So(resp.StatusCode, ShouldEqual, 200)
		})
	})
}
