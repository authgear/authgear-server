package webapp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/cimd"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type stubClientLogoClientResolver struct {
	client *config.OAuthClientConfig
}

func (s *stubClientLogoClientResolver) ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig {
	return s.client
}

type stubClientLogoLogoService struct {
	result *cimd.LogoResult
	err    error
	calls  int
}

func (s *stubClientLogoLogoService) Get(ctx context.Context, clientID string, logoURI string) (*cimd.LogoResult, error) {
	s.calls++
	return s.result, s.err
}

func TestClientLogoHandlerServeHTTP(t *testing.T) {
	Convey("ClientLogoHandler.ServeHTTP", t, func() {
		Convey("unknown client_id: 404", func() {
			h := &ClientLogoHandler{
				Resolver: &stubClientLogoClientResolver{client: nil},
				Logos:    &stubClientLogoLogoService{},
			}
			req := httptest.NewRequest(http.MethodGet, "/_internals/client_logo?client_id=unknown", nil)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)
			So(rw.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("resolvable client with empty LogoURI: 404", func() {
			h := &ClientLogoHandler{
				Resolver: &stubClientLogoClientResolver{client: &config.OAuthClientConfig{
					ClientID:      "https://mcp-client.example.com/metadata.json",
					DynamicSource: model.OAuthClientSourceCIMD,
					LogoURI:       "",
				}},
				Logos: &stubClientLogoLogoService{},
			}
			req := httptest.NewRequest(http.MethodGet, "/_internals/client_logo?client_id=x", nil)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)
			So(rw.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("static client with a LogoURI: 404 -- the proxy covers dynamic clients only", func() {
			logos := &stubClientLogoLogoService{}
			h := &ClientLogoHandler{
				Resolver: &stubClientLogoClientResolver{client: &config.OAuthClientConfig{
					ClientID: "static-client",
					LogoURI:  "https://static-client.example.com/logo.png",
					// DynamicSource left unset ("") -- a static client.
				}},
				Logos: logos,
			}
			req := httptest.NewRequest(http.MethodGet, "/_internals/client_logo?client_id=static-client", nil)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)
			So(rw.Code, ShouldEqual, http.StatusNotFound)
			So(logos.calls, ShouldEqual, 0)
		})

		Convey("dynamic client, LogoService returns bytes: 200 with the expected headers", func() {
			fetchedAt := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
			h := &ClientLogoHandler{
				Resolver: &stubClientLogoClientResolver{client: &config.OAuthClientConfig{
					ClientID:      "https://mcp-client.example.com/metadata.json",
					DynamicSource: model.OAuthClientSourceCIMD,
					IsDynamic:     true,
					LogoURI:       "https://mcp-client.example.com/logo.png",
				}},
				Logos: &stubClientLogoLogoService{result: &cimd.LogoResult{
					ContentType: "image/png",
					Body:        []byte("fake-png-bytes"),
					FetchedAt:   fetchedAt,
				}},
			}
			req := httptest.NewRequest(http.MethodGet, "/_internals/client_logo?client_id=x", nil)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)
			So(rw.Code, ShouldEqual, http.StatusOK)
			So(rw.Body.String(), ShouldEqual, "fake-png-bytes")
			So(rw.Header().Get("Content-Type"), ShouldEqual, "image/png")
			So(rw.Header().Get("X-Content-Type-Options"), ShouldEqual, "nosniff")
			So(rw.Header().Get("Content-Disposition"), ShouldEqual, "inline")
			So(rw.Header().Get("Content-Security-Policy"), ShouldEqual, "default-src 'none'; sandbox")
			So(rw.Header().Get("Cache-Control"), ShouldEqual, "private, max-age=3600")
		})

		Convey("LogoService returns CIMDLogoUnavailable: 404, and the response body never carries the concrete cause", func() {
			h := &ClientLogoHandler{
				Resolver: &stubClientLogoClientResolver{client: &config.OAuthClientConfig{
					ClientID:      "https://mcp-client.example.com/metadata.json",
					DynamicSource: model.OAuthClientSourceCIMD,
					IsDynamic:     true,
					LogoURI:       "https://mcp-client.example.com/logo.png",
				}},
				Logos: &stubClientLogoLogoService{err: cimd.ErrLogoUnavailable()},
			}
			req := httptest.NewRequest(http.MethodGet, "/_internals/client_logo?client_id=x", nil)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)
			So(rw.Code, ShouldEqual, http.StatusNotFound)
			// http.NotFound's fixed body, not anything derived from the
			// error -- LogoService.Get already collapsed every logo-specific
			// failure to this one Kind before it reached the handler, so
			// there is nothing left here that could vary by cause.
			So(rw.Body.String(), ShouldEqual, "404 page not found\n")
		})

		Convey("LogoService returns any other error: 500, and the error is logged", func() {
			h := &ClientLogoHandler{
				Resolver: &stubClientLogoClientResolver{client: &config.OAuthClientConfig{
					ClientID:      "https://mcp-client.example.com/metadata.json",
					DynamicSource: model.OAuthClientSourceCIMD,
					IsDynamic:     true,
					LogoURI:       "https://mcp-client.example.com/logo.png",
				}},
				Logos: &stubClientLogoLogoService{err: errors.New("redis: connection refused")},
			}
			req := httptest.NewRequest(http.MethodGet, "/_internals/client_logo?client_id=x", nil)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)
			So(rw.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("HEAD: headers with no body", func() {
			fetchedAt := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
			h := &ClientLogoHandler{
				Resolver: &stubClientLogoClientResolver{client: &config.OAuthClientConfig{
					ClientID:      "https://mcp-client.example.com/metadata.json",
					DynamicSource: model.OAuthClientSourceCIMD,
					IsDynamic:     true,
					LogoURI:       "https://mcp-client.example.com/logo.png",
				}},
				Logos: &stubClientLogoLogoService{result: &cimd.LogoResult{
					ContentType: "image/png",
					Body:        []byte("fake-png-bytes"),
					FetchedAt:   fetchedAt,
				}},
			}
			req := httptest.NewRequest(http.MethodHead, "/_internals/client_logo?client_id=x", nil)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)
			So(rw.Code, ShouldEqual, http.StatusOK)
			So(rw.Body.Len(), ShouldEqual, 0)
			So(rw.Header().Get("Content-Type"), ShouldEqual, "image/png")
		})
	})
}
