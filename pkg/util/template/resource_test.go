package template_test

import (
	htmltemplate "html/template"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func TestTemplateResource(t *testing.T) {
	Convey("HTML EffectiveResource", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.AferoFs{Fs: fsA},
			resource.AferoFs{Fs: fsB},
		})

		txt := &template.HTML{Name: "resource.txt"}
		r.Register(txt)

		view := resource.EffectiveResource{
			DefaultTag: "en",
		}

		writeFile := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/resource.txt", []byte(data), 0666)
		}

		read := func() (str string, err error) {
			result, err := manager.Read(txt, view)
			if err != nil {
				return
			}

			tpl := result.(*htmltemplate.Template)
			var out strings.Builder
			err = tpl.Execute(&out, nil)
			if err != nil {
				return
			}

			str = out.String()
			return
		}

		Convey("it should return single resource", func() {
			writeFile(fsA, "__default__", "default in fs A")

			data, err := read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "default in fs A")
		})

		Convey("it should return resource with preferred language", func() {
			writeFile(fsA, "en", "en in fs A")
			writeFile(fsA, "zh", "zh in fs A")
			writeFile(fsA, "__default__", "default in fs A")

			data, err := read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")

			view.PreferredTags = []string{"en"}
			data, err = read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")

			view.PreferredTags = []string{"zh"}
			data, err = read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "zh in fs A")
		})

		Convey("it should combine resources in different FS", func() {
			writeFile(fsB, "zh", "zh in fs B")
			writeFile(fsA, "__default__", "default in fs A")

			data, err := read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "default in fs A")

			view.PreferredTags = []string{"en"}
			data, err = read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "default in fs A")

			view.PreferredTags = []string{"zh"}
			data, err = read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "zh in fs B")
		})
	})

	Convey("HTML EffectiveFile", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.AferoFs{Fs: fsA},
			resource.AferoFs{Fs: fsB},
		})

		txt := &template.HTML{Name: "resource.txt"}
		r.Register(txt)

		writeFile := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/resource.txt", []byte(data), 0666)
		}

		read := func(lang string) (str string, err error) {
			view := resource.EffectiveFile{
				Path:       "templates/" + lang + "/resource.txt",
				DefaultTag: "en",
			}
			result, err := manager.Read(txt, view)
			if err != nil {
				return
			}

			str = string(result.([]byte))
			return
		}

		Convey("it should return single resource", func() {
			writeFile(fsA, "__default__", "default in fs A")

			data, err := read("__default__")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "default in fs A")

			data, err = read("en")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "default in fs A")

			_, err = read("zh")
			So(err, ShouldBeError, "specified resource is not configured")
		})

		Convey("it should return resource with preferred language", func() {
			writeFile(fsA, "en", "en in fs A")
			writeFile(fsA, "zh", "zh in fs A")
			writeFile(fsA, "__default__", "default in fs A")

			data, err := read("en")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")

			data, err = read("zh")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "zh in fs A")
		})

		Convey("it should combine resources in different FS", func() {
			writeFile(fsB, "zh", "zh in fs B")
			writeFile(fsA, "__default__", "default in fs A")

			data, err := read("en")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "default in fs A")

			data, err = read("zh")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "zh in fs B")
		})
	})
}
