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
				tpl:    "{{ .a | html | js }}",
				output: "{{.a | html | js | _value_or_empty_string}}",
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
				tpl:    "{{ html .a.b.c }}",
				output: "{{html .a.b.c | _value_or_empty_string}}",
			},
			{
				tpl: `
{{ range .Items }}
{{ if .a }}
{{ .a }}
{{ end }}
{{ end }}`,
				output: `
{{range .Items}}
{{if .a}}
{{.a | _value_or_empty_string}}
{{end}}
{{end}}`},
			{
				tpl: `
{{ define "temp1" }}
	{{ if .a }}
		{{ .a }}
	{{ end }}
{{ end }}
{{ template "temp1" }}
`,
				output: "\n\n{{template \"temp1\"}}\n",
				childOutputs: map[string]string{
					"temp1": "\n\t{{if .a}}\n\t\t{{.a | _value_or_empty_string}}\n\t{{end}}\n",
				},
			},
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
				tpl: "{{.a|html}}",
				data: map[string]interface{}{
					"a": "<button/>",
				},
				output: "&lt;button/&gt;",
			},
			{
				tpl: "{{html .a}}",
				data: map[string]interface{}{
					"a": "<button/>",
				},
				output: "&lt;button/&gt;",
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
	})
}
