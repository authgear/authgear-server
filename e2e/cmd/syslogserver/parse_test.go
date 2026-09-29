package main

import (
	"bufio"
	"bytes"
	"testing"
)

func TestParseMessage(t *testing.T) {
	frame := `<134>1 2026-07-27T10:42:36Z myproject.authgear.cloud authgear - authgear-audit-log [authgear app_id="myproject" id="00000000000a5a60" activity_type="user.authenticated" user_id="00000000-0000-0000-0000-000000000001" client_id="0000000000000000" ip_address="203.0.113.9"] {"id":"00000000000a5a60","seq":678496}`

	m := parseMessage([]byte(frame), "newline", 5140)
	if m.ParseError != "" {
		t.Fatalf("unexpected parse error: %s", m.ParseError)
	}
	if m.PRI != 134 || m.Facility != 16 || m.Severity != 6 {
		t.Errorf("PRI/facility/severity: got %d/%d/%d", m.PRI, m.Facility, m.Severity)
	}
	if m.Version != 1 {
		t.Errorf("version: got %d", m.Version)
	}
	if m.Timestamp != "2026-07-27T10:42:36Z" {
		t.Errorf("timestamp: got %q", m.Timestamp)
	}
	if m.Hostname != "myproject.authgear.cloud" {
		t.Errorf("hostname: got %q", m.Hostname)
	}
	if m.AppName != "authgear" || m.ProcID != "-" || m.MsgID != "authgear-audit-log" {
		t.Errorf("app_name/procid/msgid: got %q/%q/%q", m.AppName, m.ProcID, m.MsgID)
	}
	if m.StructuredDataID != "authgear" {
		t.Errorf("structured_data_id: got %q", m.StructuredDataID)
	}
	if m.StructuredData["app_id"] != "myproject" {
		t.Errorf("structured_data.app_id: got %q", m.StructuredData["app_id"])
	}
	if m.StructuredData["activity_type"] != "user.authenticated" {
		t.Errorf("structured_data.activity_type: got %q", m.StructuredData["activity_type"])
	}
	if m.StructuredData["ip_address"] != "203.0.113.9" {
		t.Errorf("structured_data.ip_address: got %q", m.StructuredData["ip_address"])
	}
	if m.Message != `{"id":"00000000000a5a60","seq":678496}` {
		t.Errorf("message: got %q", m.Message)
	}
}

func TestParseMessageEscapedStructuredData(t *testing.T) {
	// A value containing an escaped backslash, quote and closing bracket.
	frame := `<6>1 2021-01-01T00:00:00Z host authgear - authgear-audit-log [authgear client_id="has \"quote\", \\backslash and \]bracket"] {}`
	m := parseMessage([]byte(frame), "newline", 5140)
	if m.ParseError != "" {
		t.Fatalf("unexpected parse error: %s", m.ParseError)
	}
	got := m.StructuredData["client_id"]
	want := `has "quote", \backslash and ]bracket`
	if got != want {
		t.Errorf("client_id: got %q want %q", got, want)
	}
	if m.Message != "{}" {
		t.Errorf("message: got %q", m.Message)
	}
}

func TestParseMessageOmittedParams(t *testing.T) {
	frame := `<6>1 2021-01-01T00:00:00Z host authgear - authgear-audit-log [authgear app_id="a" id="b" activity_type="c"] {}`
	m := parseMessage([]byte(frame), "newline", 5140)
	if m.ParseError != "" {
		t.Fatalf("unexpected parse error: %s", m.ParseError)
	}
	if _, ok := m.StructuredData["user_id"]; ok {
		t.Errorf("user_id should be absent, got %q", m.StructuredData["user_id"])
	}
}

func TestParseMessageMalformed(t *testing.T) {
	cases := []string{
		"not a syslog message",
		"<134 missing closing angle bracket",
		`<134>1 2021-01-01T00:00:00Z host app - msgid no-brackets-here {}`,
	}
	for _, frame := range cases {
		m := parseMessage([]byte(frame), "newline", 5140)
		if m.ParseError == "" {
			t.Errorf("frame %q: expected a parse error, got none", frame)
		}
	}
}

func TestReadFrameNewline(t *testing.T) {
	r := bufio.NewReader(bytes.NewReader([]byte("hello\nworld\n")))

	frame, framing, err := readFrame(r)
	if err != nil {
		t.Fatal(err)
	}
	if framing != "newline" || string(frame) != "hello" {
		t.Errorf("got framing=%q frame=%q", framing, frame)
	}

	frame, framing, err = readFrame(r)
	if err != nil {
		t.Fatal(err)
	}
	if framing != "newline" || string(frame) != "world" {
		t.Errorf("got framing=%q frame=%q", framing, frame)
	}
}

func TestReadFrameOctetCounting(t *testing.T) {
	// "5 héllo" where héllo is 6 bytes (é is 2 bytes in UTF-8), matching
	// how the sender counts bytes, not runes.
	msg := "héllo"
	input := "6 " + msg + "6 " + msg
	r := bufio.NewReader(bytes.NewReader([]byte(input)))

	frame, framing, err := readFrame(r)
	if err != nil {
		t.Fatal(err)
	}
	if framing != "octet_counting" || string(frame) != msg {
		t.Errorf("got framing=%q frame=%q", framing, frame)
	}

	frame, framing, err = readFrame(r)
	if err != nil {
		t.Fatal(err)
	}
	if framing != "octet_counting" || string(frame) != msg {
		t.Errorf("second frame: got framing=%q frame=%q", framing, frame)
	}
}
