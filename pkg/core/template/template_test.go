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
			longStr := strings.Repeat("&", MaxTemplateSize-50)
			template := fmt.Sprintf(`{{html (html (html (html "%s")))}}`, longStr)

			var err error

			_, err = ParseTextTemplate("test", template, nil)
			So(err, ShouldBeError, "UnexpectedError: rendered template is too large")
		})
	})
}
