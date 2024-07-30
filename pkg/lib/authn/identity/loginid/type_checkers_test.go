package loginid

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/blocklist"
	"github.com/authgear/authgear-server/pkg/util/matchlist"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func TestLoginIDTypeCheckers(t *testing.T) {
	type Case struct {
		LoginID string
		Err     string
	}
	f := func(c Case, check TypeChecker) {
		ctx := &validation.Context{}
		check.Validate(ctx, c.LoginID)
		err := ctx.Error("invalid login ID")
		if c.Err == "" {
			So(err, ShouldBeNil)
		} else {
			So(err, ShouldBeError, c.Err)
		}
	}
	newTrue := func() *bool {
		b := true
		return &b
	}
	newFalse := func() *bool {
		b := false
		return &b
	}
	Convey("EmailChecker", t, func() {
		Convey("default setting", func() {
			cases := []Case{
				{"Faseng@Example.com", ""},
				{"Faseng+Chima@example.com", ""},
				{"faseng.the.cat", "invalid login ID:\n/login_id: format\n  map[format:email]"},
				{"fasengthecat", "invalid login ID:\n/login_id: format\n  map[format:email]"},
				{"fasengthecat@", "invalid login ID:\n/login_id: format\n  map[format:email]"},
				{"@fasengthecat", "invalid login ID:\n/login_id: format\n  map[format:email]"},
				{"Faseng <faseng@example>", "invalid login ID:\n/login_id: format\n  map[format:email]"},
				{"faseng.ℌ𝒌@測試.香港", ""},
				{`"fase ng@cat"@example.com`, ""},
				{`"faseng@"@example.com`, ""},
			}

			check := &EmailChecker{
				Config: &config.LoginIDEmailConfig{
					BlockPlusSign: newFalse(),
				},
			}

			for _, c := range cases {
				f(c, check)
			}
		})

		Convey("block plus sign", func() {
			cases := []Case{
				{"Faseng@Example.com", ""},
				{"Faseng+Chima@example.com", "invalid login ID:\n/login_id: format\n  map[format:email]"},
				{`"faseng@cat+123"@example.com`, "invalid login ID:\n/login_id: format\n  map[format:email]"},
			}

			checker := &EmailChecker{
				Config: &config.LoginIDEmailConfig{
					BlockPlusSign: newTrue(),
				},
			}

			for _, c := range cases {
				f(c, checker)
			}
		})

		Convey("email domain blocklist", func() {
			cases := []Case{
				{"Faseng@Example.com", "invalid login ID:\n/login_id: blocked\n  map[reason:EmailDomainBlocklist]"},
				{"faseng@example.com", "invalid login ID:\n/login_id: blocked\n  map[reason:EmailDomainBlocklist]"},
				{"faseng@testing.com", "invalid login ID:\n/login_id: blocked\n  map[reason:EmailDomainBlocklist]"},
				{"faseng@TESTING.COM", "invalid login ID:\n/login_id: blocked\n  map[reason:EmailDomainBlocklist]"},
				{`faseng@authgear.io`, ""},
			}

			domainsList, _ := matchlist.New(`
				example.com
				TESTING.COM
			`, true, false)
			checker := &EmailChecker{
				Config: &config.LoginIDEmailConfig{
					BlockPlusSign: newFalse(),
				},
				DomainBlockList: domainsList,
			}

			for _, c := range cases {
				f(c, checker)
			}
		})

		Convey("block free email provider domains", func() {
			cases := []Case{
				{"faseng@free-mail.com", "invalid login ID:\n/login_id: blocked\n  map[reason:EmailDomainBlocklist]"},
				{"faseng@FREE-MAIL.COM", "invalid login ID:\n/login_id: blocked\n  map[reason:EmailDomainBlocklist]"},
				{`faseng@authgear.io`, ""},
			}

			domainsList, _ := matchlist.New(`
				FREE-MAIL.COM
			`, true, false)
			checker := &EmailChecker{
				Config: &config.LoginIDEmailConfig{
					BlockPlusSign: newFalse(),
				},
				BlockFreeEmailProviderDomains: domainsList,
			}

			for _, c := range cases {
				f(c, checker)
			}
		})

		Convey("email domain allowlist", func() {
			cases := []Case{
				{"Faseng@Example.com", ""},
				{"faseng@example.com", ""},
				{"faseng@free-mail.com", ""},
				{`"faseng@cat+123"@authgear.io`, "invalid login ID:\n/login_id: blocked\n  map[reason:EmailDomainAllowlist]"},
			}

			domainsList, _ := matchlist.New(`
				example.com
				testing.com

				FREE-MAIL.COM
			`, true, false)
			checker := &EmailChecker{
				Config: &config.LoginIDEmailConfig{
					BlockPlusSign: newFalse(),
				},
				DomainAllowList: domainsList,
			}

			for _, c := range cases {
				f(c, checker)
			}
		})

	})

	Convey("UsernameChecker", t, func() {
		Convey("allow all", func() {
			cases := []Case{
				{"admin", ""},
				{"settings", ""},
				{"authgear", ""},
				{"花生thecat", ""},
				{"faseng", ""},

				// space is not allowed in Identifier class
				{"Test ID", "invalid login ID:\n/login_id: format\n  map[format:username]"},

				// confusable homoglyphs
				{"microsoft", ""},
				{"microsоft", "invalid login ID:\n/login_id: username contains confusable characters"},
				// byte array versions
				{string([]byte{109, 105, 99, 114, 111, 115, 111, 102, 116}), ""},
				{string([]byte{109, 105, 99, 114, 111, 115, 208, 190, 102, 116}), "invalid login ID:\n/login_id: username contains confusable characters"},
			}

			n := &UsernameChecker{
				Config: &config.LoginIDUsernameConfig{
					ASCIIOnly: newFalse(),
				},
			}

			for _, c := range cases {
				f(c, n)
			}
		})

		Convey("block keywords and non ascii", func() {
			cases := []Case{
				{"admin", "invalid login ID:\n/login_id: blocked\n  map[reason:UsernameReserved]"},
				{"settings", "invalid login ID:\n/login_id: blocked\n  map[reason:UsernameReserved]"},
				{"authgear", "invalid login ID:\n/login_id: blocked\n  map[reason:UsernameExcludedKeywords]"},
				{"myauthgearapp", "invalid login ID:\n/login_id: blocked\n  map[reason:UsernameExcludedKeywords]"},
				{"花生thecat", "invalid login ID:\n/login_id: format\n  map[format:username]"},
				{"faseng", ""},
				{"faseng_chima-the.cat", ""},
			}

			reversedNames, _ := blocklist.New(`
				admin
				settings
			`)
			excludedKeywords, _ := matchlist.New(`
				authgear
			`, true, true)
			n := &UsernameChecker{
				Config: &config.LoginIDUsernameConfig{
					ASCIIOnly: newTrue(),
				},
				ReservedNames:    reversedNames,
				ExcludedKeywords: excludedKeywords,
			}

			for _, c := range cases {
				f(c, n)
			}
		})
	})

	Convey("PhoneChecker", t, func() {
		Convey("country code allowlist", func() {
			cases := []Case{
				{"+85298765432", "invalid login ID:\n/login_id: blocked\n  map[reason:PhoneNumberCountryCodeAllowlist]"},
				{"+12124567890", ""},
			}

			n := &PhoneChecker{
				Alpha2AllowList: []string{"US"},
			}

			for _, c := range cases {
				f(c, n)
			}
		})
	})
}
