package template

import (
	"bytes"
	"html/template"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFuncs(t *testing.T) {
	date := time.Date(2006, 1, 2, 3, 4, 5, 0, time.UTC)

	Convey("RFC3339", t, func() {
		Convey("it supports time.Time", func() {
			So(RFC3339(date), ShouldEqual, "2006-01-02T03:04:05Z")
		})
		Convey("it supports *time.Time", func() {
			So(RFC3339(&date), ShouldEqual, "2006-01-02T03:04:05Z")
		})
		Convey("it does not fail for other data type", func() {
			So(RFC3339(nil), ShouldEqual, "INVALID_DATE")
			So(RFC3339(false), ShouldEqual, "INVALID_DATE")
			So(RFC3339(0), ShouldEqual, "INVALID_DATE")
			So(RFC3339(0.0), ShouldEqual, "INVALID_DATE")
			So(RFC3339(""), ShouldEqual, "INVALID_DATE")
			So(RFC3339(struct{}{}), ShouldEqual, "INVALID_DATE")
			So(RFC3339([]struct{}{}), ShouldEqual, "INVALID_DATE")
		})
	})

	Convey("IsNil", t, func() {
		So(IsNil(nil), ShouldBeTrue)

		var p *int64
		So(IsNil(p), ShouldBeTrue)

		var v int64
		p = &v
		So(IsNil(p), ShouldBeFalse)
		So(IsNil(v), ShouldBeFalse)

		p = nil
		So(IsNil(p), ShouldBeTrue)
	})

	Convey("ShowAttributeValue", t, func() {
		newString := func(s string) *string {
			return &s
		}

		newInt := func(i int64) *int64 {
			return &i
		}

		newFloat := func(f float64) *float64 {
			return &f
		}

		So(ShowAttributeValue(nil), ShouldEqual, "")
		So(ShowAttributeValue(1), ShouldEqual, "1")
		So(ShowAttributeValue(1.2), ShouldEqual, "1.2")
		So(ShowAttributeValue(100000000), ShouldEqual, "100000000")
		So(ShowAttributeValue(100000000.002), ShouldEqual, "100000000.002")
		So(ShowAttributeValue("test"), ShouldEqual, "test")

		var ip *int64
		So(ShowAttributeValue(ip), ShouldEqual, "")

		ip = newInt(100000000)
		So(ShowAttributeValue(ip), ShouldEqual, "100000000")

		var f32 float32 = 0.00002
		So(ShowAttributeValue(f32), ShouldEqual, "0.00002")

		var f64 float64 = 0.00002
		So(ShowAttributeValue(f64), ShouldEqual, "0.00002")

		var fp *float64
		So(ShowAttributeValue(fp), ShouldEqual, "")

		fp = newFloat(0)
		So(ShowAttributeValue(fp), ShouldEqual, "0")

		fp = newFloat(100000000.01)
		So(ShowAttributeValue(fp), ShouldEqual, "100000000.01")

		var sp *string
		So(ShowAttributeValue(sp), ShouldEqual, "")

		sp = newString("test")
		So(ShowAttributeValue(sp), ShouldEqual, "test")

	})

	Convey("include", t, func() {
		Convey("it supports variable template name", func() {
			tmpl := template.New("")
			funcMap := MakeTemplateFuncMap(tmpl)
			tmpl = tmpl.Funcs(funcMap)
			tmpl, err := tmpl.Parse(`
			{{- define "temp1" -}}
			content-of-temp1
			{{- end -}}
			{{- define "temp2" -}}
			content-of-temp2
			{{- end -}}
			{{- $tmplName := .TemplateName -}}
			{{- include $tmplName nil -}}`)
			if err != nil {
				panic(err)
			}

			buf := &bytes.Buffer{}
			err = tmpl.Execute(buf, map[string]string{"TemplateName": "temp2"})
			if err != nil {
				panic(err)
			}

			So(buf.String(), ShouldEqual, "content-of-temp2")

		})

		Convey("it supports html", func() {
			tmpl := template.New("")
			funcMap := MakeTemplateFuncMap(tmpl)
			tmpl = tmpl.Funcs(funcMap)
			tmpl, err := tmpl.Parse(`
			{{- define "some-html" -}}
			<span>content</span>
			{{- end -}}
			<div>
			{{- include "some-html" nil -}}
			</div>`)
			if err != nil {
				panic(err)
			}

			buf := &bytes.Buffer{}
			err = tmpl.Execute(buf, nil)
			if err != nil {
				panic(err)
			}

			So(buf.String(), ShouldEqual, "<div><span>content</span></div>")

		})

		Convey("it supports nesting", func() {
			tmpl := template.New("")
			funcMap := MakeTemplateFuncMap(tmpl)
			tmpl = tmpl.Funcs(funcMap)

			tmpl, err := tmpl.Parse(`
			{{- define "span" -}}
			<span>{{ .Children }}</span>
			{{- end -}}

			{{- define "div" -}}
			<div>{{ .Children }}</div>
			{{- end -}}

			{{- template "div" (dict
				"Children" (include "span" (dict
					"Children" (include "div" (dict
						"Children" "content"
					))
				))
			) -}}
			`)
			So(err, ShouldBeNil)

			buf := &bytes.Buffer{}
			err = tmpl.Execute(buf, nil)
			So(err, ShouldBeNil)
			So(buf.String(), ShouldEqual, "<div><span><div>content</div></span></div>")
		})
	},
	)

	Convey("trimHTML", t, func() {
		Convey("With string", func() {
			So(trimHTML(" A B "), ShouldEqual, "A B")
		})

		Convey("With HTML", func() {
			So(trimHTML(template.HTML("  A  B ")), ShouldEqual, template.HTML("A  B"))
		})
	})
}
