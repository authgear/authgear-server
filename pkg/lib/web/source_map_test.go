package web_test

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestSourceMapDescriptor(t *testing.T) {
	Convey("SourceMapDescriptor", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelApp},
		})

		myMap := web.SourceMapDescriptor{
			Path: "static/main.js.map",
		}
		r.Register(myMap)

		writeFile := func(fs afero.Fs, data string) {
			_ = fs.MkdirAll("static/", 0777)
			_ = afero.WriteFile(fs, "static/main.js.map", []byte(data), 0666)
		}

		read := func(view resource.View) (str string, err error) {
			result, err := manager.Read(myMap, view)
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
			writeFile(fsA, "var a = 1;")

			_, err := read(resource.AppFile{})
			So(err, ShouldBeError, "specified resource is not configured")
		})

		Convey("AppFileView: found", func() {
			writeFile(fsB, "var a = 1;")

			data, err := read(resource.AppFile{})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "var a = 1;")
		})

		Convey("it should return single resource", func() {
			writeFile(fsA, "var a = 1;")

			data, err := read(resource.EffectiveFile{})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "(function(){var a = 1;})();")
		})

		Convey("it should combine resources in different FS", func() {
			writeFile(fsA, "var a = 1;")
			writeFile(fsB, "var b = 2;")

			data, err := read(resource.EffectiveFile{})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "(function(){var a = 1;})();(function(){var b = 2;})();")
		})
	})
}
