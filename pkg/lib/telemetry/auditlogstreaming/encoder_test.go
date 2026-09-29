package auditlogstreaming

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestEncodeRFC5424(t *testing.T) {
	baseSyslog := &ResolvedSyslog{
		Format:           config.SyslogFormatRFC5424,
		Framing:          config.SyslogFramingNewline,
		Facility:         16,
		AppName:          "authgear",
		StructuredDataID: "authgear",
	}

	baseEvent := func() *event.Event {
		return &event.Event{
			ID:   "00000000000a5a60",
			Seq:  678496,
			Type: "user.authenticated",
			Context: event.Context{
				Timestamp: 1785148956,
				AppID:     "myproject",
				UserID:    new("00000000-0000-0000-0000-000000000001"),
				ClientID:  "0000000000000000",
				IPAddress: "203.0.113.9",
			},
		}
	}

	msg := []byte(`{"id":"00000000000a5a60","seq":678496,"type":"user.authenticated"}`)

	Convey("EncodeRFC5424", t, func() {
		Convey("matches the spec's UC1 example shape", func() {
			got := EncodeRFC5424(baseSyslog, "myproject.authgear.cloud", baseEvent(), msg)
			expected := `<134>1 2026-07-27T10:42:36Z myproject.authgear.cloud authgear - authgear-audit-log ` +
				`[authgear app_id="myproject" id="00000000000a5a60" activity_type="user.authenticated" ` +
				`user_id="00000000-0000-0000-0000-000000000001" client_id="0000000000000000" ip_address="203.0.113.9"] ` +
				string(msg)
			So(string(got), ShouldEqual, expected)
		})

		Convey("PRI is facility*8 + severity", func() {
			s := &ResolvedSyslog{Facility: 0, AppName: "authgear", StructuredDataID: "authgear"}
			e := baseEvent()
			e.Type = "user.authenticated" // severity 6
			got := EncodeRFC5424(s, "host", e, msg)
			So(string(got), ShouldStartWith, "<6>1 ")

			s.Facility = 23
			e.Type = nonblocking.EmailError // severity 4
			got = EncodeRFC5424(s, "host", e, msg)
			So(string(got), ShouldStartWith, "<188>1 ")
		})

		Convey("severity", func() {
			So(severity(nonblocking.EmailError), ShouldEqual, 4)
			So(severity(nonblocking.SMSError), ShouldEqual, 4)
			So(severity(nonblocking.WhatsappError), ShouldEqual, 4)
			So(severity(nonblocking.UserAuthenticated), ShouldEqual, 6)
			So(severity(nonblocking.UsageAlertTriggered), ShouldEqual, 6)
			So(severity("some.new.type"), ShouldEqual, 6)
		})

		Convey("a nil Context.UserID omits user_id entirely", func() {
			e := baseEvent()
			e.Context.UserID = nil
			got := EncodeRFC5424(baseSyslog, "host", e, msg)
			So(string(got), ShouldNotContainSubstring, "user_id")
		})

		Convey("empty client_id, ip_address and app_id are omitted", func() {
			e := baseEvent()
			e.Context.ClientID = ""
			e.Context.IPAddress = ""
			e.Context.AppID = ""
			got := EncodeRFC5424(baseSyslog, "host", e, msg)
			So(string(got), ShouldNotContainSubstring, "client_id")
			So(string(got), ShouldNotContainSubstring, "ip_address")
			So(string(got), ShouldNotContainSubstring, "app_id")
		})

		Convey("structured data parameter escaping", func() {
			e := baseEvent()
			e.Context.ClientID = `has "quote", \backslash and ]bracket, and \"already-escaped`
			got := EncodeRFC5424(baseSyslog, "host", e, msg)
			So(string(got), ShouldContainSubstring,
				`client_id="has \"quote\", \\backslash and \]bracket, and \\\"already-escaped"`)
		})

		Convey("structured_data_id containing a caret round-trips unescaped", func() {
			s := &ResolvedSyslog{Facility: 16, AppName: "authgear", StructuredDataID: "authgear^b"}
			got := EncodeRFC5424(s, "host", baseEvent(), msg)
			So(string(got), ShouldContainSubstring, "[authgear^b ")
		})

		Convey("TIMESTAMP has no fractional seconds, ends in Z, and reflects Context.Timestamp", func() {
			e := baseEvent()
			e.Context.Timestamp = 1609459200 // 2021-01-01T00:00:00Z
			got := EncodeRFC5424(baseSyslog, "host", e, msg)
			So(string(got), ShouldContainSubstring, " 2021-01-01T00:00:00Z ")
		})

		Convey("MSG is byte-identical to the supplied msg and is not BOM-prefixed", func() {
			got := EncodeRFC5424(baseSyslog, "host", baseEvent(), msg)
			So(string(got), ShouldEndWith, string(msg))
			So(got[0], ShouldNotEqual, 0xEF) // UTF-8 BOM first byte
		})
	})
}

func TestFrame(t *testing.T) {
	Convey("Frame", t, func() {
		Convey("newline appends exactly one LF", func() {
			got := Frame(config.SyslogFramingNewline, []byte("hello"))
			So(string(got), ShouldEqual, "hello\n")
		})

		Convey("octet_counting prefixes the byte length, not the rune count", func() {
			// "héllo" is 6 bytes (é is 2 bytes in UTF-8) but 5 runes.
			msg := []byte("héllo")
			got := Frame(config.SyslogFramingOctetCounting, msg)
			So(string(got), ShouldEqual, "6 héllo")
		})
	})
}
