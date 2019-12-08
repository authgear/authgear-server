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
				// no change
				{"faseng+chima@example.com", "faseng+chima@example.com"},
				{`"faseng@cat"@example.com`, `"faseng@cat"@example.com`},

				// case fold
				{"Faseng@Example.com", "faseng@example.com"},
				{"Faseng+Chima@example.com", "faseng+chima@example.com"},
				{"grüßen@example.com", "grüssen@example.com"},

				// NFKC + case fold
				{"faseng.ℌ𝒌@example.com", "faseng.hk@example.com"},
				{
					string([]byte{102, 97, 115, 101, 110, 103, 46, 226, 132, 140, 240, 157, 146, 140, 64, 101, 120, 97, 109, 112, 108, 101, 46, 99, 111, 109}),
					string([]byte{102, 97, 115, 101, 110, 103, 46, 104, 107, 64, 101, 120, 97, 109, 112, 108, 101, 46, 99, 111, 109}),
				},

				// no change for unicode domain
				{"faseng@測試.香港", "faseng@測試.香港"},
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
				// NFKC + case fold
				{"Faseng.ℌ𝒌", "faseng.hk"},
				{
					string([]byte{70, 97, 115, 101, 110, 103, 46, 226, 132, 140, 240, 157, 146, 140}),
					string([]byte{102, 97, 115, 101, 110, 103, 46, 104, 107}),
				},

				// case fold
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
				// NFKC
				{"Faseng.ℌ𝒌", "Faseng.Hk"},
				{
					string([]byte{70, 97, 115, 101, 110, 103, 46, 226, 132, 140, 240, 157, 146, 140}),
					string([]byte{70, 97, 115, 101, 110, 103, 46, 72, 107}),
				},

				// no change
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
