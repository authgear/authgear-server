package resource_test

import (
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
			resource.AferoFs{Fs: fsA},
			resource.AferoFs{Fs: fsB},
		})

		resourceA := r.Register(resource.SimpleFile{Name: "resourceA.txt"})
		resourceB := r.Register(resource.SimpleFile{Name: "resources/B.txt"})
		resourceC := r.Register(resource.SimpleFile{Name: "resources/C.txt"})

		Convey("it should reject when no FS contains the specified resource", func() {
			_, err := manager.Read(resourceA, nil)
			So(err, ShouldBeError, "specified resource is not configured")
		})

		Convey("it should merge resources from FS layers", func() {
			_ = afero.WriteFile(fsA, "resourceA.txt", []byte("resource A in fs A"), 0666)
			_ = afero.WriteFile(fsB, "resourceA.txt", []byte("resource A in fs B"), 0666)
			_ = afero.WriteFile(fsA, "resources/B.txt", []byte("resource B in fs A"), 0666)
			_ = afero.WriteFile(fsB, "resources/C.txt", []byte("resource C in fs B"), 0666)

			data, err := manager.Read(resourceA, nil)
			So(err, ShouldBeNil)
			So(data, ShouldResemble, &resource.LayerFile{
				Path: "resourceA.txt",
				Data: []byte("resource A in fs B"),
			})

			data, err = manager.Read(resourceB, nil)
			So(err, ShouldBeNil)
			So(data, ShouldResemble, &resource.LayerFile{
				Path: "resources/B.txt",
				Data: []byte("resource B in fs A"),
			})

			data, err = manager.Read(resourceC, nil)
			So(err, ShouldBeNil)
			So(data, ShouldResemble, &resource.LayerFile{
				Path: "resources/C.txt",
				Data: []byte("resource C in fs B"),
			})
		})

		Convey("it should resolve resource descriptor from path", func() {
			desc, ok := manager.Resolve("resourceA.txt")
			So(ok, ShouldBeTrue)
			So(desc, ShouldResemble, resourceA)

			desc, ok = manager.Resolve("resources/B.txt")
			So(ok, ShouldBeTrue)
			So(desc, ShouldResemble, resourceB)

			desc, ok = manager.Resolve("resources/C.txt")
			So(ok, ShouldBeTrue)
			So(desc, ShouldResemble, resourceC)
		})
	})
}
