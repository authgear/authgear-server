package custom

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/hook"
)

type mockWebHook struct {
}

var _ hook.WebHook = &mockWebHook{}

// PerformNoResponse implements hook.WebHook.
func (m *mockWebHook) PerformNoResponse(client *http.Client, request *http.Request) error {
	panic("not implemented")
}

// PerformWithResponse implements hook.WebHook.
func (m *mockWebHook) PerformWithResponse(client *http.Client, request *http.Request) (resp *http.Response, err error) {
	panic("not implemented")
}

// PrepareRequest implements hook.WebHook.
func (m *mockWebHook) PrepareRequest(ctx context.Context, u *url.URL, body interface{}) (*http.Request, error) {
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

type mockDenoHook struct {
	Output interface{}
}

// RunSync implements DenoHook.
func (m *mockDenoHook) RunSync(ctx context.Context, client hook.DenoClient, u *url.URL, input interface{}) (out interface{}, err error) {
	return m.Output, nil
}

// SupportURL implements DenoHook.
func (m *mockDenoHook) SupportURL(u *url.URL) bool {
	return true
}

var _ DenoHook = &mockDenoHook{}

type mockDenoHookClient struct{}

func (m *mockDenoHookClient) Run(ctx context.Context, script string, input interface{}) (out interface{}, err error) {
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
				Output: map[string]interface{}{"code": 1},
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
}
