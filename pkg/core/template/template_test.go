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

			_, err = RenderTextTemplate("test", template, nil)
			So(err, ShouldBeError, "failed to execute template: rendered template is too large")
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

			out, err := RenderHTMLTemplate("test", template, map[string]interface{}{
				"URL":             "https://www.example.com",
				"Title":           "Welcome to <b>My App</b>.",
				"BackgroundColor": "#101020; /* for contrast */",
				"ForegroundColor": "#e0e0ff",
				"Query":           "Lazy dog > Quick brown fox?",
				"State": map[string]interface{}{
					"todos": []string{`<b>Important things! \o/</b>`, "Cats & Dogs"},
					"query": "Lazy dog > Quick brown fox?",
				},
			})
			So(err, ShouldBeNil)
			So(out, ShouldEqual, expectation)
		})
	})
}
