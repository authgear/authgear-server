package auditlogstreaming

import (
	"compress/gzip"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

// stubDatadogClientFactory hands back a fixed client -- srv.Client() in
// every test here -- so these run without minting a certificate
// authority. httputil.NewSSRFSafeExternalClient's RootCAs handling is
// covered directly in pkg/util/httputil.
type stubDatadogClientFactory struct {
	client *http.Client
}

func (f *stubDatadogClientFactory) MakeClient(rootCAs *x509.CertPool) *http.Client {
	return f.client
}

// datadogRecorder records every request made to it, gunzipping the body
// and decoding it as a JSON array.
type datadogRecorder struct {
	mu       sync.Mutex
	requests [][]map[string]any
	headers  []http.Header
	// sizes is the uncompressed body length of each request, in the same
	// order as requests -- what Datadog's 5MB limit is measured against.
	sizes []int

	// statusFor, when set, picks the response status for the n-th request
	// (0-indexed). Returns 202 when nil or out of range.
	statusFor func(n int) int
}

func (r *datadogRecorder) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var body io.Reader = req.Body
		if req.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(req.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer gz.Close()
			body = gz
		}

		raw, err := io.ReadAll(body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var entries []map[string]any
		if err := json.Unmarshal(raw, &entries); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		r.mu.Lock()
		n := len(r.requests)
		r.requests = append(r.requests, entries)
		r.headers = append(r.headers, req.Header.Clone())
		r.sizes = append(r.sizes, len(raw))
		status := http.StatusAccepted
		if r.statusFor != nil {
			status = r.statusFor(n)
		}
		r.mu.Unlock()

		w.WriteHeader(status)
	}
}

func (r *datadogRecorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.requests)
}

func (r *datadogRecorder) all() [][]map[string]any {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([][]map[string]any, len(r.requests))
	copy(out, r.requests)
	return out
}

func newDatadogTestEntry(message string) QueuedEntry {
	raw := fmt.Appendf(nil, `{"marker":%q}`, message)
	return QueuedEntry{
		Event: &event.Event{
			ID:   message,
			Type: event.Type(message),
			Context: event.Context{
				Timestamp: 1700000000,
				AppID:     "app",
			},
		},
		Raw: raw,
	}
}

// resolvedTestDatadog resolves the exact same ResolvedDatadog
// newDatadogTestSender's stream resolves to, so a test that pre-computes
// an encoded log's length off of it stays consistent with what
// sendDatadogHTTP actually encodes.
func resolvedTestDatadog(t *testing.T) *ResolvedDatadog {
	t.Helper()
	streamConfig := newTestDatadogStreamConfig("datadog", nil, "http://example.invalid/logs")
	credentials := &config.TelemetryAuditLogStreamDatadogCredentials{
		{StreamName: "datadog", APIKey: "test-api-key"},
	}
	resolved, err := resolveStream("app", streamConfig, nil, credentials)
	if err != nil {
		t.Fatal(err)
	}
	return resolved.Datadog
}

func newDatadogTestSender(endpoint string, client *http.Client) *SenderImpl {
	streamConfig := newTestDatadogStreamConfig("datadog", nil, endpoint)
	credentials := &config.TelemetryAuditLogStreamDatadogCredentials{
		{StreamName: "datadog", APIKey: "test-api-key"},
	}
	return &SenderImpl{
		AppID:                "app",
		Hostname:             "app.example.com",
		Streams:              []*config.TelemetryAuditLogStreamConfig{streamConfig},
		DatadogCredentials:   credentials,
		DatadogClientFactory: &stubDatadogClientFactory{client: client},
	}
}

