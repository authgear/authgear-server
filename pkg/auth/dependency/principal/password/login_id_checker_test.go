package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

func TestLoginIDChecker(t *testing.T) {
	type Case struct {
		LoginID string
		Err     string
	}
	f := func(c Case, check LoginIDTypeChecker) {
		err := check.Validate(c.LoginID)

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
	Convey("TestLoginIDEmailChecker", t, func() {
		Convey("default setting", func() {
			cases := []Case{
				{"Faseng@Example.com", ""},
				{"Faseng+Chima@example.com", ""},
				{"faseng.the.cat", "invalid login ID"},
				{"Faseng <faseng@example>", "invalid login ID"},
				{"faseng.ℌ𝒌@測試.香港", ""},
			}

			check := &LoginIDEmailChecker{
				config: &config.LoginIDTypeEmailConfiguration{
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
				{"Faseng+Chima@example.com", "invalid login ID"},
			}

			checker := &LoginIDEmailChecker{
				config: &config.LoginIDTypeEmailConfiguration{
					BlockPlusSign: newTrue(),
				},
			}

			for _, c := range cases {
				f(c, checker)
			}
		})
	})

	Convey("TestLoginIDUsernameChecker", t, func() {
		Convey("allow all", func() {
			cases := []Case{
				{"admin", ""},
				{"settings", ""},
				{"skygear", ""},
				{"花生thecat", ""},
				{"faseng", ""},
			}

			n := &LoginIDUsernameChecker{
				config: &config.LoginIDTypeUsernameConfiguration{
					BlockReservedKeywords: newFalse(),
					ExcludedKeywords:      []string{},
					ASCIIOnly:             newFalse(),
				},
			}

			for _, c := range cases {
				f(c, n)
			}
		})

		Convey("block keywords and non ascii", func() {
			cases := []Case{
				{"admin", "invalid login ID"},
				{"settings", "invalid login ID"},
				{"skygear", "invalid login ID"},
				{"花生thecat", "invalid login ID"},
				{"faseng", ""},
			}

			n := &LoginIDUsernameChecker{
				config: &config.LoginIDTypeUsernameConfiguration{
					BlockReservedKeywords: newTrue(),
					ExcludedKeywords:      []string{"skygear"},
					ASCIIOnly:             newTrue(),
				},
				reservedNameSourceFile: "../../../../../reserved_name.txt",
			}

			for _, c := range cases {
				f(c, n)
			}
		})
	})
}
