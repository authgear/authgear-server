package auditlogstreaming

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
)

func TestEncodeDatadogLog(t *testing.T) {
	baseDatadog := &ResolvedDatadog{
		Service: "authgear",
		Source:  "authgear",
	}

	baseEvent := func() *event.Event {
		return &event.Event{
			ID:   "00000000000a5a60",
			Seq:  678496,
			Type: nonblocking.UserAuthenticated,
			Context: event.Context{
				Timestamp:       1785148956,
				AppID:           "myproject",
				UserID:          new("00000000-0000-0000-0000-000000000001"),
				ClientID:        "0000000000000000",
				IPAddress:       "203.0.113.9",
				GeoLocationCode: new("TW"),
				UserAgent:       "Mozilla/5.0 ...",
				AuditContext: event.AuditContext{
					"http_url":     "https://myproject.authgear.cloud/oauth2/token",
					"http_referer": "https://myproject.authgear.cloud/login",
				},
			},
		}
	}

	rawEvent := []byte(`{"id":"00000000000a5a60","seq":678496,"type":"user.authenticated"}`)
	entry := func() QueuedEntry {
		return QueuedEntry{Event: baseEvent(), Raw: rawEvent}
	}

	Convey("EncodeDatadogLog", t, func() {
		Convey("matches the spec's example shape", func() {
			got, err := EncodeDatadogLog(baseDatadog, "myproject.authgear.cloud", entry())
			So(err, ShouldBeNil)

			var actual map[string]any
			So(json.Unmarshal(got, &actual), ShouldBeNil)

			expectedJSON := `{
				"ddsource": "authgear",
				"service": "authgear",
				"hostname": "myproject.authgear.cloud",
				"ddtags": "app_id:myproject,activity_type:user.authenticated",
				"message": "user.authenticated",
				"status": "info",
				"timestamp": 1785148956000,
				"evt": { "name": "user.authenticated" },
				"usr": { "id": "00000000-0000-0000-0000-000000000001" },
				"network": {
					"client": {
						"ip": "203.0.113.9",
						"geoip": { "country": { "iso_code": "TW" } }
					}
				},
				"http": {
					"useragent": "Mozilla/5.0 ...",
					"url": "https://myproject.authgear.cloud/oauth2/token",
					"referer": "https://myproject.authgear.cloud/login"
				},
				"authgear": {
					"event": {"id":"00000000000a5a60","seq":678496,"type":"user.authenticated"}
				}
			}`
			var expected map[string]any
			So(json.Unmarshal([]byte(expectedJSON), &expected), ShouldBeNil)

			So(actual, ShouldResemble, expected)
		})

		Convey("authgear.event is byte-identical to QueuedEntry.Raw", func() {
			e := entry()
			e.Raw = []byte(`{"id":"custom","seq":1,"type":"user.authenticated","extra":"field"}`)
			got, err := EncodeDatadogLog(baseDatadog, "host", e)
			So(err, ShouldBeNil)

			var actual struct {
				Authgear struct {
					Event json.RawMessage `json:"event"`
				} `json:"authgear"`
			}
			So(json.Unmarshal(got, &actual), ShouldBeNil)
			So(string(actual.Authgear.Event), ShouldEqual, string(e.Raw))
		})

		Convey("timestamp is Context.Timestamp * 1000", func() {
			e := entry()
			e.Event.Context.Timestamp = 1609459200
			got, err := EncodeDatadogLog(baseDatadog, "host", e)
			So(err, ShouldBeNil)
			So(string(got), ShouldContainSubstring, `"timestamp":1609459200000`)
		})

		Convey("a nil Context.UserID omits the whole usr object", func() {
			e := entry()
			e.Event.Context.UserID = nil
			got, err := EncodeDatadogLog(baseDatadog, "host", e)
			So(err, ShouldBeNil)
			So(string(got), ShouldNotContainSubstring, `"usr"`)
		})

		Convey("empty IPAddress and nil GeoLocationCode omit the whole network object", func() {
			e := entry()
			e.Event.Context.IPAddress = ""
			e.Event.Context.GeoLocationCode = nil
			got, err := EncodeDatadogLog(baseDatadog, "host", e)
			So(err, ShouldBeNil)
			So(string(got), ShouldNotContainSubstring, `"network"`)
		})

		Convey("empty user agent, url and referer omit the whole http object", func() {
			e := entry()
			e.Event.Context.UserAgent = ""
			e.Event.Context.AuditContext = event.AuditContext{}
			got, err := EncodeDatadogLog(baseDatadog, "host", e)
			So(err, ShouldBeNil)
			So(string(got), ShouldNotContainSubstring, `"http"`)
		})

		Convey("a non-string under http_url or http_referer is treated as absent", func() {
			e := entry()
			e.Event.Context.UserAgent = ""
			e.Event.Context.AuditContext = event.AuditContext{
				"http_url":     123,
				"http_referer": []string{"not-a-string"},
			}
			got, err := EncodeDatadogLog(baseDatadog, "host", e)
			So(err, ShouldBeNil)
			So(string(got), ShouldNotContainSubstring, `"http"`)
		})

		Convey("hostname is omitted when resolved hostname is empty, present otherwise", func() {
			got, err := EncodeDatadogLog(baseDatadog, "", entry())
			So(err, ShouldBeNil)
			So(string(got), ShouldNotContainSubstring, `"hostname"`)

			got, err = EncodeDatadogLog(baseDatadog, "myproject.authgear.cloud", entry())
			So(err, ShouldBeNil)
			So(string(got), ShouldContainSubstring, `"hostname":"myproject.authgear.cloud"`)
		})

		Convey("ddsource, service, message, status, timestamp and authgear.event are present even when every optional field is empty", func() {
			e := QueuedEntry{
				Event: &event.Event{Type: "some.new.type", Context: event.Context{}},
				Raw:   []byte(`{}`),
			}
			got, err := EncodeDatadogLog(baseDatadog, "", e)
			So(err, ShouldBeNil)

			var actual map[string]any
			So(json.Unmarshal(got, &actual), ShouldBeNil)
			So(actual["ddsource"], ShouldEqual, "authgear")
			So(actual["service"], ShouldEqual, "authgear")
			So(actual["message"], ShouldEqual, "some.new.type")
			So(actual["status"], ShouldEqual, "info")
			So(actual["timestamp"], ShouldEqual, float64(0))
			So(actual["authgear"], ShouldNotBeNil)
		})
	})
}

