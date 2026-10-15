package attrs

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func TestValidateStringConstraints(t *testing.T) {
	intPtr := func(i int) *int { return &i }

	Convey("ValidateStringConstraints", t, func() {
		run := func(c *config.StringConstraints, value string, multiline bool) *validation.AggregatedError {
			vctx := &validation.Context{}
			ValidateStringConstraints(vctx.Child("x"), c, value, multiline)
			err := vctx.Error("invalid")
			if err == nil {
				return nil
			}
			return err.(*validation.AggregatedError)
		}
		keywords := func(e *validation.AggregatedError) []string {
			var out []string
			if e == nil {
				return out
			}
			for _, err := range e.Errors {
				out = append(out, err.Keyword)
			}
			return out
		}

		Convey("counts code points", func() {
			c := &config.StringConstraints{MaxLength: intPtr(2)}
			So(run(c, "😀😀", false), ShouldBeNil)
			So(keywords(run(c, "😀😀😀", false)), ShouldResemble, []string{"maxLength"})

			c = &config.StringConstraints{MinLength: intPtr(2)}
			e := run(c, "😀", false)
			So(e.Errors[0].Keyword, ShouldEqual, "minLength")
			So(e.Errors[0].Info, ShouldResemble, map[string]any{"expected": 2, "actual": 1})
			So(e.Errors[0].Location, ShouldEqual, "/x")
		})

		Convey("allowed characters", func() {
			letters := &config.StringConstraints{AllowedCharacters: []config.AllowedCharacters{config.AllowedCharactersLetters}}
			So(run(letters, "Ace", false), ShouldBeNil)
			So(run(letters, "é", false), ShouldBeNil)
			// Decomposed é: e + combining acute accent.
			So(run(letters, "é", false), ShouldBeNil)
			So(keywords(run(letters, "Ace1", false)), ShouldResemble, []string{"allowedCharacters"})
			So(keywords(run(letters, "A ce", false)), ShouldResemble, []string{"allowedCharacters"})
			So(keywords(run(letters, "Ace!", false)), ShouldResemble, []string{"allowedCharacters"})

			digits := &config.StringConstraints{AllowedCharacters: []config.AllowedCharacters{config.AllowedCharactersDigits}}
			So(run(digits, "0123", false), ShouldBeNil)
			So(keywords(run(digits, "01a", false)), ShouldResemble, []string{"allowedCharacters"})

			spaces := &config.StringConstraints{AllowedCharacters: []config.AllowedCharacters{config.AllowedCharactersSpaces}}
			So(run(spaces, "  ", false), ShouldBeNil)

			symbols := &config.StringConstraints{AllowedCharacters: []config.AllowedCharacters{config.AllowedCharactersSymbols}}
			So(run(symbols, "!@#-", false), ShouldBeNil)
			So(keywords(run(symbols, "a", false)), ShouldResemble, []string{"allowedCharacters"})

			Convey("rejects tab", func() {
				all := &config.StringConstraints{AllowedCharacters: []config.AllowedCharacters{
					config.AllowedCharactersLetters, config.AllowedCharactersDigits,
					config.AllowedCharactersSpaces, config.AllowedCharactersSymbols,
				}}
				So(keywords(run(all, "a\tb", false)), ShouldResemble, []string{"allowedCharacters"})
			})

			Convey("a mark takes the kind of the character before it", func() {
				// A heart with an emoji variation selector.
				So(run(symbols, "\u2764\ufe0f", false), ShouldBeNil)
				// A keycap emoji counts as a digit.
				So(run(digits, "1\ufe0f\u20e3", false), ShouldBeNil)
				// A mark with nothing before it.
				So(keywords(run(letters, "\u0301a", false)), ShouldResemble, []string{"allowedCharacters"})
			})

			Convey("a joiner takes the kind of the character before it", func() {
				// A Persian word with a zero width non-joiner.
				So(run(letters, "می\u200cخواهم", false), ShouldBeNil)
				// Family emoji, with zero width joiners.
				So(run(symbols, "👨\u200d👩\u200d👧", false), ShouldBeNil)
				So(keywords(run(letters, "👨\u200d👩", false)), ShouldResemble, []string{"allowedCharacters"})
				So(keywords(run(letters, "\u200ca", false)), ShouldResemble, []string{"allowedCharacters"})
				So(keywords(run(letters, "a\n\u200cb", true)), ShouldResemble, []string{"allowedCharacters"})
			})

			Convey("newline only when multiline", func() {
				c := &config.StringConstraints{AllowedCharacters: []config.AllowedCharacters{config.AllowedCharactersLetters}}
				So(keywords(run(c, "a\nb", false)), ShouldResemble, []string{"allowedCharacters"})
				So(run(c, "a\nb", true), ShouldBeNil)
				So(run(c, "a\r\nb", true), ShouldBeNil)
			})

			Convey("info lists the config order", func() {
				c := &config.StringConstraints{AllowedCharacters: []config.AllowedCharacters{
					config.AllowedCharactersDigits, config.AllowedCharactersLetters,
				}}
				e := run(c, "!", false)
				So(e.Errors[0].Info, ShouldResemble, map[string]any{
					"expected": []config.AllowedCharacters{config.AllowedCharactersDigits, config.AllowedCharactersLetters},
				})
			})
		})

		Convey("no character check without allowed_characters", func() {
			c := &config.StringConstraints{MinLength: intPtr(1)}
			So(run(c, "a\tb!😀", false), ShouldBeNil)
		})

		Convey("pattern is unanchored", func() {
			c := &config.StringConstraints{Pattern: "[0-9]+"}
			So(run(c, "abc123", false), ShouldBeNil)
			e := run(c, "abc", false)
			So(e.Errors[0].Keyword, ShouldEqual, "pattern")
			So(e.Errors[0].Info, ShouldResemble, map[string]any{"expected": "[0-9]+"})
		})

		Convey("emits all failures together", func() {
			c := &config.StringConstraints{
				MinLength:         intPtr(5),
				AllowedCharacters: []config.AllowedCharacters{config.AllowedCharactersLetters},
				Pattern:           "^z",
			}
			So(keywords(run(c, "a1", false)), ShouldResemble, []string{"minLength", "allowedCharacters", "pattern"})
		})

		Convey("skips an empty value and nil constraints", func() {
			So(run(&config.StringConstraints{MinLength: intPtr(5)}, "", false), ShouldBeNil)
			So(run(nil, "a", false), ShouldBeNil)
		})
	})

	Convey("IsMultilineStandardAttribute", t, func() {
		So(IsMultilineStandardAttribute("/address/formatted"), ShouldBeTrue)
		So(IsMultilineStandardAttribute("/address/street_address"), ShouldBeTrue)
		So(IsMultilineStandardAttribute("/address/locality"), ShouldBeFalse)
		So(IsMultilineStandardAttribute("/given_name"), ShouldBeFalse)
	})
}
