package password

import (
	"math/rand"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type TestRandSource struct {
}

// Pseudo-random pick
func (s *TestRandSource) Pick(list string) (int, error) {
	index := rand.Intn(len(list))
	return index, nil
}

func (s *TestRandSource) Shuffle(list string) (string, error) {
	return list, nil
}

func TestBasicPasswordGeneration(t *testing.T) {
	Convey("Given a password generator with default settings", t, func() {
		generator := &Generator{
			Checker:     &Checker{},
			RandSource:  &TestRandSource{},
			PwMinLength: 8,
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
			Checker:             &Checker{},
			RandSource:          &TestRandSource{},
			PwMinLength:         8,
			PwUppercaseRequired: true,
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
			Checker:             &Checker{},
			RandSource:          &TestRandSource{},
			PwMinLength:         8,
			PwLowercaseRequired: true,
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
			Checker:             &Checker{},
			RandSource:          &TestRandSource{},
			PwMinLength:         12,
			PwUppercaseRequired: true,
			PwLowercaseRequired: true,
			PwDigitRequired:     true,
			PwSymbolRequired:    true,
		}

		password, err := generator.Generate()

		Convey("should meet all requirements", func() {
			So(checkPasswordUppercase(password), ShouldBeTrue)
			So(checkPasswordLowercase(password), ShouldBeTrue)
			So(checkPasswordDigit(password), ShouldBeTrue)
			So(checkPasswordSymbol(password), ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})
}

func TestMinLengthRequirement(t *testing.T) {
	Convey("Given a password generator with a minimum length requirement", t, func() {
		generator := &Generator{
			Checker:     &Checker{},
			RandSource:  &TestRandSource{},
			PwMinLength: 40,
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
			Checker:             &Checker{},
			RandSource:          &TestRandSource{},
			PwMinLength:         8,
			PwMinGuessableLevel: 4,
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
		excluded := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}
		generator := &Generator{
			Checker: &Checker{
				PwExcludedKeywords: excluded,
			},
			RandSource:         &TestRandSource{},
			PwMinLength:        8,
			PwDigitRequired:    true,
			PwExcludedKeywords: excluded,
		}

		password, err := generator.Generate()

		Convey("should not contain any excluded keywords", func() {
			So(checkPasswordExcludedKeywords(password, excluded), ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})
}
