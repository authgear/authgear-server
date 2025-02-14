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

type mockHook struct {
}

var _ hook.WebHook = &mockHook{}

// PerformNoResponse implements hook.WebHook.
func (m *mockHook) PerformNoResponse(client *http.Client, request *http.Request) error {
	panic("not implemented")
}

// PerformWithResponse implements hook.WebHook.
func (m *mockHook) PerformWithResponse(client *http.Client, request *http.Request) (resp *http.Response, err error) {
	panic("not implemented")
}

// PrepareRequest implements hook.WebHook.
func (m *mockHook) PrepareRequest(ctx context.Context, u *url.URL, body interface{}) (*http.Request, error) {
	return &http.Request{}, nil
}

// SupportURL implements hook.WebHook.
func (m *mockHook) SupportURL(u *url.URL) bool {
	return true
}

type mockHookClient struct {
	ResponseStatusCode int
	ResponseBody       io.ReadCloser
}

// Do implements HookHTTPClient.
func (m *mockHookClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: m.ResponseStatusCode, Body: m.ResponseBody}, nil
}

var _ HookHTTPClient = &mockHookClient{}

func TestCustomClient(t *testing.T) {
	Convey("Is compatible with old clients", t, func() {
		// Originally we only check the status code of the webhook response
		// And do not care about the response body
		// We don't want to break this behavior

		var webhook hook.WebHook = &mockHook{}

		smsWebHook := &SMSWebHook{
			WebHook: webhook,
			Client: &mockHookClient{
				ResponseStatusCode: 200,
				ResponseBody:       io.NopCloser(strings.NewReader("")),
			},
		}
		ctx := context.Background()
		url := &url.URL{}
		err := smsWebHook.Call(ctx, url, SendOptions{})

		So(err, ShouldBeNil)
	})
}
