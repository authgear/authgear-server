package resources_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/portal/util/resources"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestList(t *testing.T) {
	Convey("List", t, func() {
		reg := &resource.Registry{}
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		res := resource.NewManager(reg, []resource.Fs{
			&resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			&resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelApp},
		})

		reg.Register(resource.SimpleDescriptor{Path: "test/a/x.txt"})
		reg.Register(resource.SimpleDescriptor{Path: "test/b/z.txt"})
		reg.Register(resource.SimpleDescriptor{Path: "test/x.txt"})
		reg.Register(resource.SimpleDescriptor{Path: "w.txt"})

		_ = fsA.MkdirAll("test/a", 0666)
		_ = fsA.MkdirAll("test/b", 0666)
		_ = fsB.MkdirAll("test/a", 0666)
		_ = afero.WriteFile(fsA, "test/a/x.txt", nil, 0666)
		_ = afero.WriteFile(fsA, "test/a/y.txt", nil, 0666)
		_ = afero.WriteFile(fsA, "test/b/z.txt", nil, 0666)
		_ = afero.WriteFile(fsB, "test/x.txt", nil, 0666)
		_ = afero.WriteFile(fsB, "test/b/z.txt", nil, 0666)
		_ = afero.WriteFile(fsB, "w.txt", nil, 0666)

		paths, err := resources.List(res)
		So(err, ShouldBeNil)
		So(paths, ShouldResemble, []string{
			"test/a/x.txt",
			"test/b/z.txt",
			"test/x.txt",
			"w.txt",
		})
	})
}
