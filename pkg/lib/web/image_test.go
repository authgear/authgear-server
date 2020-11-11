package web_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestTemplateResource(t *testing.T) {
	Convey("ImageDescriptor EffectiveResource", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.AferoFs{Fs: fsA},
			resource.AferoFs{Fs: fsB},
		})

		img := web.ImageDescriptor{
			Name: "myimage",
		}
		r.Register(img)

		writeFile := func(fs afero.Fs, lang string, ext string, data []byte) {
			_ = fs.MkdirAll("static/"+lang, 0777)
			_ = afero.WriteFile(fs, "static/"+lang+"/myimage"+ext, data, 0666)
		}

		read := func(view resource.View) (asset *web.StaticAsset, err error) {
			result, err := manager.Read(img, view)
			if err != nil {
				return
			}

			asset = result.(*web.StaticAsset)
			return
		}

		makeImage := func(c color.Color) image.Image {
			i := image.NewRGBA(image.Rect(0, 0, 1, 1))
			i.Set(0, 0, c)
			return i
		}

		encodePNG := func(i image.Image) ([]byte, error) {
			var buf bytes.Buffer
			err := png.Encode(&buf, i)
			if err != nil {
				return nil, err
			}
			return buf.Bytes(), nil
		}

		pngA, err := encodePNG(makeImage(color.RGBA{255, 0, 0, 255}))
		So(err, ShouldBeNil)

		pngB, err := encodePNG(makeImage(color.RGBA{0, 255, 0, 255}))
		So(err, ShouldBeNil)

		So(pngA, ShouldNotResemble, pngB)

		Convey("it should return single resource", func() {
			writeFile(fsA, "en", ".png", pngA)

			asset, err := read(resource.EffectiveResource{
				DefaultTag: "en",
			})
			So(err, ShouldBeNil)
			So(asset.Path, ShouldEqual, "static/en/myimage.png")
			So(asset.Data, ShouldResemble, pngA)
		})

		Convey("it should return resource with preferred language", func() {
			writeFile(fsA, "en", ".png", pngA)
			writeFile(fsA, "zh", ".png", pngB)

			asset, err := read(resource.EffectiveResource{
				PreferredTags: []string{"zh", "en"},
				DefaultTag:    "en",
			})
			So(err, ShouldBeNil)
			So(asset.Path, ShouldEqual, "static/zh/myimage.png")
			So(asset.Data, ShouldResemble, pngB)
		})

		Convey("it should use more specific resource", func() {
			writeFile(fsA, "en", ".png", pngA)
			writeFile(fsB, "en", ".png", pngB)

			asset, err := read(resource.EffectiveResource{
				DefaultTag: "en",
			})
			So(err, ShouldBeNil)
			So(asset.Path, ShouldEqual, "static/en/myimage.png")
			So(asset.Data, ShouldResemble, pngB)
		})
	})

	Convey("ImageDescriptor EffectiveFile", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.AferoFs{Fs: fsA},
			resource.AferoFs{Fs: fsB},
		})

		img := web.ImageDescriptor{
			Name: "myimage",
		}
		r.Register(img)

		writeFile := func(fs afero.Fs, lang string, ext string, data []byte) {
			_ = fs.MkdirAll("static/"+lang, 0777)
			_ = afero.WriteFile(fs, "static/"+lang+"/myimage"+ext, data, 0666)
		}

		read := func(view resource.View) (b []byte, err error) {
			result, err := manager.Read(img, view)
			if err != nil {
				return
			}

			b = result.([]byte)
			return
		}

		makeImage := func(c color.Color) image.Image {
			i := image.NewRGBA(image.Rect(0, 0, 1, 1))
			i.Set(0, 0, c)
			return i
		}

		encodePNG := func(i image.Image) ([]byte, error) {
			var buf bytes.Buffer
			err := png.Encode(&buf, i)
			if err != nil {
				return nil, err
			}
			return buf.Bytes(), nil
		}

		pngA, err := encodePNG(makeImage(color.RGBA{255, 0, 0, 255}))
		So(err, ShouldBeNil)

		pngB, err := encodePNG(makeImage(color.RGBA{0, 255, 0, 255}))
		So(err, ShouldBeNil)

		So(pngA, ShouldNotResemble, pngB)

		Convey("it should return single resource", func() {
			writeFile(fsA, "en", ".png", pngA)

			data, err := read(resource.EffectiveFile{
				Path: "static/en/myimage.png",
			})
			So(err, ShouldBeNil)
			So(data, ShouldResemble, pngA)

			_, err = read(resource.EffectiveFile{
				Path: "static/zh/myimage.png",
			})
			So(err, ShouldBeError, "specified resource is not configured")

			_, err = read(resource.EffectiveFile{
				Path: "static/myimage.png",
			})
			So(err, ShouldBeError, "specified resource is not configured")
		})

		Convey("it should return resource with specific path", func() {
			writeFile(fsA, "en", ".png", pngA)
			writeFile(fsA, "zh", ".png", pngB)

			data, err := read(resource.EffectiveFile{
				Path: "static/en/myimage.png",
			})
			So(err, ShouldBeNil)
			So(data, ShouldResemble, pngA)

			data, err = read(resource.EffectiveFile{
				Path: "static/zh/myimage.png",
			})
			So(err, ShouldBeNil)
			So(data, ShouldResemble, pngB)
		})

		Convey("it should use more specific resource", func() {
			writeFile(fsA, "en", ".png", pngA)
			writeFile(fsB, "en", ".png", pngB)

			data, err := read(resource.EffectiveFile{
				Path: "static/en/myimage.png",
			})
			So(err, ShouldBeNil)
			So(data, ShouldResemble, pngB)
		})
	})

	Convey("ImageDescriptor AppFile", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.AferoFs{Fs: fsA},
			resource.AferoFs{Fs: fsB, IsAppFs: true},
		})

		img := web.ImageDescriptor{
			Name: "myimage",
		}
		r.Register(img)

		writeFile := func(fs afero.Fs, lang string, ext string, data []byte) {
			_ = fs.MkdirAll("static/"+lang, 0777)
			_ = afero.WriteFile(fs, "static/"+lang+"/myimage"+ext, data, 0666)
		}

		read := func(view resource.View) (b []byte, err error) {
			result, err := manager.Read(img, view)
			if err != nil {
				return
			}

			b = result.([]byte)
			return
		}

		makeImage := func(c color.Color) image.Image {
			i := image.NewRGBA(image.Rect(0, 0, 1, 1))
			i.Set(0, 0, c)
			return i
		}

		encodePNG := func(i image.Image) ([]byte, error) {
			var buf bytes.Buffer
			err := png.Encode(&buf, i)
			if err != nil {
				return nil, err
			}
			return buf.Bytes(), nil
		}

		pngA, err := encodePNG(makeImage(color.RGBA{255, 0, 0, 255}))
		So(err, ShouldBeNil)

		pngB, err := encodePNG(makeImage(color.RGBA{0, 255, 0, 255}))
		So(err, ShouldBeNil)

		So(pngA, ShouldNotResemble, pngB)

		Convey("not found", func() {
			writeFile(fsA, "en", ".png", pngA)

			_, err := read(resource.AppFile{
				Path: "static/en/myimage.png",
			})
			So(err, ShouldBeError, "specified resource is not configured")
		})

		Convey("found", func() {
			writeFile(fsB, "en", ".png", pngB)

			data, err := read(resource.AppFile{
				Path: "static/en/myimage.png",
			})
			So(err, ShouldBeNil)
			So(data, ShouldResemble, pngB)
		})
	})
}
