package hook

import (
	"context"
	"path"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestDenoFileDescriptor(t *testing.T) {
	Convey("DenoFileDescriptor", t, func() {
		builtin := afero.NewMemMapFs()
		app := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: builtin, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: app, FsLevel: resource.FsLevelApp},
		})

		r.Register(DenoFile)

		writeFile := func(fs afero.Fs, p string, content string) {
			err := fs.MkdirAll(path.Dir(p), 0777)
			So(err, ShouldBeNil)
			err = afero.WriteFile(fs, p, []byte(content), 0666)
			So(err, ShouldBeNil)
		}

		Convey("MatchResource", func() {
			test := func(path string, expected bool) {
				_, ok := DenoFile.MatchResource(path)
				So(ok, ShouldEqual, expected)
			}

			test("", false)
			test("deno", false)
			test("deno/some.ts", true)
		})

		Convey("Read", func() {
			writeFile(builtin, "deno/c.ts", "c.ts from builtin")
			writeFile(app, "deno/b.ts", "b.ts from app")
			writeFile(app, "deno/c.ts", "c.ts from app")

			out, err := manager.Read(context.Background(), DenoFile, resource.AppFile{
				Path: "deno/c.ts",
			})
			So(err, ShouldBeNil)
			So(out, ShouldResemble, []byte("c.ts from app"))
		})
	})
}
