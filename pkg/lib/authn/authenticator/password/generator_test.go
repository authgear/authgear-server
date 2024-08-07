package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func newInt(v int) *int { return &v }

type TestRandSource struct {
}

// Pseudo-random pick
func (s *TestRandSource) RandomBytes(n int) ([]byte, error) {
	return []byte("1234567890"), nil
}

func (s *TestRandSource) Shuffle(list string) (string, error) {
	return list, nil
}

func TestBasicPasswordGeneration(t *testing.T) {
	Convey("Given a password generator with default settings", t, func() {
		generator := &Generator{
			Checker:    &Checker{},
			RandSource: &TestRandSource{},
			Policy: &config.PasswordPolicyConfig{
				MinLength: newInt(8),
			},
		}

		password, err := generator.Generate()

		Convey("should be at least 8 characters long", func() {
			So(len(password), ShouldBeGreaterThanOrEqualTo, 8)
			So(err, ShouldBeNil)
		})
	})
}

func TestUppercaseRequirement(t *testing.T) {
	Convey("Given a password generator requiring at least one uppercase letter", t, func() {
		generator := &Generator{
			Checker:    &Checker{},
			RandSource: &TestRandSource{},
			Policy: &config.PasswordPolicyConfig{
				MinLength:         newInt(8),
				UppercaseRequired: true,
			},
		}

		password, err := generator.Generate()

		Convey("should include at least one uppercase letter", func() {
			So(checkPasswordUppercase(password), ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})
}

func TestLowercaseRequirement(t *testing.T) {
	Convey("Given a password generator requiring at least one lowercase letter", t, func() {
		generator := &Generator{
			Checker:    &Checker{},
			RandSource: &TestRandSource{},
			Policy: &config.PasswordPolicyConfig{
				MinLength:         newInt(8),
				LowercaseRequired: true,
			},
		}

		password, err := generator.Generate()

		Convey("should include at least one lowercase letter", func() {
			So(checkPasswordLowercase(password), ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})
}

func TestCombinedRequirements(t *testing.T) {
	Convey("Given a password generator with multiple requirements", t, func() {
		generator := &Generator{
			Checker:    &Checker{},
			RandSource: &TestRandSource{},
			Policy: &config.PasswordPolicyConfig{
				MinLength:         newInt(12),
				UppercaseRequired: true,
				LowercaseRequired: true,
				DigitRequired:     true,
				SymbolRequired:    true,
			},
		}

		password, err := generator.Generate()

		Convey("should meet all requirements", func() {
			So(checkPasswordUppercase(password), ShouldBeTrue)
			So(checkPasswordLowercase(password), ShouldBeTrue)
			So(checkPasswordDigit(password), ShouldBeTrue)
			So(checkPasswordSymbol(password), ShouldBeTrue)
			So(len(password), ShouldEqual, 12)
			So(err, ShouldBeNil)
		})
	})
}

func TestMinLengthRequirement(t *testing.T) {
	Convey("Given a password generator with a minimum length requirement", t, func() {
		generator := &Generator{
			Checker:    &Checker{},
			RandSource: &TestRandSource{},
			Policy: &config.PasswordPolicyConfig{
				MinLength: newInt(40),
			},
		}

		password, err := generator.Generate()

		Convey("should meet the minimum length requirement", func() {
			So(len(password), ShouldBeGreaterThanOrEqualTo, 40)
			So(err, ShouldBeNil)
		})
	})
}

func TestMinGuessableLevelRequirement(t *testing.T) {
	Convey("Given a password generator with a minimum guessable level requirement", t, func() {
		generator := &Generator{
			Checker:    &Checker{},
			RandSource: &CryptoRandSource{},
			Policy: &config.PasswordPolicyConfig{
				MinLength:             newInt(8),
				MinimumGuessableLevel: 4,
			},
		}

		password, err := generator.Generate()

		Convey("should meet the minimum guessable level requirement", func() {
			level, _ := checkPasswordGuessableLevel(password, 4)
			So(len(password), ShouldBeGreaterThanOrEqualTo, 32)
			So(level, ShouldBeGreaterThanOrEqualTo, 4)
			So(err, ShouldBeNil)
		})
	})
}

func TestExcludedKeywordsRequirement(t *testing.T) {
	Convey("Given a password generator with excluded keywords", t, func() {
		excluded := []string{"1", "2", "3"}
		generator := &Generator{
			Checker: &Checker{
				PwExcludedKeywords: excluded,
			},
			RandSource: &CryptoRandSource{},
			Policy: &config.PasswordPolicyConfig{
				MinLength:        newInt(8),
				DigitRequired:    true,
				ExcludedKeywords: excluded,
			},
		}

		password, err := generator.Generate()

		Convey("should not contain any excluded keywords", func() {
			So(checkPasswordExcludedKeywords(password, excluded), ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})
}
