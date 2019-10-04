package template

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTemplateRender(t *testing.T) {
	Convey("template rendering", t, func() {
		Convey("should not allow disabled tags", func() {
			var err error

			_, err = ParseHTMLTemplate("{% include /etc/passwd %}", nil)
			So(err, ShouldBeError, "[Error (where: parser) in <string> | Line 1 Col 4 near 'include'] Usage of tag 'include' is not allowed (sandbox restriction active).")
		})
		Convey("should not render large templates", func() {
			var err error

			_, err = ParseHTMLTemplate(`{%for i in "0123456789abcdef"%}{%for i in "0123456789abcdef"%}{%for i in "0123456789abcdef"%}{%for i in "0123456789abcdef"%}{%for i in "0123456789abcdef"%}{%for i in "0123456789abcdef"%}{%for i in "0123456789abcdef"%}{%for i in "0123456789abcdef"%}{%for i in "0123456789abcdef"%}0123456789abcdef{%endfor%}{%endfor%}{%endfor%}{%endfor%}{%endfor%}{%endfor%}{%endfor%}{%endfor%}{%endfor%}`, nil)
			So(err, ShouldBeError, "UnexpectedError: rendered template is too large")
		})
	})
}
