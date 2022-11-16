package hook

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/lestrrat-go/jwx/jwk"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestWebHook(t *testing.T) {
	mustURL := func(s string) *url.URL {
		u, err := url.Parse(s)
		if err != nil {
			panic(err)
		}
		return u
	}

	Convey("WebHook", t, func() {
		key, err := jwk.New([]byte("aG9vay1zZWNyZXQ"))
		So(err, ShouldBeNil)
		set := jwk.NewSet()
		_ = set.Add(key)
		secret := &config.WebhookKeyMaterials{
			Set: set,
		}
		httpClient := &http.Client{}

		gock.InterceptClient(httpClient)
		defer gock.Off()

		webhook := &WebHookImpl{
			Secret:    secret,
			SyncHTTP:  SyncHTTPClient{httpClient},
			AsyncHTTP: AsyncHTTPClient{httpClient},
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

			resp, err := webhook.DeliverBlockingEvent(mustURL("https://example.com/a"), &e)

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

			err := webhook.DeliverNonBlockingEvent(mustURL("https://example.com/a"), &e)
			So(err, ShouldBeNil)
		})

		Convey("invalid status code", func() {
			e := event.Event{
				ID:   "event-id",
				Type: MockBlockingEventType1,
			}
			gock.New("https://example.com").
				Post("/a").
				JSON(e).
				HeaderPresent(HeaderRequestBodySignature).
				Reply(500)
			defer func() { gock.Flush() }()

			err := webhook.DeliverNonBlockingEvent(mustURL("https://example.com/a"), &e)
			So(err, ShouldBeError, "invalid status code")
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

			_, err := webhook.DeliverBlockingEvent(mustURL("https://example.com/a"), &e)
			So(err, ShouldBeError, "invalid response body")
		})
	})
}
