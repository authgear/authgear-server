package template_test

import (
	"fmt"
	"strings"
	"testing"
	texttemplate "text/template"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/template"
)

func TestAGTextTemplate(t *testing.T) {
	Convey("AGTextTemplate.Parse", t, func() {
		type testCase struct {
			tpl          string
			output       string
			childOutputs map[string]string
		}
		cases := []testCase{
			{
				tpl:    "{{ eq (not .a) }}",
				output: "{{eq (not .a) | _value_or_empty_string}}",
			},
			{
				tpl:    "{{ .a }}",
				output: "{{.a | _value_or_empty_string}}",
			},
			{
				tpl:    "{{ .a.b.c }}",
				output: "{{.a.b.c | _value_or_empty_string}}",
			},
			{
				tpl:    "{{ len .a.b.c }}",
				output: "{{len .a.b.c | _value_or_empty_string}}",
			},
			{
				tpl: `
{{ range .Items }}
{{ if .a }}
{{ .a }}
{{ end }}
{{ else }}
.
{{ end }}`,
				output: `
{{range .Items}}{{_record_iteration}}
{{if .a}}
{{.a | _value_or_empty_string}}
{{end}}
{{else}}{{_record_iteration}}
.
{{end}}`},
		}

		for idx, c := range cases {
			Convey(fmt.Sprintf("rewrite %d", idx), func() {
				tpl := &template.AGTextTemplate{}
				textTpl := &texttemplate.Template{}
				textTpl = texttemplate.Must(textTpl.Parse(c.tpl))
				err := tpl.Wrap(textTpl)
				So(err, ShouldBeNil)
				So(tpl.String(""), ShouldEqual, c.output)
				for tplName, childOutput := range c.childOutputs {
					So(tpl.String(tplName), ShouldEqual, childOutput)
				}
			})
		}
	})

	Convey("AGTextTemplate.Execute", t, func() {
		type testCase struct {
			tpl    string
			data   any
			output string
		}

		cases := []testCase{
			{
				tpl:    "{{.a}}",
				data:   map[string]interface{}{},
				output: "",
			},
			{
				tpl: "{{.a.b.c}}",
				data: map[string]interface{}{
					"a": map[string]interface{}{},
				},
				output: "",
			},
			{
				tpl: "{{.a}}",
				data: map[string]interface{}{
					"a": "<button/>",
				},
				output: "<button/>",
			},
			{
				tpl: "{{.f}}",
				data: map[string]interface{}{
					"f": 1.23,
				},
				output: "1.23",
			},
			{
				tpl: "{{.b}}",
				data: map[string]interface{}{
					"b": false,
				},
				output: "false",
			},
			{
				tpl: `{{if .a}}Test{{end}}`,
				data: map[string]interface{}{
					"a": false,
				},
				output: "",
			},
			{
				tpl: `{{if eq .a "A"}}Test{{end}}`,
				data: map[string]interface{}{
					"a": "A",
				},
				output: "Test",
			},
			{
				tpl: `{{range .items}}{{.a}}{{end}}`,
				data: map[string]interface{}{
					"items": []interface{}{map[string]interface{}{
						"a": 1,
					}, map[string]interface{}{
						"b": 2,
					}},
				},
				output: "1",
			},
			{
				tpl: "{{.ptr}}",
				data: map[string]interface{}{
					"ptr": func() *string {
						s := "p"
						return &s
					}(),
				},
				output: "p",
			},
		}

		for idx, c := range cases {
			Convey(fmt.Sprintf("execute %d", idx), func() {

				tpl := &template.AGTextTemplate{}
				textTpl := &texttemplate.Template{}
				textTpl = texttemplate.Must(textTpl.Parse(c.tpl))
				err := tpl.Wrap(textTpl)
				So(err, ShouldBeNil)
				var buf strings.Builder
				err = tpl.Execute(&buf, c.data)
				So(err, ShouldBeNil)
				So(buf.String(), ShouldEqual, c.output)
			})
		}

		Convey("should error if max iteration reached", func() {
			Convey("simple loop", func() {
				textTpl := &texttemplate.Template{}
				textTpl = texttemplate.Must(textTpl.Parse("{{range . }}.{{end}}"))
				tpl := &template.AGTextTemplate{}
				err := tpl.Wrap(textTpl)
				So(err, ShouldBeNil)
				var buf strings.Builder
				err = tpl.Execute(&buf, make([]string, 101))
				So(err, ShouldBeError, "template: :1:4: executing \"\" at <_record_iteration>: error calling _record_iteration: max iteration exceeded")

				err = tpl.Execute(&buf, make([]string, 100))
				So(err, ShouldBeNil)
			})
			Convey("nested loop", func() {

				// Go 1.23.6 does not support {{ range int }} yet, so we simulate it with a nested data
				makeNestedData := func(n int) [][]string {
					nestedData := [][]string{}
					for i := 0; i < n; i++ {
						el := []string{}
						for j := 0; j < n; j++ {
							el = append(el, ".")
						}
						nestedData = append(nestedData, el)
					}
					return nestedData
				}

				textTpl := &texttemplate.Template{}
				textTpl = texttemplate.Must(textTpl.Parse("{{range . }}{{range . }}.{{end}}{{end}}"))
				tpl := &template.AGTextTemplate{}
				err := tpl.Wrap(textTpl)
				So(err, ShouldBeNil)
				var buf strings.Builder
				err = tpl.Execute(&buf, makeNestedData(10))
				So(err, ShouldBeError, "template: :1:4: executing \"\" at <_record_iteration>: error calling _record_iteration: max iteration exceeded")

				err = tpl.Execute(&buf, makeNestedData(9))
				So(err, ShouldBeNil)
			})
		})
	})
}
