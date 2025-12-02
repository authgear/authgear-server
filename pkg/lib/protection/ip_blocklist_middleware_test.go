package protection

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIPBlocklistMiddleware(t *testing.T) {
	Convey("IP inside blocklist CIDR should be forbidden", t, func() {
		cfg := &config.ProtectionConfig{
			IPBlocklist: &config.IPBlocklistConfig{
				CIDRs: []string{"203.0.113.0/24"},
			},
		}

		mw := &IPBlocklistMiddleware{
			RemoteIP: httputil.RemoteIP("203.0.113.5"),
			Config:   cfg,
		}

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)

		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		handler := mw.Handle(next)
		handler.ServeHTTP(rec, req)

		So(rec.Code, ShouldEqual, http.StatusForbidden)
		So(called, ShouldBeFalse)
	})

	Convey("IP outside blocklist CIDR should pass through", t, func() {
		cfg := &config.ProtectionConfig{
			IPBlocklist: &config.IPBlocklistConfig{
				CIDRs: []string{"203.0.113.0/24"},
			},
		}

		mw := &IPBlocklistMiddleware{
			RemoteIP: httputil.RemoteIP("198.51.100.10"),
			Config:   cfg,
		}

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)

		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		handler := mw.Handle(next)
		handler.ServeHTTP(rec, req)

		So(rec.Code, ShouldEqual, http.StatusOK)
		So(called, ShouldBeTrue)
	})
}
