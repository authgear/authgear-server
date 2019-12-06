package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

func TestLoginIDNormalizer(t *testing.T) {
	type Case struct {
		LoginID           string
		NormalizedLoginID string
	}
	f := func(c Case, n LoginIDNormalizer) {
		result, _ := n.Normalize(c.LoginID)
		So(result, ShouldEqual, c.NormalizedLoginID)
	}
	newTrue := func() *bool {
		b := true
		return &b
	}
	newFalse := func() *bool {
		b := false
		return &b
	}
	Convey("TestLoginIDEmailNormalizer", t, func() {
		Convey("default setting", func() {
			cases := []Case{
				{"Faseng@Example.com", "faseng@example.com"},
				{"Faseng+Chima@example.com", "faseng+chima@example.com"},
				{"faseng.the.cat@example.com", "faseng.the.cat@example.com"},
				{"faseng.ℌ𝒌@example.com", "faseng.hk@example.com"},
				{"faseng.ℌ𝒌@測試.香港", "faseng.hk@測試.香港"},
			}

			n := &LoginIDEmailNormalizer{
				config: &config.LoginIDTypeEmailConfiguration{
					CaseSensitive: newFalse(),
					BlockPlusSign: newFalse(),
					IgnoreDot:     newFalse(),
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
					CaseSensitive: newTrue(),
					BlockPlusSign: newFalse(),
					IgnoreDot:     newFalse(),
				},
			}

			for _, c := range cases {
				f(c, n)
			}
		})

		Convey("ingore dot", func() {
			cases := []Case{
				{"Faseng@Example.com", "faseng@example.com"},
				{"Faseng.The.Cat@example.com", "fasengthecat@example.com"},
			}

			n := &LoginIDEmailNormalizer{
				config: &config.LoginIDTypeEmailConfiguration{
					CaseSensitive: newFalse(),
					BlockPlusSign: newTrue(),
					IgnoreDot:     newTrue(),
				},
			}

			for _, c := range cases {
				f(c, n)
			}
		})

		Convey("compute unique key", func() {
			n := &LoginIDEmailNormalizer{}
			var uniqueKey string

			uniqueKey, _ = n.ComputeUniqueKey("Faseng+Chima@example.com")
			So(uniqueKey, ShouldEqual, "Faseng+Chima@example.com")

			uniqueKey, _ = n.ComputeUniqueKey("Faseng.The.Cat@測試.香港")
			So(uniqueKey, ShouldEqual, "Faseng.The.Cat@xn--g6w251d.xn--j6w193g")

		})
	})

	Convey("TestLoginIDEmailNormalizer", t, func() {
		Convey("case insensitive username", func() {
			cases := []Case{
				{"Faseng.ℌ𝒌", "faseng.hk"},
				{"fasengChima", "fasengchima"},
				{"grüßen", "grüssen"},
			}

			n := &LoginIDUsernameNormalizer{
				config: &config.LoginIDTypeUsernameConfiguration{
					CaseSensitive: newFalse(),
				},
			}

			for _, c := range cases {
				f(c, n)
			}
		})

		Convey("case sensitive username", func() {
			cases := []Case{
				{"Faseng.ℌ𝒌", "Faseng.Hk"},
				{"fasengChima", "fasengChima"},
				{"grüßen", "grüßen"},
			}

			n := &LoginIDUsernameNormalizer{
				config: &config.LoginIDTypeUsernameConfiguration{
					CaseSensitive: newTrue(),
				},
			}

			for _, c := range cases {
				f(c, n)
			}
		})
	})
}
