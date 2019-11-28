package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

func TestLoginIDEmailNormalizer(t *testing.T) {
	newTrue := func() *bool {
		b := true
		return &b
	}
	newFalse := func() *bool {
		b := false
		return &b
	}
	Convey("TestLoginIDEmailNormalizer", t, func() {
		type Case struct {
			Email           string
			NormalizedEmail string
		}
		f := func(c Case, n *LoginIDEmailNormalizer) {
			result, _ := n.Normalize(c.Email)
			So(result, ShouldEqual, c.NormalizedEmail)
		}

		Convey("default setting", func() {
			cases := []Case{
				{"Faseng@Example.com", "faseng@example.com"},
				{"Faseng+Chima@example.com", "faseng+chima@example.com"},
				{"faseng.the.cat@example.com", "faseng.the.cat@example.com"},
				{"faseng.ℌ𝒌@example.com", "faseng.hk@example.com"},
				{"faseng.ℌ𝒌@測試.香港", "faseng.hk@xn--g6w251d.xn--j6w193g"},
			}

			n := &LoginIDEmailNormalizer{
				config: &config.LoginIDTypeEmailConfiguration{
					CaseSensitive:                newFalse(),
					IgnoreLocalPartAfterPlusSign: newFalse(),
					IgnoreDot:                    newFalse(),
				},
			}

			for _, c := range cases {
				f(c, n)
			}
		})

		Convey("case sensitive email", func() {
			cases := []Case{
				{"Faseng@Example.com", "Faseng@example.com"},
				{"Faseng+Chima@example.com", "Faseng+Chima@example.com"},
				{"Faseng.The.Cat@example.com", "Faseng.The.Cat@example.com"},
			}

			n := &LoginIDEmailNormalizer{
				config: &config.LoginIDTypeEmailConfiguration{
					CaseSensitive:                newTrue(),
					IgnoreLocalPartAfterPlusSign: newFalse(),
					IgnoreDot:                    newFalse(),
				},
			}

			for _, c := range cases {
				f(c, n)
			}
		})

		Convey("ingore plus and dot", func() {
			cases := []Case{
				{"Faseng@Example.com", "faseng@example.com"},
				{"Faseng+Chima@example.com", "faseng@example.com"},
				{"Faseng.The.Cat@example.com", "fasengthecat@example.com"},
			}

			n := &LoginIDEmailNormalizer{
				config: &config.LoginIDTypeEmailConfiguration{
					CaseSensitive:                newFalse(),
					IgnoreLocalPartAfterPlusSign: newTrue(),
					IgnoreDot:                    newTrue(),
				},
			}

			for _, c := range cases {
				f(c, n)
			}
		})

	})
}