func TestSenderImplSendDatadog(t *testing.T) {
	Convey("SenderImpl.Send: datadog/http", t, func() {
		Convey("one entry produces one POST, gzipped, with the expected headers", func() {
			rec := &datadogRecorder{}
			srv := httptest.NewServer(rec.handler())
			defer srv.Close()

			sender := newDatadogTestSender(srv.URL+"/api/v2/logs", srv.Client())
			sender.Send(t.Context(), []QueuedEntry{newDatadogTestEntry("user.created")})

			So(rec.count(), ShouldEqual, 1)
			entries := rec.all()[0]
			So(entries, ShouldHaveLength, 1)
			So(entries[0]["message"], ShouldEqual, "user.created")

			headers := rec.headers[0]
			So(headers.Get("DD-API-KEY"), ShouldEqual, "test-api-key")
			So(headers.Get("Content-Type"), ShouldEqual, "application/json")
			So(headers.Get("Content-Encoding"), ShouldEqual, "gzip")
		})

		Convey("entries split by count when datadogMaxEntriesPerRequest is shrunk", func() {
			oldMax := datadogMaxEntriesPerRequest
			datadogMaxEntriesPerRequest = 3
			defer func() { datadogMaxEntriesPerRequest = oldMax }()

			rec := &datadogRecorder{}
			srv := httptest.NewServer(rec.handler())
			defer srv.Close()

			sender := newDatadogTestSender(srv.URL+"/api/v2/logs", srv.Client())
			var entries []QueuedEntry
			for i := range 7 {
				entries = append(entries, newDatadogTestEntry(fmt.Sprintf("event-%d", i)))
			}
			sender.Send(t.Context(), entries)

			So(rec.count(), ShouldEqual, 3)
			all := rec.all()
			So(all[0], ShouldHaveLength, 3)
			So(all[1], ShouldHaveLength, 3)
			So(all[2], ShouldHaveLength, 1)

			// Occurrence order is preserved across requests.
			So(all[0][0]["message"], ShouldEqual, "event-0")
			So(all[1][0]["message"], ShouldEqual, "event-3")
			So(all[2][0]["message"], ShouldEqual, "event-6")
		})

		Convey("entries split by size, not by count, when datadogMaxRequestBytes is shrunk", func() {
			testDatadogSplitBySize(t)
		})

		Convey("the boundary: a log of exactly datadogMaxLogBytes is delivered alone, one byte larger is dropped", func() {
			testDatadogMaxLogBytesBoundary(t)
		})

		Convey("an entry alone larger than the limit is skipped, and the entries around it are still delivered", func() {
			rec := &datadogRecorder{}
			srv := httptest.NewServer(rec.handler())
			defer srv.Close()

			sender := newDatadogTestSender(srv.URL+"/api/v2/logs", srv.Client())

			huge := newDatadogTestEntry("huge")
			huge.Raw = make([]byte, datadogMaxLogBytes+1000)
			// EncodeDatadogLog does not use Raw's content, only its bytes
			// via authgear.event, so this needs to be valid JSON.
			huge.Raw[0] = '"'
			for i := 1; i < len(huge.Raw)-1; i++ {
				huge.Raw[i] = 'a'
			}
			huge.Raw[len(huge.Raw)-1] = '"'

			sender.Send(t.Context(), []QueuedEntry{
				newDatadogTestEntry("before"), huge, newDatadogTestEntry("after"),
			})

			So(rec.count(), ShouldEqual, 1)
			entries := rec.all()[0]
			So(entries, ShouldHaveLength, 2)
			So(entries[0]["message"], ShouldEqual, "before")
			So(entries[1]["message"], ShouldEqual, "after")
		})

		Convey("no request ever carries an empty array, and no request is made when every entry is dropped for size", func() {
			rec := &datadogRecorder{}
			srv := httptest.NewServer(rec.handler())
			defer srv.Close()

			sender := newDatadogTestSender(srv.URL+"/api/v2/logs", srv.Client())

			huge := newDatadogTestEntry("huge")
			huge.Raw = make([]byte, datadogMaxLogBytes+1000)
			huge.Raw[0] = '"'
			for i := 1; i < len(huge.Raw)-1; i++ {
				huge.Raw[i] = 'a'
			}
			huge.Raw[len(huge.Raw)-1] = '"'

			sender.Send(t.Context(), []QueuedEntry{huge})

			So(rec.count(), ShouldEqual, 0)
		})

		Convey("a 500 on the first of three chunks still sends the remaining two", func() {
			oldMax := datadogMaxEntriesPerRequest
			datadogMaxEntriesPerRequest = 1
			defer func() { datadogMaxEntriesPerRequest = oldMax }()

			rec := &datadogRecorder{statusFor: func(n int) int {
				if n == 0 {
					return http.StatusInternalServerError
				}
				return http.StatusAccepted
			}}
			srv := httptest.NewServer(rec.handler())
			defer srv.Close()

			sender := newDatadogTestSender(srv.URL+"/api/v2/logs", srv.Client())
			sender.Send(t.Context(), []QueuedEntry{
				newDatadogTestEntry("one"), newDatadogTestEntry("two"), newDatadogTestEntry("three"),
			})

			So(rec.count(), ShouldEqual, 3)
		})

		Convey("a 403 on the first of three chunks sends no further chunk", func() {
			oldMax := datadogMaxEntriesPerRequest
			datadogMaxEntriesPerRequest = 1
			defer func() { datadogMaxEntriesPerRequest = oldMax }()

			rec := &datadogRecorder{statusFor: func(n int) int {
				return http.StatusForbidden
			}}
			srv := httptest.NewServer(rec.handler())
			defer srv.Close()

			sender := newDatadogTestSender(srv.URL+"/api/v2/logs", srv.Client())
			sender.Send(t.Context(), []QueuedEntry{
				newDatadogTestEntry("one"), newDatadogTestEntry("two"), newDatadogTestEntry("three"),
			})

			So(rec.count(), ShouldEqual, 1)
		})

		Convey("a transport error drops that chunk and continues", func() {
			oldMax := datadogMaxEntriesPerRequest
			datadogMaxEntriesPerRequest = 1
			defer func() { datadogMaxEntriesPerRequest = oldMax }()

			var n atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if n.Add(1) == 1 {
					// Close the connection without responding, to force a
					// transport error on the first chunk only.
					hj, ok := w.(http.Hijacker)
					if ok {
						conn, _, _ := hj.Hijack()
						_ = conn.Close()
						return
					}
				}
				w.WriteHeader(http.StatusAccepted)
			}))
			defer srv.Close()

			sender := newDatadogTestSender(srv.URL+"/api/v2/logs", srv.Client())
			sender.Send(t.Context(), []QueuedEntry{
				newDatadogTestEntry("one"), newDatadogTestEntry("two"),
			})

			// Both chunks were attempted, even though the first failed at
			// the transport level.
			So(n.Load(), ShouldEqual, 2)
		})

		Convey("a server that never responds is abandoned when datadogBatchTimeout elapses", func() {
			oldBatchTimeout := datadogBatchTimeout
			datadogBatchTimeout = 200 * time.Millisecond
			defer func() { datadogBatchTimeout = oldBatchTimeout }()
			oldRequestTimeout := datadogRequestTimeout
			datadogRequestTimeout = 5 * time.Second
			defer func() { datadogRequestTimeout = oldRequestTimeout }()

			block := make(chan struct{})
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				<-block
			}))
			// close(block) must run before srv.Close(), or srv.Close()
			// deadlocks waiting for the still-blocked handler to return.
			defer func() {
				close(block)
				srv.Close()
			}()

			sender := newDatadogTestSender(srv.URL+"/api/v2/logs", srv.Client())

			done := make(chan struct{})
			go func() {
				sender.Send(t.Context(), []QueuedEntry{newDatadogTestEntry("one")})
				close(done)
			}()

			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("Send did not return after datadogBatchTimeout should have fired")
			}
		})
	})
}

