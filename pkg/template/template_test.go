package template

import (
	"fmt"
	"net/url"
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

			_, err = RenderTextTemplate(RenderOptions{
				Name:         "test",
				TemplateBody: template,
				Context:      nil,
			})
			So(err, ShouldBeError, "failed to execute template: rendered template is too large")
		})
		Convey("should supports defines", func() {
			actual, err := RenderHTMLTemplate(RenderOptions{
				Name: "test",
				TemplateBody: `
				{{ template "A" }}
				{{ template "B" }}
				`,
				Context: map[string]interface{}{
					"a": "42",
				},
				Defines: []string{
					`{{ define "A" }}This is A{{ end }}`,
					`{{ define "B" }}This is B{{ end }}`,
				},
				ValidatorOpts: []ValidatorOption{AllowTemplateNode(true)},
			})
			expected := `
				This is A
				This is B
				`
			So(err, ShouldBeNil)
			So(actual, ShouldEqual, expected)
		})
		Convey("should supports funcs", func() {
			actual, err := RenderHTMLTemplate(RenderOptions{
				Name: "test",
				TemplateBody: `
				{{ localize "key" "string" 1 true .foobar }}
				`,
				Context: map[string]interface{}{
					"foobar": 42,
				},
				Funcs: map[string]interface{}{
					"localize": func(key string, args ...interface{}) (string, error) {
						buf := &strings.Builder{}
						for i, arg := range args {
							if i != 0 {
								buf.WriteRune(' ')
							}
							buf.WriteString(fmt.Sprintf("%v", arg))
						}
						return buf.String(), nil
					},
				},
			})
			expected := `
				string 1 true 42
				`
			So(err, ShouldBeNil)
			So(actual, ShouldEqual, expected)
		})
		Convey("should auto-escape templates", func() {
			template := `
			<!DOCTYPE html>
			<html lang="en">
			
			<head>
				<meta charset="utf-8" />
				<title>My App</title>
				<style type="text/css">
					body {
						background-color: {{ .BackgroundColor }};
						color: {{ .ForegroundColor }};
					}
				</style>
			</head>
			
			<body>
				<div id="root">
					<h1>{{ .Title }}</h1>
					<a href="/?query={{ .Query }}">Search for "{{ .Query }}"</a>
					<ul>
						<li>{{ index .State.todos 0 }}</li>
						<li>{{ index .State.todos 1 }}</li>
					</ul>
				</div>
				<script src="/app.js"></script>
				<script>
					const state = {{ .State }};
					renderApp(state);
				</script>
			</body>
			
			</html>
			`
			expectation := `
			<!DOCTYPE html>
			<html lang="en">
			
			<head>
				<meta charset="utf-8" />
				<title>My App</title>
				<style type="text/css">
					body {
						background-color: ZgotmplZ;
						color: #e0e0ff;
					}
				</style>
			</head>
			
			<body>
				<div id="root">
					<h1>Welcome to &lt;b&gt;My App&lt;/b&gt;.</h1>
					<a href="/?query=Lazy%20dog%20%3e%20Quick%20brown%20fox%3f">Search for "Lazy dog &gt; Quick brown fox?"</a>
					<ul>
						<li>&lt;b&gt;Important things! \o/&lt;/b&gt;</li>
						<li>Cats &amp; Dogs</li>
					</ul>
				</div>
				<script src="/app.js"></script>
				<script>
					const state = {"query":"Lazy dog \u003e Quick brown fox?","todos":["\u003cb\u003eImportant things! \\o/\u003c/b\u003e","Cats \u0026 Dogs"]};
					renderApp(state);
				</script>
			</body>
			
			</html>
			`

			out, err := RenderHTMLTemplate(RenderOptions{
				Name:         "test",
				TemplateBody: template,
				Context: map[string]interface{}{
					"URL":             "https://www.example.com",
					"Title":           "Welcome to <b>My App</b>.",
					"BackgroundColor": "#101020; /* for contrast */",
					"ForegroundColor": "#e0e0ff",
					"Query":           "Lazy dog > Quick brown fox?",
					"State": map[string]interface{}{
						"todos": []string{`<b>Important things! \o/</b>`, "Cats & Dogs"},
						"query": "Lazy dog > Quick brown fox?",
					},
				},
			})
			So(err, ShouldBeNil)
			So(out, ShouldEqual, expectation)
		})
	})
}

func TestEncodeContextToURLQueryParamValue(t *testing.T) {
	Convey("EncodeContextToURLQueryParamValue", t, func() {
		cases := []struct {
			Input map[string]interface{}
		}{
			{nil},
			{map[string]interface{}{}},
			{map[string]interface{}{
				"a": "b",
				"nested": map[string]interface{}{
					"a": "b",
				},
			}},
		}
		for _, c := range cases {
			encoded, err := EncodeContextToURLQueryParamValue(c.Input)
			So(err, ShouldBeNil)
			decoded, err := DecodeURLQueryParamValueToContext(encoded)
			So(err, ShouldBeNil)
			So(decoded, ShouldResemble, c.Input)
		}
	})
}

func TestSetContextToURLQuery(t *testing.T) {
	Convey("SetContextToURLQuery", t, func() {
		cases := []struct {
			URL      string
			Context  map[string]interface{}
			Expected string
		}{
			{
				"http://example.com",
				nil,
				"http://example.com?x-authgear-redirect-data=",
			},
			{
				"http://example.com",
				map[string]interface{}{},
				"http://example.com?x-authgear-redirect-data=e30",
			},
			{
				"http://example.com",
				map[string]interface{}{
					"a": "b",
				},
				"http://example.com?x-authgear-redirect-data=eyJhIjoiYiJ9",
			},
			{
				"http://example.com",
				map[string]interface{}{
					"a": "b",
					"nested": map[string]interface{}{
						"a": "b",
					},
				},
				"http://example.com?x-authgear-redirect-data=eyJhIjoiYiIsIm5lc3RlZCI6eyJhIjoiYiJ9fQ",
			},
		}
		for _, c := range cases {
			u, err := url.Parse(c.URL)
			So(err, ShouldBeNil)
			err = SetContextToURLQuery(u, c.Context)
			So(err, ShouldBeNil)
			So(u.String(), ShouldEqual, c.Expected)
		}
	})
}
