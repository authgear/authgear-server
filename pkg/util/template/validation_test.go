package template

import (
	"fmt"
	"html/template"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTemplateValidation(t *testing.T) {
	Convey("template validation", t, func() {
		template := func(s string) *template.Template {
			return template.Must(template.New("email").Parse(s))
		}

		Convey("should allow good templates", func() {
			var err error
			v := NewValidator()

			err = v.ValidateHTMLTemplate(template(`{{ if ne .UserName "" }}Welcome, {{ .UserName }}{{ else }}Please login{{ end }}`))
			So(err, ShouldBeNil)
		})

		Convey("should not allow disabled constructs", func() {
			var err error
			v := NewValidator()

			err = v.ValidateHTMLTemplate(template(`{{ range $i, $e := . }}{{$i}}{{$e}}{{ end }}`))
			So(err, ShouldBeError, "email:1:9: forbidden construct *parse.RangeNode")

			err = v.ValidateHTMLTemplate(template(`{{block "name" ""}} Test {{ template "name" }} {{end}}`))
			So(err, ShouldBeError, "email:1:8: forbidden construct *parse.TemplateNode")

			err = v.ValidateHTMLTemplate(template(`
			{{ with $v := js "\\" }}
				{{ with $v := js $v }}
					{{ with $v := js $v }}
						{{ with $v := js $v }}
							{{$v}}
						{{end}}
					{{end}}
				{{end}}
			{{end}}`))
			So(err, ShouldBeError, "email:2:11: forbidden construct *parse.WithNode")
		})

		Convey("should not allow disabled functions", func() {
			var err error
			v := NewValidator()

			err = v.ValidateHTMLTemplate(template(`{{printf "%010000000d" 0}}`))
			So(err, ShouldBeError, "email:1:2: forbidden identifier printf")
		})

		Convey("should not allow variable declaration", func() {
			var err error
			v := NewValidator()
			longStr := strings.Repeat("\\", 1024*512)

			err = v.ValidateHTMLTemplate(template(fmt.Sprintf(`{{if $v := "%s" | js}}{{$v|js}}{{$v|js}}{{$v|js}}{{$v|js}}{{end}}`, longStr)))
			So(err, ShouldBeError, "email:1:5: declaration is forbidden")

			err = v.ValidateHTMLTemplate(template(fmt.Sprintf(`{{$v = "%s"}}{{$v|js}}{{$v|js}}{{$v|js}}{{$v|js}}`, longStr)))
			So(err, ShouldBeError, "email:1:2: declaration is forbidden")
		})

		Convey("should not allow nesting too deep", func() {
			var err error

			v := NewValidator()
			err = v.ValidateHTMLTemplate(template(`{{ js (js (js "\\" | js | js | js) | js | js | js) | js | js | js }}`))
			So(err, ShouldBeError, "email:1:3: pipeline is forbidden")

			err = v.ValidateHTMLTemplate(template(`{{ js (js (js (js "\\"))) }}`))
			So(err, ShouldBeNil)

			err = v.ValidateHTMLTemplate(template(`{{ js (js (js (js (js "\\")))) }}`))
			So(err, ShouldBeError, "email:1:19: template nested too deep")

			err = v.ValidateHTMLTemplate(template(`
			{{ if true }}
				{{ if true }}
					{{ if true }}
						{{ if true }}
						{{end}}
					{{end}}
				{{end}}
			{{end}}`))
			So(err, ShouldBeError, "email:5:19: template nested too deep")
		})

		Convey("should allow range node if explicitly allowed", func() {
			var err error
			v := NewValidator(AllowRangeNode(true))

			err = v.ValidateHTMLTemplate(template(`{{ range . }}{{ end }}`))
			So(err, ShouldBeNil)
		})

		Convey("should traverse into template nodes pipe", func() {
			var err error
			v := NewValidator(AllowTemplateNode(true))

			err = v.ValidateHTMLTemplate((template(`{{ template "name" js (js (js (js (js "\\")))) }}`)))
			So(err, ShouldBeError, "email:1:31: template nested too deep")
		})
	})
}
