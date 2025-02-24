package resource_test

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestResourceManager(t *testing.T) {
	Convey("ResourceManager", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelApp},
		})

		resourceA := r.Register(resource.NewlineJoinedDescriptor{
			Path: "resourceA.txt",
			Parse: func(data []byte) (interface{}, error) {
				return data, nil
			},
		})
		resourceB := r.Register(resource.SimpleDescriptor{Path: "resources/B.txt"})
		resourceC := r.Register(resource.SimpleDescriptor{Path: "resources/C.txt"})

		Convey("it should reject when no FS contains the specified resource", func() {
			_, err := manager.Read(context.Background(), resourceA, resource.EffectiveFile{})
			So(err, ShouldBeError, "specified resource is not configured")
		})

		Convey("it should merge resources from FS layers", func() {
			_ = afero.WriteFile(fsA, "resourceA.txt", []byte("resource A in fs A"), 0666)
			_ = afero.WriteFile(fsB, "resourceA.txt", []byte("resource A in fs B"), 0666)
			_ = afero.WriteFile(fsA, "resources/B.txt", []byte("resource B in fs A"), 0666)
			_ = afero.WriteFile(fsB, "resources/C.txt", []byte("resource C in fs B"), 0666)

			data, err := manager.Read(context.Background(), resourceA, resource.EffectiveFile{})
			So(err, ShouldBeNil)
			So(data, ShouldResemble, []byte("resource A in fs A\nresource A in fs B\n"))

			data, err = manager.Read(context.Background(), resourceB, resource.EffectiveFile{})
			So(err, ShouldBeNil)
			So(data, ShouldResemble, []byte("resource B in fs A"))

			data, err = manager.Read(context.Background(), resourceC, resource.EffectiveFile{})
			So(err, ShouldBeNil)
			So(data, ShouldResemble, []byte("resource C in fs B"))
		})

		Convey("it should resolve resource descriptor from path", func() {
			desc, ok := manager.Resolve("resourceA.txt")
			So(ok, ShouldBeTrue)
			So(desc, ShouldHaveSameTypeAs, resource.NewlineJoinedDescriptor{})

			desc, ok = manager.Resolve("resources/B.txt")
			So(ok, ShouldBeTrue)
			So(desc, ShouldResemble, resourceB)

			desc, ok = manager.Resolve("resources/C.txt")
			So(ok, ShouldBeTrue)
			So(desc, ShouldResemble, resourceC)
		})
	})
}
