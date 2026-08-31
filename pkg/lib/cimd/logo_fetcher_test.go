package cimd

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// Minimal byte sequences that satisfy http.DetectContentType's signature
// sniffing without being full, renderable images -- the sniff only reads a
// short magic-bytes prefix.
var (
	pngMagic  = []byte("\x89PNG\r\n\x1a\n" + "rest of a png file")
	jpegMagic = []byte("\xff\xd8\xff" + "rest of a jpeg file")
	htmlBody  = []byte("<html><body>not an image</body></html>")
	svgBody   = []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`)
)

func logoFetcherFor(httpClient *http.Client, featureConfig *config.OAuthClientIDMetadataDocumentFeatureConfig) *LogoFetcher {
	return &LogoFetcher{
		HTTPClients:        &CIMDHTTPClients{Strict: httpClient, Insecure: httpClient},
		OAuthFeatureConfig: &config.OAuthFeatureConfig{ClientIDMetadataDocument: featureConfig},
		AppID:              "test-app",
	}
}

func TestLogoFetcherFetch(t *testing.T) {
	Convey("LogoFetcher.Fetch", t, func() {
		ctx := context.Background()

		Convey("200, Content-Type: image/png, real PNG bytes: ok, returned type image/png", func() {
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				_, _ = w.Write(pngMagic)
			}))
			defer srv.Close()

			f := logoFetcherFor(newLoopbackHTTPClient(certPool(srv), nil), nil)
			body, contentType, err := f.Fetch(ctx, srv.URL)
			So(err, ShouldBeNil)
			So(body, ShouldResemble, pngMagic)
			So(contentType, ShouldEqual, "image/png")
		})

		Convey("200, image/png declared, real JPEG bytes: ok, returned type is the SNIFFED type", func() {
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				_, _ = w.Write(jpegMagic)
			}))
			defer srv.Close()

			f := logoFetcherFor(newLoopbackHTTPClient(certPool(srv), nil), nil)
			_, contentType, err := f.Fetch(ctx, srv.URL)
			So(err, ShouldBeNil)
			So(contentType, ShouldEqual, "image/jpeg")
		})

		Convey("200, image/png declared, HTML bytes: ErrLogoContentMismatch -- sniff rejects", func() {
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				_, _ = w.Write(htmlBody)
			}))
			defer srv.Close()

			f := logoFetcherFor(newLoopbackHTTPClient(certPool(srv), nil), nil)
			_, _, err := f.Fetch(ctx, srv.URL)
			So(errors.Is(err, ErrLogoContentMismatch), ShouldBeTrue)
		})

		Convey("200, image/svg+xml declared, SVG bytes: ErrLogoContentType -- SVG not allowed", func() {
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/svg+xml")
				_, _ = w.Write(svgBody)
			}))
			defer srv.Close()

			f := logoFetcherFor(newLoopbackHTTPClient(certPool(srv), nil), nil)
			_, _, err := f.Fetch(ctx, srv.URL)
			So(errors.Is(err, ErrLogoContentType), ShouldBeTrue)
		})

		Convey("200, text/html declared: ErrLogoContentType, before the body is read", func() {
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				_, _ = w.Write(htmlBody)
			}))
			defer srv.Close()

			f := logoFetcherFor(newLoopbackHTTPClient(certPool(srv), nil), nil)
			_, _, err := f.Fetch(ctx, srv.URL)
			So(errors.Is(err, ErrLogoContentType), ShouldBeTrue)
		})

		Convey("200, no explicit Content-Type: ErrLogoContentType -- net/http's own auto-sniffed type for plain text is not in the allowlist", func() {
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Deliberately not "Content-Type"-image-shaped bytes: if this
				// were pngMagic, net/http's own ResponseWriter would
				// auto-sniff and set Content-Type: image/png itself before
				// this test ever reaches LogoFetcher's declared-type check,
				// defeating the point of testing "no header".
				_, _ = w.Write([]byte("plain text body, no header set"))
			}))
			defer srv.Close()

			f := logoFetcherFor(newLoopbackHTTPClient(certPool(srv), nil), nil)
			_, _, err := f.Fetch(ctx, srv.URL)
			So(errors.Is(err, ErrLogoContentType), ShouldBeTrue)
		})

		Convey("Content-Type: image/png; charset=utf-8: ok, parameters stripped", func() {
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png; charset=utf-8")
				_, _ = w.Write(pngMagic)
			}))
			defer srv.Close()

			f := logoFetcherFor(newLoopbackHTTPClient(certPool(srv), nil), nil)
			_, contentType, err := f.Fetch(ctx, srv.URL)
			So(err, ShouldBeNil)
			So(contentType, ShouldEqual, "image/png")
		})

		Convey("exactly MaxLogoBytes is accepted", func() {
			exact := append(append([]byte{}, pngMagic...), bytes.Repeat([]byte("a"), MaxLogoBytes-len(pngMagic))...)
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				_, _ = w.Write(exact)
			}))
			defer srv.Close()

			f := logoFetcherFor(newLoopbackHTTPClient(certPool(srv), nil), nil)
			body, _, err := f.Fetch(ctx, srv.URL)
			So(err, ShouldBeNil)
			So(len(body), ShouldEqual, MaxLogoBytes)
		})

		Convey("MaxLogoBytes+1 is refused, via progressive enforcement rather than Content-Length", func() {
			tooLarge := append(append([]byte{}, pngMagic...), bytes.Repeat([]byte("a"), MaxLogoBytes+1-len(pngMagic))...)
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				_, _ = w.Write(tooLarge)
			}))
			defer srv.Close()

			f := logoFetcherFor(newLoopbackHTTPClient(certPool(srv), nil), nil)
			_, _, err := f.Fetch(ctx, srv.URL)
			So(errors.Is(err, ErrLogoTooLarge), ShouldBeTrue)
		})

		Convey("logo_uri is http://..., insecure_http_allowed: false: ErrLogoInvalidURI, no request made", func() {
			hits := 0
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits++
				w.Header().Set("Content-Type", "image/png")
				_, _ = w.Write(pngMagic)
			}))
			defer srv.Close()
			httpURL := "http://" + srv.Listener.Addr().String() + "/logo.png"

			f := logoFetcherFor(newLoopbackHTTPClient(certPool(srv), nil), &config.OAuthClientIDMetadataDocumentFeatureConfig{
				InsecureHTTPAllowed: boolPtr(false),
			})
			_, _, err := f.Fetch(ctx, httpURL)
			So(errors.Is(err, ErrLogoInvalidURI), ShouldBeTrue)
			So(hits, ShouldEqual, 0)
		})

		Convey("logo_uri resolves to 127.0.0.1, insecure_fetch_address_allowed: false: blocked", func() {
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				_, _ = w.Write(pngMagic)
			}))
			defer srv.Close()

			dialer := &SafeDialer{AllowNonPublicAddresses: false}
			transport := &http.Transport{
				DialContext:     dialer.DialContext,
				TLSClientConfig: &tls.Config{RootCAs: certPool(srv)},
			}
			client := httputil.NewExternalClientWithOptions(FetchTimeout, httputil.ExternalClientOptions{
				FollowRedirect: false,
				Transport:      transport,
			})
			f := logoFetcherFor(client, &config.OAuthClientIDMetadataDocumentFeatureConfig{
				InsecureFetchAddressAllowed: boolPtr(false),
			})
			_, _, err := f.Fetch(ctx, srv.URL)
			So(errors.Is(err, errBlockedAddress), ShouldBeTrue)
		})

		Convey("logo_uri resolves to 127.0.0.1, insecure_fetch_address_allowed: true: fetch proceeds", func() {
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				_, _ = w.Write(pngMagic)
			}))
			defer srv.Close()

			f := logoFetcherFor(newLoopbackHTTPClient(certPool(srv), nil), &config.OAuthClientIDMetadataDocumentFeatureConfig{
				InsecureFetchAddressAllowed: boolPtr(true),
			})
			_, _, err := f.Fetch(ctx, srv.URL)
			So(err, ShouldBeNil)
		})

		Convey("a 301 redirect to a valid image is refused (0 redirects followed) and the target is never requested", func() {
			targetHits := 0
			target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				targetHits++
				w.Header().Set("Content-Type", "image/png")
				_, _ = w.Write(pngMagic)
			}))
			defer target.Close()

			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, target.URL, http.StatusMovedPermanently)
			}))
			defer srv.Close()

			pool := certPool(srv)
			addCert(pool, target)
			f := logoFetcherFor(newLoopbackHTTPClient(pool, nil), nil)
			_, _, err := f.Fetch(ctx, srv.URL)
			So(errors.Is(err, ErrResponseNotOK), ShouldBeTrue)
			So(targetHits, ShouldEqual, 0)
		})
	})
}
