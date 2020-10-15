package template_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template2"
)

func TestTemplateResource(t *testing.T) {
	Convey("PlainText", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.AferoFs{Fs: fsA},
			resource.AferoFs{Fs: fsB},
		})

		txt := &template.HTML{Name: "resource.txt"}
		r.Register(txt)

		args := map[string]interface{}{
			template.ResourceArgDefaultLanguageTag: "en",
		}

		writeFile := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/resource.txt", []byte(data), 0666)
		}

		read := func() (string, string, error) {
			data, err := manager.Read(txt, args)
			if err != nil {
				return "", "", err
			}
			return data.Path, string(data.Data), nil
		}

		Convey("it should return single resource", func() {
			writeFile(fsA, "__default__", "default in fs A")

			path, data, err := read()
			So(err, ShouldBeNil)
			So(path, ShouldEqual, "templates/__default__/resource.txt")
			So(data, ShouldEqual, "default in fs A")
		})

		Convey("it should return resource with preferred language", func() {
			writeFile(fsA, "en", "en in fs A")
			writeFile(fsA, "zh", "zh in fs A")
			writeFile(fsA, "__default__", "default in fs A")

			path, data, err := read()
			So(err, ShouldBeNil)
			So(path, ShouldEqual, "templates/en/resource.txt")
			So(data, ShouldEqual, "en in fs A")

			args[template.ResourceArgPreferredLanguageTag] = []string{"en"}
			path, data, err = read()
			So(err, ShouldBeNil)
			So(path, ShouldEqual, "templates/en/resource.txt")
			So(data, ShouldEqual, "en in fs A")

			args[template.ResourceArgPreferredLanguageTag] = []string{"zh"}
			path, data, err = read()
			So(err, ShouldBeNil)
			So(path, ShouldEqual, "templates/zh/resource.txt")
			So(data, ShouldEqual, "zh in fs A")
		})

		Convey("it should combine resources in different FS", func() {
			writeFile(fsB, "zh", "zh in fs B")
			writeFile(fsA, "__default__", "default in fs A")

			path, data, err := read()
			So(err, ShouldBeNil)
			So(path, ShouldEqual, "templates/__default__/resource.txt")
			So(data, ShouldEqual, "default in fs A")

			args[template.ResourceArgPreferredLanguageTag] = []string{"en"}
			path, data, err = read()
			So(err, ShouldBeNil)
			So(path, ShouldEqual, "templates/__default__/resource.txt")
			So(data, ShouldEqual, "default in fs A")

			args[template.ResourceArgPreferredLanguageTag] = []string{"zh"}
			path, data, err = read()
			So(err, ShouldBeNil)
			So(path, ShouldEqual, "templates/zh/resource.txt")
			So(data, ShouldEqual, "zh in fs B")
		})
	})
}
