package template

import (
	"html/template"
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

			err = ValidateHTMLTemplate(template(`{{ if ne .UserName "" }}Welcome, {{ .UserName }}{{ else }}Please login{{ end }}`))
			So(err, ShouldBeNil)
		})

		Convey("should not allow disabled constructs", func() {
			var err error

			err = ValidateHTMLTemplate(template(`{{ range $i, $e := . }}{{$i}}{{$e}}{{ end }}`))
			So(err, ShouldBeError, "template: email:1:9: forbidden construct *parse.RangeNode")

			err = ValidateHTMLTemplate(template(`{{block "name" ""}} Test {{ template "name" }} {{end}}`))
			So(err, ShouldBeError, "template: email:1:8: forbidden construct *parse.TemplateNode")

			err = ValidateHTMLTemplate(template(`
			{{ with $v := js "\\" }}
				{{ with $v := js $v }}
					{{ with $v := js $v }}
						{{ with $v := js $v }}
							{{$v}}
						{{end}}
					{{end}}
				{{end}}
			{{end}}`))
			So(err, ShouldBeError, "template: email:2:11: forbidden construct *parse.WithNode")
		})

		Convey("should not allow disabled functions", func() {
			var err error

			err = ValidateHTMLTemplate(template(`{{printf "%010000000d" 0}}`))
			So(err, ShouldBeError, "template: email:1:2: forbidden identifier printf")
		})

		Convey("should not allow nesting too deep", func() {
			var err error

			err = ValidateHTMLTemplate(template(`{{ "\\" | js | js | js }}`))
			So(err, ShouldBeNil)

			err = ValidateHTMLTemplate(template(`{{ "\\" | js | js | js | js }}`))
			So(err, ShouldBeError, "template: email:1:3: pipeline is too long")

			err = ValidateHTMLTemplate(template(`{{ js (js (js (js "\\"))) }}`))
			So(err, ShouldBeNil)

			err = ValidateHTMLTemplate(template(`{{ js (js (js (js (js "\\")))) }}`))
			So(err, ShouldBeError, "template: email:1:19: template nested too deep")

			err = ValidateHTMLTemplate(template(`
			{{ if $v := js "\\" }}
				{{ if $v := js $v }}
					{{ if $v := js $v }}
						{{ if $v := js $v }}
							{{$v}}
						{{end}}
					{{end}}
				{{end}}
			{{end}}`))
			So(err, ShouldBeError, "template: email:5:26: template nested too deep")
		})
	})
}