// testDatadogSplitBySize is split out of TestSenderImplSendDatadog to keep
// its cognitive complexity under the gocognit threshold.
func testDatadogSplitBySize(t *testing.T) {
	datadog := resolvedTestDatadog(t)
	sample, err := EncodeDatadogLog(datadog, "app.example.com", newDatadogTestEntry("e00"))
	So(err, ShouldBeNil)
	logLen := len(sample)

	// Sized so that exactly 3 same-length entries fit a chunk: the fits()
	// bound after 2 already-added entries is 2*(logLen+1)+logLen+2, and
	// one byte under that for a 4th entry to overflow.
	oldMaxBytes := datadogMaxRequestBytes
	oldMaxLog := datadogMaxLogBytes
	datadogMaxRequestBytes = 3*(logLen+1) + logLen + 1
	datadogMaxLogBytes = datadogMaxRequestBytes
	defer func() {
		datadogMaxRequestBytes = oldMaxBytes
		datadogMaxLogBytes = oldMaxLog
	}()

	rec := &datadogRecorder{}
	srv := httptest.NewServer(rec.handler())
	defer srv.Close()

	sender := newDatadogTestSender(srv.URL+"/api/v2/logs", srv.Client())
	var entries []QueuedEntry
	for i := range 7 {
		entries = append(entries, newDatadogTestEntry(fmt.Sprintf("e%02d", i)))
	}
	sender.Send(t.Context(), entries)

	So(rec.count(), ShouldEqual, 3)
	all := rec.all()
	So(all[0], ShouldHaveLength, 3)
	So(all[1], ShouldHaveLength, 3)
	So(all[2], ShouldHaveLength, 1)
	for _, size := range rec.sizes {
		So(size, ShouldBeLessThanOrEqualTo, datadogMaxRequestBytes)
	}
}

