package validation

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func backgroundCtx() context.Context {
	return context.Background()
}

func TestFormatEmail(t *testing.T) {
	Convey("FormatEmail", t, func() {
		f := FormatEmail{}.CheckFormat

		So(f(backgroundCtx(), 1), ShouldBeNil)
		So(f(backgroundCtx(), "user@example.com"), ShouldBeNil)
		So(f(backgroundCtx(), ""), ShouldBeError, "mail: no address")
		So(f(backgroundCtx(), "John Doe <user@example.com>"), ShouldBeError, "input email must not have name")
		So(f(backgroundCtx(), "foobar"), ShouldBeError, "mail: missing '@' or angle-addr")
	})

	Convey("FormatEmail with name", t, func() {
		f := FormatEmail{AllowName: true}.CheckFormat

		So(f(backgroundCtx(), 1), ShouldBeNil)
		So(f(backgroundCtx(), "user@example.com"), ShouldBeNil)
		So(f(backgroundCtx(), "John Doe <user@example.com>"), ShouldBeNil)
		So(f(backgroundCtx(), ""), ShouldBeError, "mail: no address")
		So(f(backgroundCtx(), "foobar"), ShouldBeError, "mail: missing '@' or angle-addr")
	})
}

func TestFormatURI(t *testing.T) {
	Convey("FormatURI", t, func() {
		f := FormatURI{}.CheckFormat

		So(f(backgroundCtx(), 1), ShouldBeNil)
		So(f(backgroundCtx(), ""), ShouldBeError, "input URL must be absolute")
		So(f(backgroundCtx(), "/a"), ShouldBeError, "input URL must be absolute")
		So(f(backgroundCtx(), "#a"), ShouldBeError, "input URL must be absolute")
		So(f(backgroundCtx(), "?a"), ShouldBeError, "input URL must be absolute")

		So(f(backgroundCtx(), "http://example.com/../.."), ShouldBeError, "invalid path: /../..")

		So(f(backgroundCtx(), "http://example.com"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com/"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com/a"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com/a/"), ShouldBeNil)

		So(f(backgroundCtx(), "http://example.com?"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com?a"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com?a=b"), ShouldBeNil)

		So(f(backgroundCtx(), "http://example.com#"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com#a"), ShouldBeNil)
	})
}

func TestFormatPicture(t *testing.T) {
	Convey("FormatPicture", t, func() {
		f := FormatPicture{}.CheckFormat

		So(f(backgroundCtx(), 1), ShouldBeNil)
		So(f(backgroundCtx(), ""), ShouldBeError, "invalid scheme: ")
		So(f(backgroundCtx(), "foobar:"), ShouldBeError, "invalid scheme: foobar")

		So(f(backgroundCtx(), "http://example.com"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com/"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com/a"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com/a/"), ShouldBeNil)

		So(f(backgroundCtx(), "http://example.com?"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com?a"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com?a=b"), ShouldBeNil)

		So(f(backgroundCtx(), "http://example.com#"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com#a"), ShouldBeNil)

		So(f(backgroundCtx(), "authgearimages:///app/object"), ShouldBeNil)
		So(f(backgroundCtx(), "authgearimages:///../app/object"), ShouldBeError, "invalid path: /../app/object")
		So(f(backgroundCtx(), "authgearimages://host/"), ShouldBeError, "authgearimages URI does not have host")
	})
}

func TestFormatHTTPOrigin(t *testing.T) {
	Convey("FormatHTTPOrigin", t, func() {
		f := FormatHTTPOrigin{}.CheckFormat

		So(f(backgroundCtx(), 1), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com?"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com#"), ShouldBeNil)

		So(f(backgroundCtx(), ""), ShouldBeError, "expect input URL with scheme http / https")
		So(f(backgroundCtx(), "http://user:password@example.com"), ShouldBeError, "expect input URL without user info, path, query and fragment")
		So(f(backgroundCtx(), "http://example.com/"), ShouldBeError, "expect input URL without user info, path, query and fragment")
		So(f(backgroundCtx(), "http://example.com?a"), ShouldBeError, "expect input URL without user info, path, query and fragment")
		So(f(backgroundCtx(), "http://example.com#a"), ShouldBeError, "expect input URL without user info, path, query and fragment")
	})
}

func TestFormatHTTPOriginSpec(t *testing.T) {
	Convey("FormatHTTPOriginSpec", t, func() {
		f := FormatHTTPOriginSpec{}.CheckFormat

		So(f(backgroundCtx(), 1), ShouldBeNil)
		So(f(backgroundCtx(), ""), ShouldBeNil)
		So(f(backgroundCtx(), "127.0.0.1"), ShouldBeNil)
		So(f(backgroundCtx(), "127.0.0.1/"), ShouldBeError, "127.0.0.1/ is not strict")
	})
}

func TestFormatLDAPURL(t *testing.T) {
	Convey("FormatLDAPURL", t, func() {
		f := FormatLDAPURL{}.CheckFormat

		So(f(backgroundCtx(), 1), ShouldBeNil)
		So(f(backgroundCtx(), "ldap://example.com"), ShouldBeNil)
		So(f(backgroundCtx(), "ldap://localhost:389"), ShouldBeNil)
		So(f(backgroundCtx(), "ldaps://example.com"), ShouldBeNil)

		So(f(backgroundCtx(), ""), ShouldBeError, "expect input URL with scheme ldap / ldaps")
		So(f(backgroundCtx(), "http://example.com"), ShouldBeError, "expect input URL with scheme ldap / ldaps")
		So(f(backgroundCtx(), "ldap://user:password@example.com"), ShouldBeError, "expect input URL without user info, path, query and fragment")
		So(f(backgroundCtx(), "ldap://example.com/"), ShouldBeError, "expect input URL without user info, path, query and fragment")
		So(f(backgroundCtx(), "ldap://example.com?a"), ShouldBeError, "expect input URL without user info, path, query and fragment")
		So(f(backgroundCtx(), "ldap://example.com#a"), ShouldBeError, "expect input URL without user info, path, query and fragment")
	})
}

func TestFormatLDAPDN(t *testing.T) {
	Convey("FormatLDAPDN", t, func() {
		f := FormatLDAPDN{}.CheckFormat

		So(f(backgroundCtx(), 1), ShouldBeNil)
		So(f(backgroundCtx(), "dc=example,dc=com"), ShouldBeNil)
		So(f(backgroundCtx(), "cn=admin,dc=example,dc=org"), ShouldBeNil)
		// AttributeValue can be empty according to the spec.
		// https://datatracker.ietf.org/doc/html/rfc4514#section-3
		// go-ldap v3.4.5 used to return parse error in case of empty value.
		// That behavior was changed since v3.4.8
		So(f(backgroundCtx(), "ou="), ShouldBeNil)
		So(f(backgroundCtx(), "asbbalskjedkbwk"), ShouldBeError, "invalid DN")
		So(f(backgroundCtx(), ""), ShouldBeError, "expect non-empty base DN")
	})
}

func TestFormatLDAPSearchFilterTemplate(t *testing.T) {
	Convey("FormatLDAPSearchFilterTemplate", t, func() {
		f := FormatLDAPSearchFilterTemplate{}.CheckFormat

		So(f(backgroundCtx(), 1), ShouldBeNil)
		So(f(backgroundCtx(), "(uid=%s)"), ShouldBeNil)
		So(f(backgroundCtx(), "((uid=%s)(cn=%s))"), ShouldBeError, "invalid search filter")
		So(f(backgroundCtx(), "(uid=%s)(cn=%s"), ShouldBeError, "invalid search filter")
		So(f(backgroundCtx(), ""), ShouldBeError, "invalid search filter")

		correct_template := `{{if eq $.Username "test@test.com"}}(mail={{$.Username}}){{else if eq $.Username "+852"}}(telephoneNumber={{$.Username}}){{else}}(uid={{$.Username}}){{end}}`
		So(f(backgroundCtx(), correct_template), ShouldBeNil)

		multiline_template := `
		{{if eq .Username "test@test.com"}}
									(mail={{$.Username}})
		{{else }}
			(uid={{$.Username}})
		{{end}}

		`
		So(f(backgroundCtx(), multiline_template), ShouldBeNil)

		missing_end := `{{if eq .Username "test@test.com"}}(mail={{.Username}})`
		So(f(backgroundCtx(), missing_end), ShouldBeError, "invalid template")

		missing_parenthesis := `{{if eq .Username "test@test.com"}}(mail={{.Username}}{{end}}`
		So(f(backgroundCtx(), missing_parenthesis), ShouldBeError, "invalid search filter")
	})
}

func TestFormatLDAPAttribute(t *testing.T) {
	Convey("FormatLDAPAttribute", t, func() {
		f := FormatLDAPAttribute{}.CheckFormat

		So(f(backgroundCtx(), 1), ShouldBeNil)
		So(f(backgroundCtx(), "uid"), ShouldBeNil)
		So(f(backgroundCtx(), "cn"), ShouldBeNil)
		So(f(backgroundCtx(), "ou"), ShouldBeNil)
		So(f(backgroundCtx(), "dc"), ShouldBeNil)
		So(f(backgroundCtx(), "hyphen-"), ShouldBeNil)
		So(f(backgroundCtx(), "abcdEFG123-"), ShouldBeNil)
		So(f(backgroundCtx(), ""), ShouldBeError, "expect non-empty attribute")
		So(f(backgroundCtx(), "-abc"), ShouldBeError, "invalid attribute")
		So(f(backgroundCtx(), "123"), ShouldBeError, "invalid attribute")
		So(f(backgroundCtx(), "should have no space"), ShouldBeError, "invalid attribute")
		So(f(backgroundCtx(), "abcd!@#$%^&*()"), ShouldBeError, "invalid attribute")
		So(f(backgroundCtx(), "dc="), ShouldBeError, "invalid attribute")
		So(f(backgroundCtx(), "dc=example,dc=com"), ShouldBeError, "invalid attribute")

	})
}

func TestFormatWeChatAccountID(t *testing.T) {
	Convey("TestFormatWeChatAccountID", t, func() {
		f := FormatWeChatAccountID{}.CheckFormat

		So(f(backgroundCtx(), 1), ShouldBeNil)
		So(f(backgroundCtx(), "foobar"), ShouldBeError, "expect WeChat account id start with gh_")
		So(f(backgroundCtx(), "gh_foobar"), ShouldBeNil)
	})
}

func TestFormatBCP47(t *testing.T) {
	f := FormatBCP47{}.CheckFormat

	Convey("FormatBCP47", t, func() {
		So(f(backgroundCtx(), 1), ShouldBeNil)

		So(f(backgroundCtx(), ""), ShouldBeError, "invalid BCP 47 tag: language: tag is not well-formed")
		So(f(backgroundCtx(), "a"), ShouldBeError, "invalid BCP 47 tag: language: tag is not well-formed")
		So(f(backgroundCtx(), "foobar"), ShouldBeError, "invalid BCP 47 tag: language: tag is not well-formed")

		So(f(backgroundCtx(), "en"), ShouldBeNil)
		So(f(backgroundCtx(), "zh-TW"), ShouldBeNil)
		So(f(backgroundCtx(), "und"), ShouldBeNil)

		So(f(backgroundCtx(), "zh_TW"), ShouldBeError, "non-canonical BCP 47 tag: zh_TW != zh-TW")
	})
}

func TestFormatTimezone(t *testing.T) {
	f := FormatTimezone{}.CheckFormat

	Convey("FormatTimezone", t, func() {
		So(f(backgroundCtx(), 1), ShouldBeNil)

		So(f(backgroundCtx(), ""), ShouldBeError, `valid timezone name has at least 1 slash: ""`)
		So(f(backgroundCtx(), "UTC"), ShouldBeError, `valid timezone name has at least 1 slash: "UTC"`)

		So(f(backgroundCtx(), "Asia/Hong_Kong"), ShouldBeNil)
		// It seems that support for uppercase name is machine specific.
		// So(f(backgroundCtx(),"ASIA/HONG_KONG"), ShouldBeNil)
		So(f(backgroundCtx(), "Etc/UTC"), ShouldBeNil)
	})
}

func TestFormatBirthdate(t *testing.T) {
	f := FormatBirthdate{}.CheckFormat

	Convey("FormatBirthdate", t, func() {
		So(f(backgroundCtx(), 1), ShouldBeNil)

		So(f(backgroundCtx(), ""), ShouldBeError, `invalid birthdate: ""`)
		So(f(backgroundCtx(), "foobar"), ShouldBeError, `invalid birthdate: "foobar"`)

		So(f(backgroundCtx(), "2021"), ShouldBeNil)
		So(f(backgroundCtx(), "--01-01"), ShouldBeNil)
		So(f(backgroundCtx(), "0000-01-01"), ShouldBeNil)
		So(f(backgroundCtx(), "2021-01-01"), ShouldBeNil)
	})
}

func TestFormatAlpha2(t *testing.T) {
	f := FormatAlpha2{}.CheckFormat

	Convey("FormatAlpha2", t, func() {
		So(f(backgroundCtx(), 1), ShouldBeNil)

		So(f(backgroundCtx(), ""), ShouldBeError, `invalid ISO 3166-1 alpha-2 code: ""`)
		So(f(backgroundCtx(), "foobar"), ShouldBeError, `invalid ISO 3166-1 alpha-2 code: "foobar"`)

		So(f(backgroundCtx(), "US"), ShouldBeNil)
		So(f(backgroundCtx(), "HK"), ShouldBeNil)
		So(f(backgroundCtx(), "CN"), ShouldBeNil)
		So(f(backgroundCtx(), "TW"), ShouldBeNil)
		So(f(backgroundCtx(), "JP"), ShouldBeNil)
	})
}

func TestFormatCustomAttributePointer(t *testing.T) {
	f := FormatCustomAttributePointer{}.CheckFormat

	Convey("FormatCustomAttributePointer", t, func() {
		So(f(backgroundCtx(), 1), ShouldBeNil)

		So(f(backgroundCtx(), ""), ShouldBeError, "custom attribute pointer must be one-level but found 0")
		So(f(backgroundCtx(), "/"), ShouldBeError, "custom attribute pointer must not be empty")
		So(f(backgroundCtx(), "/0"), ShouldBeError, `invalid character at 0: "0"`)
		So(f(backgroundCtx(), "/_"), ShouldBeError, `invalid character at 0: "_"`)
		So(f(backgroundCtx(), "/a_"), ShouldBeError, `invalid character at 1: "_"`)
		So(f(backgroundCtx(), "/a-a"), ShouldBeError, `invalid character at 1: "-"`)
		So(f(backgroundCtx(), "/aあa"), ShouldBeError, `invalid character at 1: "あ"`)

		So(f(backgroundCtx(), "/a0"), ShouldBeNil)
		So(f(backgroundCtx(), "/aa"), ShouldBeNil)
		So(f(backgroundCtx(), "/aaa"), ShouldBeNil)
		So(f(backgroundCtx(), "/a_a"), ShouldBeNil)
		So(f(backgroundCtx(), "/mystring"), ShouldBeNil)
		So(f(backgroundCtx(), "/my_string"), ShouldBeNil)
	})
}

func TestFormatGoogleTagManagerContainerID(t *testing.T) {
	f := FormatGoogleTagManagerContainerID{}.CheckFormat

	Convey("FormatGoogleTagManagerContainerID", t, func() {
		So(f(backgroundCtx(), 1), ShouldBeNil)

		So(f(backgroundCtx(), ""), ShouldBeError, "expect google tag manager container ID to start with GTM-")
		So(f(backgroundCtx(), "GTM-AAAAAA"), ShouldBeNil)
	})
}

func TestFormatHookURI(t *testing.T) {
	Convey("FormatHookURI", t, func() {
		f := FormatHookURI{}.CheckFormat

		So(f(backgroundCtx(), 1), ShouldBeNil)
		So(f(backgroundCtx(), ""), ShouldBeError, "invalid scheme: ")
		So(f(backgroundCtx(), "foobar:"), ShouldBeError, "invalid scheme: foobar")

		So(f(backgroundCtx(), "http://example.com"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com/"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com/a"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com/a/"), ShouldBeNil)

		So(f(backgroundCtx(), "http://example.com?"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com?a"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com?a=b"), ShouldBeNil)

		So(f(backgroundCtx(), "http://example.com#"), ShouldBeNil)
		So(f(backgroundCtx(), "http://example.com#a"), ShouldBeNil)

		So(f(backgroundCtx(), "authgeardeno:///deno/a.ts"), ShouldBeNil)
		So(f(backgroundCtx(), "authgeardeno:///../deno/a.ts"), ShouldBeError, "invalid path: /../deno/a.ts")
		So(f(backgroundCtx(), "authgeardeno://host/"), ShouldBeError, "authgeardeno URI does not have host")
	})
}

func TestFormatDurationString(t *testing.T) {
	f := FormatDurationString{}.CheckFormat

	Convey("FormatDurationString", t, func() {
		So(f(backgroundCtx(), 1), ShouldBeNil)

		So(f(backgroundCtx(), ""), ShouldBeError, `time: invalid duration ""`)
		So(f(backgroundCtx(), "foobar"), ShouldBeError, `time: invalid duration "foobar"`)

		So(f(backgroundCtx(), "0s"), ShouldBeError, `non-positive duration "0s"`)
		So(f(backgroundCtx(), "1.1s"), ShouldBeNil)
		So(f(backgroundCtx(), "2m3s"), ShouldBeNil)
		So(f(backgroundCtx(), "4h5m6s"), ShouldBeNil)
	})
}

func TestFormatBase64URL(t *testing.T) {
	f := FormatBase64URL{}.CheckFormat

	Convey("FormatBase64URL", t, func() {
		So(f(backgroundCtx(), 1), ShouldBeNil)
		So(f(backgroundCtx(), ""), ShouldBeNil)

		So(f(backgroundCtx(), "aA"), ShouldBeNil)
		So(f(backgroundCtx(), "aA=="), ShouldBeError, "illegal base64 data at input byte 2")
		So(f(backgroundCtx(), "a\nA"), ShouldBeNil)
	})
}

func TestFormatDateTime(t *testing.T) {
	f := FormatDateTime{}.CheckFormat

	Convey("TestFormatDateTime", t, func() {
		So(f(backgroundCtx(), 1), ShouldBeNil)
		So(f(backgroundCtx(), "2024-05-17T08:08:13.26635Z"), ShouldBeNil)
		So(f(backgroundCtx(), "2024-05-17T08:08:13Z"), ShouldBeNil)
		So(f(backgroundCtx(), "2024-05-17T08:08:13.26635+08:00"), ShouldBeNil)

		So(f(backgroundCtx(), ""), ShouldBeError, `date-time must be in rfc3339 format`)
	})
}

func TestFormatX509CertPem(t *testing.T) {
	f := FormatX509CertPem{}.CheckFormat

	Convey("TestFormatX509CertPem", t, func() {
		pemStr := "-----BEGIN CERTIFICATE-----\nMIIDejCCAmKgAwIBAgIgLKKTB6GZMFHZVUiFIq8LcNIr0p8HFHwKM6r5/BQ/un4w\nDQYJKoZIhvcNAQEFBQAwUDEJMAcGA1UEBhMAMQkwBwYDVQQKDAAxCTAHBgNVBAsM\nADENMAsGA1UEAwwEdGVzdDEPMA0GCSqGSIb3DQEJARYAMQ0wCwYDVQQDDAR0ZXN0\nMB4XDTI0MDgwODA2NTY0OFoXDTM0MDgwOTA2NTY0OFowQTEJMAcGA1UEBhMAMQkw\nBwYDVQQKDAAxCTAHBgNVBAsMADENMAsGA1UEAwwEdGVzdDEPMA0GCSqGSIb3DQEJ\nARYAMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA5zRfTtkaa7cIsQS+\nF1Dg25wPEvcjHsHcq598n+RzRJzfSLRtYwgEfs0VhyjHfo2O7KhNFh5cqdkEfzwA\nbfxtgVLvy3yUjTMFO0FnJqrO3dkGiOAl654XUlXb4rF8DF1sPnUdd9QEZaZHGV/8\nYuVOc3RV15jsr2jB9rra9//guAQ0CSP4XLJ5m9vf9nJILAHLryFIzDSgOVmhi4Ig\no59e9n3Hemavrta2C5Zj4cP6RNwuCV/i5lQOkzJIgksH9/EZCsR93DMEgkBS5oQQ\nrt9Bzlr03TNGW4n/CYKNULK/osqJd5r5g3zUaQZY2KAan+oSsEXvBjzYtrehN1dm\ndfbUEQIDAQABo08wTTAdBgNVHQ4EFgQUiXG6MG9PSB/clTIuzm8rW+8xLWkwHwYD\nVR0jBBgwFoAUiXG6MG9PSB/clTIuzm8rW+8xLWkwCwYDVR0RBAQwAoIAMA0GCSqG\nSIb3DQEBBQUAA4IBAQBTjdS9po3eEXukksMK6xBL3kQF1MEFUaWcgoN+h497lS9J\nXe1rmWpdZ1Aehp21GQmniRKU8uPLPRQKoX8Mhc/d3fHyv9u0YPns/2Wm8TBzxwHY\nV2KdXZfpBdN+Z5bBRbgtKxx1z2GBfB39S2WCakS9xK8f7fuQPLIZz8eq7so5T8Hm\nTU95acndEpnA0u6/MjbvXtZesTRZCewQw4CkcSLTCzB8dLG55UXHytnISWlCpuAx\n8svq/ryZIi5vhBQFO/hG9s2Q32VvfKt2ZW8qA+gvOxEVDfAEFekKokP0Taiz77Q2\nAVZxEXeABxJGtiMunQTr2q1tCrJQN0d08xlA5jXl\n-----END CERTIFICATE-----\n"
		So(f(backgroundCtx(), pemStr), ShouldBeNil)
		So(f(backgroundCtx(), "1234"), ShouldBeError, "invalid pem")
		So(f(backgroundCtx(), "-----BEGIN CERTIFICATE-----\nasdf\n-----END CERTIFICATE-----"), ShouldBeError, "invalid x509 cert")
	})
}

func TestFormatPublicImageURI(t *testing.T) {
	Convey("FormatPublicImageURI", t, func() {
		f := FormatPublicImageURI{}.CheckFormat

		So(f(backgroundCtx(), 1), ShouldBeNil)
		So(f(backgroundCtx(), ""), ShouldBeError, "invalid scheme: ")
		So(f(backgroundCtx(), "foobar:"), ShouldBeError, "invalid scheme: foobar")
		So(f(backgroundCtx(), "http://"), ShouldBeError, "invalid scheme: http")

		So(f(backgroundCtx(), "https://example.com"), ShouldBeNil)
		So(f(backgroundCtx(), "https://example.com/image.png"), ShouldBeNil)
	})
}
