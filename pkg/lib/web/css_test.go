package web_test

import (
	"fmt"
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
			resource.AferoFs{Fs: fsB, IsAppFs: true},
		})

		myCSS := web.CSSDescriptor{
			Path: "static/main.css",
		}
		r.Register(myCSS)

		writeFile := func(fs afero.Fs, data string) {
			_ = fs.MkdirAll("static/", 0777)
			_ = afero.WriteFile(fs, "static/main.css", []byte(data), 0666)
		}

		read := func(view resource.View) (str string, err error) {
			result, err := manager.Read(myCSS, view)
			if err != nil {
				return
			}
			switch v := result.(type) {
			case *web.StaticAsset:
				str = string(v.Data)
			case []byte:
				str = string(v)
			default:
				panic(fmt.Errorf("unexpected type %T", result))
			}
			return
		}

		Convey("AppFileView: not found", func() {
			writeFile(fsA, ".a { color: red; }")

			_, err := read(resource.AppFile{})
			So(err, ShouldBeError, "specified resource is not configured")
		})

		Convey("AppFileView: found", func() {
			writeFile(fsB, ".a { color: red; }")

			data, err := read(resource.AppFile{})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, ".a { color: red; }")
		})

		Convey("it should return single resource", func() {
			writeFile(fsA, ".a { color: red; }")

			data, err := read(resource.EffectiveFile{})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, ".a { color: red; } ")
		})

		Convey("it should combine resources in different FS", func() {
			writeFile(fsA, ".a { color: red; }")
			writeFile(fsB, ".a { color: blue; }")

			data, err := read(resource.EffectiveFile{})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, ".a { color: red; } .a { color: blue; } ")
		})
	})
}