// testDatadogMaxLogBytesBoundary is split out of TestSenderImplSendDatadog
// to keep its cognitive complexity under the gocognit threshold.
func testDatadogMaxLogBytesBoundary(t *testing.T) {
	oldMaxBytes := datadogMaxRequestBytes
	oldMaxLog := datadogMaxLogBytes
	datadogMaxRequestBytes = 2000
	datadogMaxLogBytes = datadogMaxRequestBytes - 3
	defer func() {
		datadogMaxRequestBytes = oldMaxBytes
		datadogMaxLogBytes = oldMaxLog
	}()

	datadog := resolvedTestDatadog(t)

	// Find the base length of the encoded log with an empty
	// authgear.event, then pad Raw so the whole log is exactly
	// datadogMaxLogBytes.
	baseEntry := newDatadogTestEntry("e")
	baseEntry.Raw = []byte(`""`)
	base, err := EncodeDatadogLog(datadog, "app.example.com", baseEntry)
	So(err, ShouldBeNil)
	baseLen := len(base) - len(baseEntry.Raw)

	padTo := func(n int) []byte {
		raw := make([]byte, n)
		raw[0] = '"'
		for i := 1; i < n-1; i++ {
			raw[i] = 'a'
		}
		raw[n-1] = '"'
		return raw
	}

	exact := newDatadogTestEntry("e")
	exact.Raw = padTo(datadogMaxLogBytes - baseLen)
	exactLog, err := EncodeDatadogLog(datadog, "app.example.com", exact)
	So(err, ShouldBeNil)
	So(len(exactLog), ShouldEqual, datadogMaxLogBytes)

	tooBig := newDatadogTestEntry("e")
	tooBig.Raw = padTo(datadogMaxLogBytes - baseLen + 1)
	tooBigLog, err := EncodeDatadogLog(datadog, "app.example.com", tooBig)
	So(err, ShouldBeNil)
	So(len(tooBigLog), ShouldEqual, datadogMaxLogBytes+1)

	rec := &datadogRecorder{}
	srv := httptest.NewServer(rec.handler())
	defer srv.Close()

	sender := newDatadogTestSender(srv.URL+"/api/v2/logs", srv.Client())
	sender.Send(t.Context(), []QueuedEntry{exact})
	So(rec.count(), ShouldEqual, 1)
	So(rec.all()[0], ShouldHaveLength, 1)

	rec2 := &datadogRecorder{}
	srv2 := httptest.NewServer(rec2.handler())
	defer srv2.Close()
	sender2 := newDatadogTestSender(srv2.URL+"/api/v2/logs", srv2.Client())
	sender2.Send(t.Context(), []QueuedEntry{tooBig})
	So(rec2.count(), ShouldEqual, 0)
}
