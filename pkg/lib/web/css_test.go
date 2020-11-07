package web_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestCSSDescriptor(t *testing.T) {
	Convey("CSSDescriptor", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.AferoFs{Fs: fsA},
			resource.AferoFs{Fs: fsB},
		})

		myCSS := web.CSSDescriptor{
			Path: "static/main.css",
		}
		r.Register(myCSS)

		writeFile := func(fs afero.Fs, data string) {
			_ = fs.MkdirAll("static/", 0777)
			_ = afero.WriteFile(fs, "static/main.css", []byte(data), 0666)
		}

		read := func() (str string, err error) {
			result, err := manager.Read(myCSS, nil)
			if err != nil {
				return
			}
			asset := result.(*web.StaticAsset)
			str = string(asset.Data)
			return
		}

		Convey("it should return single resource", func() {
			writeFile(fsA, ".a { color: red; }")

			data, err := read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, ".a { color: red; } ")
		})

		Convey("it should combine resources in different FS", func() {
			writeFile(fsA, ".a { color: red; }")
			writeFile(fsB, ".a { color: blue; }")

			data, err := read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, ".a { color: red; } .a { color: blue; } ")
		})
	})
}
