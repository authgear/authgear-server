package httputil_test

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestMaxResponseBytes(t *testing.T) {
	Convey("NewSSRFSafeExternalClient caps the response body", t, func() {
		ctx := t.Context()

		// The servers below are on loopback, so the address rules have to be
		// lifted for the client to reach them at all. The cap is independent
		// of those rules and still applies.
		newClient := func() *http.Client {
			return httputil.NewSSRFSafeExternalClient(5*time.Second, httputil.SSRFSafeExternalClientOptions{
				AllowNonPublicAddresses: true,
			})
		}

		serve := func(h http.HandlerFunc) string {
			server := httptest.NewServer(h)
			t.Cleanup(server.Close)
			return server.URL
		}

		read := func(url string) ([]byte, error) {
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			So(err, ShouldBeNil)
			resp, err := newClient().Do(req)
			So(err, ShouldBeNil)
			defer resp.Body.Close()
			return io.ReadAll(resp.Body)
		}

		Convey("a body of exactly the limit is read in full", func() {
			url := serve(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(bytes.Repeat([]byte("a"), int(httputil.MaxResponseBytes)))
			})

			body, err := read(url)
			So(err, ShouldBeNil)
			So(int64(len(body)), ShouldEqual, httputil.MaxResponseBytes)
		})

		Convey("one byte more is refused", func() {
			url := serve(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(bytes.Repeat([]byte("a"), int(httputil.MaxResponseBytes)+1))
			})

			_, err := read(url)
			So(errors.Is(err, httputil.ErrResponseTooLarge), ShouldBeTrue)
		})

		Convey("a small gzip body that expands past the limit is refused", func() {
			// Counting compressed bytes would let this through: it is a few
			// KiB on the wire and 8 MiB once decompressed.
			var compressed bytes.Buffer
			gz := gzip.NewWriter(&compressed)
			_, err := gz.Write(bytes.Repeat([]byte("a"), 8*1024*1024))
			So(err, ShouldBeNil)
			So(gz.Close(), ShouldBeNil)
			So(int64(compressed.Len()), ShouldBeLessThan, httputil.MaxResponseBytes)

			url := serve(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Encoding", "gzip")
				_, _ = w.Write(compressed.Bytes())
			})

			_, err = read(url)
			So(errors.Is(err, httputil.ErrResponseTooLarge), ShouldBeTrue)
		})
	})
}