func TestDatadogStatus(t *testing.T) {
	Convey("datadogStatus", t, func() {
		So(datadogStatus(nonblocking.EmailError), ShouldEqual, "warning")
		So(datadogStatus(nonblocking.SMSError), ShouldEqual, "warning")
		So(datadogStatus(nonblocking.WhatsappError), ShouldEqual, "warning")
		So(datadogStatus(nonblocking.UserAuthenticated), ShouldEqual, "info")
		So(datadogStatus("some.new.type"), ShouldEqual, "info")
	})
}

func TestBuildDDTags(t *testing.T) {
	Convey("buildDDTags", t, func() {
		Convey("configured pairs come first in order, then app_id, then activity_type", func() {
			tags := []DatadogTag{
				{Key: "env", Value: "production"},
				{Key: "team", Value: "security"},
			}
			got := buildDDTags(tags, "myproject", "user.authenticated")
			So(got, ShouldEqual, "env:production,team:security,app_id:myproject,activity_type:user.authenticated")
		})

		Convey("no configured tags still yields app_id and activity_type", func() {
			got := buildDDTags(nil, "myproject", "user.authenticated")
			So(got, ShouldEqual, "app_id:myproject,activity_type:user.authenticated")
		})

		Convey("an empty appID omits the app_id tag", func() {
			got := buildDDTags(nil, "", "user.authenticated")
			So(got, ShouldEqual, "activity_type:user.authenticated")
		})
	})
}
