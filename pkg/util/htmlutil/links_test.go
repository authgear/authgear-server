package htmlutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSanitizeLinksHTML(t *testing.T) {
	Convey("SanitizeLinksHTML", t, func() {
		test := func(in string, expected string) {
			So(string(SanitizeLinksHTML(in)), ShouldEqual, expected)
		}

		Convey("keeps https and http links", func() {
			test(
				`I agree to the <a href="https://example.com/terms">Terms</a>`,
				`I agree to the <a href="https://example.com/terms" target="_blank" rel="noopener noreferrer">Terms</a>`,
			)
			test(
				`<a href="http://example.com">x</a>`,
				`<a href="http://example.com" target="_blank" rel="noopener noreferrer">x</a>`,
			)
		})

		Convey("closes a link left open", func() {
			test(
				`<a href="https://example.com">x`,
				`<a href="https://example.com" target="_blank" rel="noopener noreferrer">x</a>`,
			)
		})

		Convey("drops attributes other than href", func() {
			test(
				`<a href="https://example.com" onclick="alert(1)" target="_self">x</a>`,
				`<a href="https://example.com" target="_blank" rel="noopener noreferrer">x</a>`,
			)
		})

		Convey("escapes a link with an unsafe scheme", func() {
			test(
				`<a href="javascript:alert(1)">x</a>`,
				`&lt;a href=&#34;javascript:alert(1)&#34;&gt;x&lt;/a&gt;`,
			)
			test(
				`<a href="JaVaScRiPt:alert(1)">x</a>`,
				`&lt;a href=&#34;JaVaScRiPt:alert(1)&#34;&gt;x&lt;/a&gt;`,
			)
			test(`<a>x</a>`, `&lt;a&gt;x&lt;/a&gt;`)
		})

		Convey("escapes other tags", func() {
			test(`<script>alert(1)</script>`, `&lt;script&gt;alert(1)&lt;/script&gt;`)
			test(`<b>bold</b>`, `&lt;b&gt;bold&lt;/b&gt;`)
			test(`<img src=x onerror=alert(1)>`, `&lt;img src=x onerror=alert(1)&gt;`)
		})

		Convey("escapes text", func() {
			test(`a & b < c`, `a &amp; b &lt; c`)
		})
	})
}

func TestLinksToText(t *testing.T) {
	Convey("LinksToText", t, func() {
		So(LinksToText(`I agree to the <a href="https://example.com">Terms</a>`), ShouldEqual, "I agree to the Terms")
		So(LinksToText(`plain`), ShouldEqual, "plain")
	})
}
