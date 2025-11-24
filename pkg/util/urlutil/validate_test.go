package urlutil

import (
	"errors"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateHTTPSStrict(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError error
	}{
		{"Valid public domain", "https://valid-test.com/path?query=1", nil},
		{"Valid subdomain", "https://cdn.valid-test.com/image.png", nil},
		{"Valid HTTPS with port", "https://valid-test.com:443/path", nil},
		{"HTTP scheme rejected", "http://valid-test.com", ErrNotHTTPS},
		{"Single-label localhost rejected", "https://localhost/path", ErrSingleLabel},
		{"Single-label internal rejected", "https://internal/resource", ErrSingleLabel},
		{"IPv4 address rejected", "https://127.0.0.1/path", ErrIPNotAllowed},
		{"IPv6 address rejected", "https://[::1]/path", ErrIPNotAllowed},
		{"Metadata domain blocked", "https://metadata.google.internal/", ErrBlockedHost},
		{"Localdomain blocked", "https://example.localdomain/path", ErrBlockedHost},
		{"Userinfo in URL rejected", "https://user:pass@valid-test.com/path", ErrUserInfo},
		{"URL with fragment rejected", "https://valid-test.com/path#section", ErrFragment},
		{"URL with fragment and query rejected", "https://valid-test.com/path?x=1#frag", ErrFragment},
		{"URL with newline rejected", "https://valid-test.com/\npath", errors.New("parse \"https://valid-test.com/\\npath\": net/url: invalid control character in URL")},
		{"URL exceeds maximum length", "https://valid-test.com/" + strings.Repeat("a", MaxURLLength), ErrTooLong},
		{"Empty domain rejected", "  ", ErrBadHost},
		{"Leading hyphen in hostname rejected", "https://-bad.valid-test.com/path", ErrBadHost},
		{"Trailing hyphen in hostname rejected", "https://bad-.valid-test.com/path", ErrBadHost},
		{"Valid hyphen inside label accepted", "https://good-label.valid-test.com/path", nil},
		{"Double dot in hostname rejected", "https://a..b.valid-test.com/path", ErrBadHost},
		{"Leading dot in hostname rejected", "https://.valid-test.com/path", ErrBadHost},
		{"Trailing dot in hostname accepted", "https://valid-test.com./path", nil},
		{"Underscore in hostname rejected", "https://invalid_name.com/path", ErrBadHost},
		{"Valid punycode domain accepted", "https://xn--dummydomain-pun.com/path", nil},
		{"Single-label punycode rejected", "https://xn--dummydomain-pun/", ErrSingleLabel},
		{"Multiple valid subdomains accepted", "https://a.b.c.d.valid-test.com/path", nil},
		{"Hostname with digits only accepted", "https://123.valid-test.com/path", nil},
		{"Label starting with digit accepted", "https://3dexample.valid-test.com/path", nil},
		{"Non-ASCII hostname rejected", "https://éxample.com/path", ErrBadHost},
	}

	Convey("TestValidateHTTPSStrict", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				err := ValidateHTTPSStrict(tt.input)
				if tt.expectError != nil {
					So(err, ShouldBeError, tt.expectError)
				} else {
					So(err, ShouldBeNil)
				}
			})
		}
	})
}
