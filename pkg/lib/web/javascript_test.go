package web_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestJavaScriptDescriptor(t *testing.T) {
	Convey("JavaScriptDescriptor", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.AferoFs{Fs: fsA},
			resource.AferoFs{Fs: fsB},
		})

		myJS := web.JavaScriptDescriptor{
			Path: "static/main.js",
		}
		r.Register(myJS)

		writeFile := func(fs afero.Fs, data string) {
			_ = fs.MkdirAll("static/", 0777)
			_ = afero.WriteFile(fs, "static/main.js", []byte(data), 0666)
		}

		read := func() (str string, err error) {
			result, err := manager.Read(myJS, nil)
			if err != nil {
				return
			}
			asset := result.(*web.StaticAsset)
			str = string(asset.Data)
			return
		}

		Convey("it should return single resource", func() {
			writeFile(fsA, "var a = 1;")

			data, err := read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "(function(){var a = 1;})();")
		})

		Convey("it should combine resources in different FS", func() {
			writeFile(fsA, "var a = 1;")
			writeFile(fsB, "var b = 2;")

			data, err := read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "(function(){var a = 1;})();(function(){var b = 2;})();")
		})
	})
}
