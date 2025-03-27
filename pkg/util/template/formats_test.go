package template_test

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/template"
)

func backgroundCtx() context.Context {
	return context.Background()
}

func TestFormatTextTemplate(t *testing.T) {
	f := template.FormatTextTemplate{}.CheckFormat

	Convey("TestFormatTextTemplate", t, func() {
		So(f(backgroundCtx(), "{{.preferred_username}}@gmail.com"), ShouldBeNil)
		So(f(backgroundCtx(), "{{{}}"), ShouldBeError, "invalid text template")
		So(f(backgroundCtx(), "{{html .preferred_username}}@gmail.com"), ShouldBeError, ":1:2: forbidden identifier html")
		So(f(backgroundCtx(), "{{template \"t1\"}}"), ShouldBeError, ":1:11: forbidden construct *parse.TemplateNode")
	})
}
