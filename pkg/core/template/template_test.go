package template

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTemplateRender(t *testing.T) {
	Convey("template rendering", t, func() {
		Convey("should not render large templates", func() {
			longStr := strings.Repeat("\\", 1024*512)
			template := fmt.Sprintf(`{{if $v := "%s" | js}}{{$v|js}}{{$v|js}}{{$v|js}}{{$v|js}}{{end}}`, longStr)

			var err error

			_, err = ParseHTMLTemplate("test", template, nil)
			So(err, ShouldBeError, "UnexpectedError: rendered template is too large")
		})
	})
}
