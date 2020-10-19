package model_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestAppListResources(t *testing.T) {
	Convey("App.ListResources", t, func() {
		reg := &resource.Registry{}
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		res := resource.NewManager(reg, []resource.Fs{
			&resource.AferoFs{Fs: fsA},
			&resource.AferoFs{Fs: fsB},
		})
		app := &model.App{
			Context: &config.AppContext{Resources: res},
		}

		reg.Register(resource.SimpleFile{Name: "test/a/x.txt"})
		reg.Register(resource.SimpleFile{Name: "test/b/z.txt"})
		reg.Register(resource.SimpleFile{Name: "test/x.txt"})
		reg.Register(resource.SimpleFile{Name: "w.txt"})

		_ = fsA.MkdirAll("/test/a", 0666)
		_ = fsA.MkdirAll("/test/b", 0666)
		_ = fsB.MkdirAll("/test/a", 0666)
		_ = afero.WriteFile(fsA, "/test/a/x.txt", nil, 0666)
		_ = afero.WriteFile(fsA, "/test/a/y.txt", nil, 0666)
		_ = afero.WriteFile(fsA, "/test/b/z.txt", nil, 0666)
		_ = afero.WriteFile(fsB, "/test/x.txt", nil, 0666)
		_ = afero.WriteFile(fsB, "/test/b/z.txt", nil, 0666)
		_ = afero.WriteFile(fsB, "/w.txt", nil, 0666)

		paths, err := app.ListResources()
		So(err, ShouldBeNil)
		So(paths, ShouldResemble, []string{
			"test/a/x.txt",
			"test/b/z.txt",
			"test/x.txt",
			"w.txt",
		})
	})
}
