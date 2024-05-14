package loginid

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/internalinterface"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestNormalizers(t *testing.T) {
	type Case struct {
		LoginID           string
		NormalizedLoginID string
	}
	f := func(c Case, n internalinterface.LoginIDNormalizer) {
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
	Convey("EmailNormalizer", t, func() {
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

			n := &EmailNormalizer{
				Config: &config.LoginIDEmailConfig{
					CaseSensitive: newFalse(),
					BlockPlusSign: newFalse(),
					IgnoreDotSign: newFalse(),
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

			n := &EmailNormalizer{
				Config: &config.LoginIDEmailConfig{
					CaseSensitive: newTrue(),
					BlockPlusSign: newFalse(),
					IgnoreDotSign: newFalse(),
				},
			}

			for _, c := range cases {
				f(c, n)
			}
		})

		Convey("ignore dot", func() {
			cases := []Case{
				{"Faseng@Example.com", "faseng@example.com"},
				{"Faseng.The.Cat@example.com", "fasengthecat@example.com"},
			}

			n := &EmailNormalizer{
				Config: &config.LoginIDEmailConfig{
					CaseSensitive: newFalse(),
					BlockPlusSign: newTrue(),
					IgnoreDotSign: newTrue(),
				},
			}

			for _, c := range cases {
				f(c, n)
			}
		})

		Convey("compute unique key", func() {
			n := &EmailNormalizer{}
			var uniqueKey string

			uniqueKey, _ = n.ComputeUniqueKey("Faseng+Chima@example.com")
			So(uniqueKey, ShouldEqual, "Faseng+Chima@example.com")

			uniqueKey, _ = n.ComputeUniqueKey("Faseng.The.Cat@測試.香港")
			So(uniqueKey, ShouldEqual, "Faseng.The.Cat@xn--g6w251d.xn--j6w193g")

		})
	})

	Convey("UsernameNormalizer", t, func() {
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

			n := &UsernameNormalizer{
				Config: &config.LoginIDUsernameConfig{
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

			n := &UsernameNormalizer{
				Config: &config.LoginIDUsernameConfig{
					CaseSensitive: newTrue(),
				},
			}

			for _, c := range cases {
				f(c, n)
			}
		})
	})

	Convey("PhoneNumberNormalizer", t, func() {
		Convey("normalize to e164", func() {
			cases := []Case{
				{"+85298887766", "+85298887766"},
				{
					"+852-98887766",
					"+85298887766",
				},
				{
					"+852-98-88-77-66",
					"+85298887766",
				},
			}

			n := &PhoneNumberNormalizer{}

			for _, c := range cases {
				f(c, n)
			}
		})
	})
}
