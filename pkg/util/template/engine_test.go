package template_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func TestEngine(t *testing.T) {
	Convey("Engine", t, func() {
		fs := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{resource.LeveledAferoFs{
			Fs:      fs,
			FsLevel: resource.FsLevelBuiltin,
		}})
		resolver := &template.Resolver{
			Resources:             manager,
			DefaultLanguageTag:    "en",
			SupportedLanguageTags: []string{"zh", "en"},
		}
		engine := &template.Engine{Resolver: resolver}

		header := &template.HTML{Name: "header.html"}
		footer := &template.HTML{Name: "footer.html"}
		pageA := &template.HTML{Name: "pageA.html", ComponentDependencies: []*template.HTML{header, footer}}
		pageB := &template.HTML{Name: "pageB.html", ComponentDependencies: []*template.HTML{header, footer}}
		index := &template.HTML{Name: "index.html", ComponentDependencies: []*template.HTML{pageA, pageB}}

		writeFile := func(lang string, name string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/"+name, []byte(data), 0666)
		}

		writeFile("en", "header.html", `en header`)
		writeFile("zh", "header.html", `zh header`)

		writeFile("en", "footer.html", `{{ template "footer-name" }}`)

		writeFile("en", "pageA.html",
			`{{ template "header.html" }};{{ template "a-title" }};{{ template "footer.html" }}`,
		)
		writeFile("en", "pageB.html",
			`{{ template "header.html" }};{{ template "b-title" }};{{ template "footer.html" }}`,
		)
		writeFile("en", "index.html", `{{ template "pageA.html" }}`)
		writeFile("zh", "index.html", `{{ template "pageB.html" }}`)

		writeFile("en", "translation.json", `{
			"footer-name": "en footer",
			"a-title": "en a title",
			"b-title": "en b title"
		}`)
		writeFile("zh", "translation.json", `{
			"footer-name": "zh footer",
			"b-title": "zh b title"
		}`)

		Convey("it should render correct localized template", func() {
			data, err := engine.Render(index, []string{}, nil)
			So(err, ShouldBeNil)
			So(data.String, ShouldEqual, "en header;en a title;en footer")

			data, err = engine.Render(index, []string{"en"}, nil)
			So(err, ShouldBeNil)
			So(data.String, ShouldEqual, "en header;en a title;en footer")

			data, err = engine.Render(index, []string{"zh"}, nil)
			So(err, ShouldBeNil)
			So(data.String, ShouldEqual, "zh header;zh b title;zh footer")
		})

		Convey("it should render correct localized translation", func() {
			m, err := engine.Translation([]string{"en"})
			So(err, ShouldBeNil)
			footer, err := m.RenderText("footer-name", nil)
			So(err, ShouldBeNil)
			So(footer, ShouldEqual, "en footer")
			a, err := m.RenderText("a-title", nil)
			So(err, ShouldBeNil)
			So(a, ShouldEqual, "en a title")
			b, err := m.RenderText("b-title", nil)
			So(err, ShouldBeNil)
			So(b, ShouldEqual, "en b title")

			m, err = engine.Translation([]string{"zh"})
			So(err, ShouldBeNil)
			footer, err = m.RenderText("footer-name", nil)
			So(err, ShouldBeNil)
			So(footer, ShouldEqual, "zh footer")
			a, err = m.RenderText("a-title", nil)
			So(err, ShouldBeNil)
			So(a, ShouldEqual, "en a title")
			b, err = m.RenderText("b-title", nil)
			So(err, ShouldBeNil)
			So(b, ShouldEqual, "zh b title")
		})
	})
}
