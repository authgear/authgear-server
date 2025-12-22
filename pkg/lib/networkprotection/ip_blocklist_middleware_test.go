package networkprotection

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestIPFilterMiddleware(t *testing.T) {
	usIP := "64.233.160.0"
	hkIP := "172.253.5.0"

	runTest := func(ip string, cfg *config.NetworkProtectionConfig, expectedStatus int) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)

		mw := &IPFilterMiddleware{
			RemoteIP: httputil.RemoteIP(ip),
			Config:   cfg,
		}
		// The default action in config must be explicitly set for predictable testing.
		// The real application parsing would call SetDefaults(), but here we construct the config manually.
		if mw.Config.IPFilter != nil && mw.Config.IPFilter.DefaultAction == "" {
			mw.Config.IPFilter.SetDefaults()
		}

		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		handler := mw.Handle(next)
		handler.ServeHTTP(rec, req)

		So(rec.Code, ShouldEqual, expectedStatus)
		So(called, ShouldEqual, expectedStatus == http.StatusOK)
	}

	Convey("IP filter middleware", t, func() {
		Convey("should allow by default when no rules match", func() {
			cfg := &config.NetworkProtectionConfig{
				IPFilter: &config.IPFilterConfig{
					DefaultAction: config.IPFilterActionAllow,
				},
			}
			runTest(usIP, cfg, http.StatusOK)
		})

		Convey("should deny by default when no rules match", func() {
			cfg := &config.NetworkProtectionConfig{
				IPFilter: &config.IPFilterConfig{
					DefaultAction: config.IPFilterActionDeny,
				},
			}
			runTest(usIP, cfg, http.StatusForbidden)
		})

		Convey("should allow by CIDR rule, overriding deny default", func() {
			cfg := &config.NetworkProtectionConfig{
				IPFilter: &config.IPFilterConfig{
					DefaultAction: config.IPFilterActionDeny,
					Rules: []*config.IPFilterRule{
						{
							Action: config.IPFilterActionAllow,
							Source: config.IPFilterSource{CIDRs: []string{usIP + "/32"}},
						},
					},
				},
			}
			runTest(usIP, cfg, http.StatusOK)
		})

		Convey("should deny by CIDR rule, overriding allow default", func() {
			cfg := &config.NetworkProtectionConfig{
				IPFilter: &config.IPFilterConfig{
					DefaultAction: config.IPFilterActionAllow,
					Rules: []*config.IPFilterRule{
						{
							Action: config.IPFilterActionDeny,
							Source: config.IPFilterSource{CIDRs: []string{usIP + "/32"}},
						},
					},
				},
			}
			runTest(usIP, cfg, http.StatusForbidden)
		})

		Convey("should allow by GeoIP rule, overriding deny default", func() {
			cfg := &config.NetworkProtectionConfig{
				IPFilter: &config.IPFilterConfig{
					DefaultAction: config.IPFilterActionDeny,
					Rules: []*config.IPFilterRule{
						{
							Action: config.IPFilterActionAllow,
							Source: config.IPFilterSource{GeoLocationCodes: []string{"HK"}},
						},
					},
				},
			}
			runTest(hkIP, cfg, http.StatusOK)
		})

		Convey("should deny by GeoIP rule, overriding allow default", func() {
			cfg := &config.NetworkProtectionConfig{
				IPFilter: &config.IPFilterConfig{
					DefaultAction: config.IPFilterActionAllow,
					Rules: []*config.IPFilterRule{
						{
							Action: config.IPFilterActionDeny,
							Source: config.IPFilterSource{GeoLocationCodes: []string{"HK"}},
						},
					},
				},
			}
			runTest(hkIP, cfg, http.StatusForbidden)
		})

		Convey("should match with OR logic for combined rules", func() {
			cfg := &config.NetworkProtectionConfig{
				IPFilter: &config.IPFilterConfig{
					DefaultAction: config.IPFilterActionDeny,
					Rules: []*config.IPFilterRule{
						{
							Action: config.IPFilterActionAllow,
							Source: config.IPFilterSource{
								CIDRs:            []string{"1.1.1.1/32"}, // Doesn't match
								GeoLocationCodes: []string{"HK"},         // Matches
							},
						},
					},
				},
			}
			// Should allow because GeoIP matches
			runTest(hkIP, cfg, http.StatusOK)
		})

		Convey("should fall back to default when no part of a rule matches", func() {
			cfg := &config.NetworkProtectionConfig{
				IPFilter: &config.IPFilterConfig{
					DefaultAction: config.IPFilterActionDeny,
					Rules: []*config.IPFilterRule{
						{
							Action: config.IPFilterActionAllow,
							Source: config.IPFilterSource{GeoLocationCodes: []string{"US"}},
						},
					},
				},
			}
			// IP from HK does not match rule, should be denied by default.
			runTest(hkIP, cfg, http.StatusForbidden)
		})

		Convey("should apply the first matched rule when multiple rules exist", func() {
			cfg := &config.NetworkProtectionConfig{
				IPFilter: &config.IPFilterConfig{
					DefaultAction: config.IPFilterActionDeny,
					Rules: []*config.IPFilterRule{
						{
							Action: config.IPFilterActionAllow,
							Source: config.IPFilterSource{CIDRs: []string{usIP + "/32"}},
						},
						{
							Action: config.IPFilterActionDeny,
							Source: config.IPFilterSource{CIDRs: []string{usIP + "/32"}},
						},
					},
				},
			}
			runTest(usIP, cfg, http.StatusOK)
		})

		Convey("should deny if the first matched rule is deny", func() {
			cfg := &config.NetworkProtectionConfig{
				IPFilter: &config.IPFilterConfig{
					DefaultAction: config.IPFilterActionAllow,
					Rules: []*config.IPFilterRule{
						{
							Action: config.IPFilterActionDeny,
							Source: config.IPFilterSource{CIDRs: []string{usIP + "/32"}},
						},
						{
							Action: config.IPFilterActionAllow,
							Source: config.IPFilterSource{CIDRs: []string{usIP + "/32"}},
						},
					},
				},
			}
			runTest(usIP, cfg, http.StatusForbidden)
		})

		Convey("should match a rule in the middle of the list", func() {
			cfg := &config.NetworkProtectionConfig{
				IPFilter: &config.IPFilterConfig{
					DefaultAction: config.IPFilterActionDeny,
					Rules: []*config.IPFilterRule{
						{
							Action: config.IPFilterActionAllow,
							Source: config.IPFilterSource{GeoLocationCodes: []string{"JP"}},
						},
						{
							Action: config.IPFilterActionAllow,
							Source: config.IPFilterSource{GeoLocationCodes: []string{"HK"}},
						},
						{
							Action: config.IPFilterActionDeny,
							Source: config.IPFilterSource{GeoLocationCodes: []string{"HK"}},
						},
					},
				},
			}
			runTest(hkIP, cfg, http.StatusOK)
		})

		Convey("should pass through for invalid IP", func() {
			cfg := &config.NetworkProtectionConfig{
				IPFilter: &config.IPFilterConfig{
					DefaultAction: config.IPFilterActionDeny,
				},
			}
			runTest("not-an-ip", cfg, http.StatusOK)
		})
	})
}
