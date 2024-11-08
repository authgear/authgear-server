package hook

import (
	"context"
	"net/http"
	"net/url"
	"runtime"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestEventWebHook(t *testing.T) {
	mustURL := func(s string) *url.URL {
		u, err := url.Parse(s)
		if err != nil {
			panic(err)
		}
		return u
	}

	Convey("EventWebHook", t, func() {
		key, err := jwk.FromRaw([]byte("aG9vay1zZWNyZXQ"))
		So(err, ShouldBeNil)
		set := jwk.NewSet()
		_ = set.AddKey(key)
		secret := &config.WebhookKeyMaterials{
			Set: set,
		}
		httpClient := &http.Client{}

		gock.InterceptClient(httpClient)
		defer gock.Off()

		webhook := &EventWebHookImpl{
			WebHookImpl: WebHookImpl{Secret: secret},
			SyncHTTP:    SyncHTTPClient{httpClient},
			AsyncHTTP:   AsyncHTTPClient{httpClient},
		}

		Convey("DeliverBlockingEvent", func() {
			e := event.Event{
				ID:   "event-id",
				Type: MockBlockingEventType1,
			}
			gock.New("https://example.com").
				Post("/a").
				JSON(e).
				HeaderPresent(HeaderRequestBodySignature).
				Reply(200).
				JSON(map[string]interface{}{
					"is_allowed": true,
				})
			defer func() { gock.Flush() }()

			ctx := context.Background()
			resp, err := webhook.DeliverBlockingEvent(ctx, mustURL("https://example.com/a"), &e)

			So(err, ShouldBeNil)
			So(resp, ShouldResemble, &event.HookResponse{
				IsAllowed: true,
			})
		})

		Convey("DeliverNonBlockingEvent", func() {
			e := event.Event{
				ID:   "event-id",
				Type: MockNonBlockingEventType1,
			}
			gock.New("https://example.com").
				Post("/a").
				JSON(e).
				HeaderPresent(HeaderRequestBodySignature).
				Reply(200)
			defer func() { gock.Flush() }()

			ctx := context.Background()
			err := webhook.DeliverNonBlockingEvent(ctx, mustURL("https://example.com/a"), &e)
			runtime.Gosched()
			time.Sleep(500 * time.Millisecond)
			So(err, ShouldBeNil)
		})

		Convey("invalid response body", func() {
			e := event.Event{
				ID:   "event-id",
				Type: MockBlockingEventType1,
			}
			gock.New("https://example.com").
				Post("/a").
				JSON(e).
				HeaderPresent(HeaderRequestBodySignature).
				Reply(200)
			defer func() { gock.Flush() }()

			ctx := context.Background()
			_, err := webhook.DeliverBlockingEvent(ctx, mustURL("https://example.com/a"), &e)
			So(err, ShouldBeError, "invalid response body")
		})
	})
}
