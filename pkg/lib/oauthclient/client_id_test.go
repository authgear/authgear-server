package oauthclient_test

import (
	"net/url"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
)

func TestParseCIMDClientID(t *testing.T) {
	Convey("ParseCIMDClientID with allowInsecureHTTP=false", t, func() {
		cases := []struct {
			name    string
			input   string
			wantErr error
			wantOK  bool
		}{
			{"valid document URL", "https://mcp-client.example.com/oauth/client-metadata.json", nil, true},
			{"query string allowed (D3)", "https://mcp-client.example.com/a?b=c", nil, true},
			{"scheme case-insensitive", "HTTPS://Example.com/x", nil, true},
			{"http rejected", "http://example.com/x", oauthclient.ErrCIMDClientIDNotHTTPS, false},
			{"empty path rejected", "https://example.com", oauthclient.ErrCIMDClientIDNoPath, false},
			{"bare-root path allowed (D2)", "https://example.com/", nil, true},
			{"fragment rejected", "https://example.com/x#frag", oauthclient.ErrCIMDClientIDHasFragment, false},
			{"bare fragment rejected", "https://example.com/x#", oauthclient.ErrCIMDClientIDHasFragment, false},
			{"userinfo rejected", "https://user:pw@example.com/x", oauthclient.ErrCIMDClientIDHasUserInfo, false},
			{"dot-dot segment rejected", "https://example.com/a/../b", oauthclient.ErrCIMDClientIDHasDotSegment, false},
			{"dot segment rejected", "https://example.com/a/./b", oauthclient.ErrCIMDClientIDHasDotSegment, false},
			{"percent-encoded dot-dot rejected", "https://example.com/a/%2e%2e/b", oauthclient.ErrCIMDClientIDHasDotSegment, false},
			{"empty host rejected", "https:///x", oauthclient.ErrCIMDClientIDBadHost, false},
			{"too long rejected", "https://example.com/" + strings.Repeat("a", 2001), oauthclient.ErrCIMDClientIDTooLong, false},
			{"dcrc id is not a CIMD shape", "dcrc_abc", nil, false},
			{"static client id is not a CIMD shape", "my-static-client", nil, false},
			{"empty string is not a CIMD shape", "", nil, false},
			{"loopback ip host allowed (D4: no address policy here)", "https://127.0.0.1/x", nil, true},
			{"localhost host allowed (D4: no address policy here)", "https://localhost/x", nil, true},
		}

		for _, tc := range cases {
			Convey(tc.name, func() {
				u, err := oauthclient.ParseCIMDClientID(tc.input, false)
				if tc.wantOK {
					So(err, ShouldBeNil)
					So(u, ShouldNotBeNil)
				} else if tc.wantErr != nil {
					So(err, ShouldEqual, tc.wantErr)
				} else {
					So(err, ShouldBeError)
				}
				So(oauthclient.IsCIMDClientID(tc.input, false), ShouldEqual, tc.wantOK)
			})
		}

		Convey("host casing is preserved (only the scheme comparison is case-insensitive)", func() {
			u, err := oauthclient.ParseCIMDClientID("HTTPS://Example.com/x", false)
			So(err, ShouldBeNil)
			So(u.Host, ShouldEqual, "Example.com")
		})
	})

	Convey("ParseCIMDClientID: allowInsecureHTTP widens only the scheme rule", t, func() {
		cases := []struct {
			name        string
			input       string
			strictErr   error
			relaxedOK   bool
			relaxedErr  error
			checkStrict bool
		}{
			{"http rejected strict, allowed relaxed", "http://example.com/x", oauthclient.ErrCIMDClientIDNotHTTPS, true, nil, true},
			{"http loopback with port", "http://localhost:2727/x.json", oauthclient.ErrCIMDClientIDNotHTTPS, true, nil, true},
			{"http private address (D13: scheme rule is not the host's business)", "http://10.0.0.5:2727/x.json", oauthclient.ErrCIMDClientIDNotHTTPS, true, nil, true},
			{"http bare-root path (D2)", "http://example.com/", oauthclient.ErrCIMDClientIDNotHTTPS, true, nil, true},
			{"https still fine relaxed", "https://example.com/x", nil, true, nil, true},
			{"ftp scheme still rejected relaxed", "ftp://example.com/x", nil, false, nil, false},
			{"scheme-relative still rejected relaxed", "//example.com/x", nil, false, nil, false},
			{"no scheme still rejected relaxed", "example.com/x", nil, false, nil, false},
			{"empty path still rejected relaxed", "http://example.com", nil, false, oauthclient.ErrCIMDClientIDNoPath, false},
			{"userinfo still rejected relaxed", "http://user:pw@example.com/x", nil, false, oauthclient.ErrCIMDClientIDHasUserInfo, false},
		}

		for _, tc := range cases {
			Convey(tc.name, func() {
				if tc.checkStrict {
					_, err := oauthclient.ParseCIMDClientID(tc.input, false)
					So(err, ShouldEqual, tc.strictErr)
				}

				u, err := oauthclient.ParseCIMDClientID(tc.input, true)
				if tc.relaxedOK {
					So(err, ShouldBeNil)
					So(u, ShouldNotBeNil)
				} else if tc.relaxedErr != nil {
					So(err, ShouldEqual, tc.relaxedErr)
				} else {
					So(err, ShouldBeError)
				}
			})
		}
	})
}

func TestIsCIMDClientIDAllowed(t *testing.T) {
	mustParse := func(raw string) *url.URL {
		u, err := oauthclient.ParseCIMDClientID(raw, false)
		if err != nil {
			panic(err)
		}
		return u
	}

	Convey("IsCIMDClientIDAllowed", t, func() {
		cases := []struct {
			name     string
			patterns []string
			host     string
			expect   bool
		}{
			{"empty allowlist allows everything", nil, "example.com", true},
			{"exact match", []string{"example.com"}, "example.com", true},
			{"exact pattern does not match a subdomain", []string{"example.com"}, "a.example.com", false},
			{"wildcard matches one label", []string{"*.example.com"}, "a.example.com", true},
			{"wildcard does not match two labels (D6)", []string{"*.example.com"}, "a.b.example.com", false},
			{"wildcard does not match apex", []string{"*.example.com"}, "example.com", false},
			{"wildcard does not match empty label", []string{"*.example.com"}, ".example.com", false},
			{"wildcard requires a label boundary", []string{"*.example.com"}, "xexample.com", false},
			{"apex and wildcard both listed: apex host allowed", []string{"example.com", "*.example.com"}, "example.com", true},
			{"apex and wildcard both listed: subdomain host allowed", []string{"example.com", "*.example.com"}, "a.example.com", true},
			{"case-insensitive on both sides", []string{"EXAMPLE.com"}, "example.COM", true},
			{"trailing dot on host is stripped", []string{"example.com"}, "example.com.", true},
			{"single-label pattern allowed (D15)", []string{"localhost"}, "localhost", true},
			{"no match", []string{"example.com"}, "other.test", false},
		}

		for _, tc := range cases {
			Convey(tc.name, func() {
				u := mustParse("https://" + tc.host + "/x")
				So(oauthclient.IsCIMDClientIDAllowed(tc.patterns, u), ShouldEqual, tc.expect)
			})
		}
	})
}
