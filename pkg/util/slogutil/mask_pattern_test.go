package slogutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRegexMaskPattern(t *testing.T) {
	Convey("RegexMaskPattern", t, func() {
		Convey("should mask matching patterns", func() {
			pattern := NewRegexMaskPattern(`\d+`)
			result := pattern.Mask("user123", "***")
			So(result, ShouldEqual, "user***")
		})

		Convey("should mask multiple occurrences", func() {
			pattern := NewRegexMaskPattern(`\d+`)
			result := pattern.Mask("user123 has 456 points", "***")
			So(result, ShouldEqual, "user*** has *** points")
		})

		Convey("should handle no matches", func() {
			pattern := NewRegexMaskPattern(`\d+`)
			result := pattern.Mask("no numbers here", "***")
			So(result, ShouldEqual, "no numbers here")
		})

		Convey("should handle empty string", func() {
			pattern := NewRegexMaskPattern(`\d+`)
			result := pattern.Mask("", "***")
			So(result, ShouldEqual, "")
		})

		Convey("should handle complex regex patterns", func() {
			pattern := NewRegexMaskPattern(`[a-zA-Z]+@[a-zA-Z]+\.[a-zA-Z]+`)
			result := pattern.Mask("Contact user@example.com for help", "***")
			So(result, ShouldEqual, "Contact *** for help")
		})

		Convey("should handle different mask strings", func() {
			pattern := NewRegexMaskPattern(`password`)
			result := pattern.Mask("password=secret", "[REDACTED]")
			So(result, ShouldEqual, "[REDACTED]=secret")
		})

		Convey("should handle case sensitive matching", func() {
			pattern := NewRegexMaskPattern(`Password`)
			result := pattern.Mask("password and Password", "***")
			So(result, ShouldEqual, "password and ***")
		})

		Convey("should handle word boundaries", func() {
			pattern := NewRegexMaskPattern(`\bsecret\b`)
			result := pattern.Mask("secret and secretive", "***")
			So(result, ShouldEqual, "*** and secretive")
		})
	})
}

func TestPlainMaskPattern(t *testing.T) {
	Convey("PlainMaskPattern", t, func() {
		Convey("should mask exact string matches", func() {
			pattern := NewPlainMaskPattern("password")
			result := pattern.Mask("password=secret", "***")
			So(result, ShouldEqual, "***=secret")
		})

		Convey("should mask multiple occurrences", func() {
			pattern := NewPlainMaskPattern("test")
			result := pattern.Mask("test and test again", "***")
			So(result, ShouldEqual, "*** and *** again")
		})

		Convey("should handle no matches", func() {
			pattern := NewPlainMaskPattern("password")
			result := pattern.Mask("no sensitive data", "***")
			So(result, ShouldEqual, "no sensitive data")
		})

		Convey("should handle empty string", func() {
			pattern := NewPlainMaskPattern("password")
			result := pattern.Mask("", "***")
			So(result, ShouldEqual, "")
		})

		Convey("empty pattern mask nothing", func() {
			pattern := NewPlainMaskPattern("")
			result := pattern.Mask("some text", "***")
			So(result, ShouldEqual, "some text")
		})

		Convey("should handle case sensitive matching", func() {
			pattern := NewPlainMaskPattern("Password")
			result := pattern.Mask("password and Password", "***")
			So(result, ShouldEqual, "password and ***")
		})

		Convey("should handle substring matches", func() {
			pattern := NewPlainMaskPattern("secret")
			result := pattern.Mask("secret and secretive", "***")
			So(result, ShouldEqual, "*** and ***ive")
		})

		Convey("should handle different mask strings", func() {
			pattern := NewPlainMaskPattern("token")
			result := pattern.Mask("token=abc123", "[HIDDEN]")
			So(result, ShouldEqual, "[HIDDEN]=abc123")
		})

		Convey("should handle special characters", func() {
			pattern := NewPlainMaskPattern("$pecial")
			result := pattern.Mask("$pecial characters", "***")
			So(result, ShouldEqual, "*** characters")
		})
	})
}
