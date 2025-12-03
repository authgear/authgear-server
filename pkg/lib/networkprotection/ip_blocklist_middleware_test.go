package networkprotection

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestIPBlocklistMiddleware(t *testing.T) {
	hkGoogleIP := "172.253.5.0"

	Convey("IP inside blocklist CIDR should be forbidden", t, func() {
		cfg := &config.NetworkProtectionConfig{
			IPBlocklist: &config.NetworkIPBlocklistConfig{
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
		cfg := &config.NetworkProtectionConfig{
			IPBlocklist: &config.NetworkIPBlocklistConfig{
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

	Convey("Country code matching should be forbidden (case-insensitive)", t, func() {
		cfg := &config.NetworkProtectionConfig{
			IPBlocklist: &config.NetworkIPBlocklistConfig{
				CountryCodes: []string{"HK"},
			},
		}

		mw := &IPBlocklistMiddleware{
			RemoteIP: httputil.RemoteIP(hkGoogleIP),
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

	Convey("Country code non-matching should pass through", t, func() {
		cfg := &config.NetworkProtectionConfig{
			IPBlocklist: &config.NetworkIPBlocklistConfig{
				CountryCodes: []string{"US"},
			},
		}

		mw := &IPBlocklistMiddleware{
			RemoteIP: httputil.RemoteIP(hkGoogleIP),
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

	Convey("Invalid IP should skip geoip lookup and pass through", t, func() {
		cfg := &config.NetworkProtectionConfig{
			IPBlocklist: &config.NetworkIPBlocklistConfig{
				CountryCodes: []string{"US"},
			},
		}

		mw := &IPBlocklistMiddleware{
			RemoteIP: httputil.RemoteIP("not-an-ip"),
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
