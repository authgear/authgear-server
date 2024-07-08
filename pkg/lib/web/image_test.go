package web_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"mime"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestExtensionsByType(t *testing.T) {
	Convey("mime.ExtensionsByType", t, func() {
		exts, err := mime.ExtensionsByType("image/png")
		So(err, ShouldBeNil)
		So(exts, ShouldResemble, []string{".png"})

		exts, err = mime.ExtensionsByType("image/jpeg")
		So(err, ShouldBeNil)
		So(exts, ShouldResemble, []string{".jpe", ".jpeg", ".jpg"})

		exts, err = mime.ExtensionsByType("image/gif")
		So(err, ShouldBeNil)
		So(exts, ShouldResemble, []string{".gif"})
	})
}

func TestTemplateResource(t *testing.T) {
	Convey("LocaleAwareImageDescriptor EffectiveResource", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelApp},
		})

		img := web.LocaleAwareImageDescriptor{
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
				DefaultTag:    "en",
				SupportedTags: []string{"en"},
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
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(asset.Path, ShouldEqual, "static/zh/myimage.png")
			So(asset.Data, ShouldResemble, pngB)
		})

		Convey("it should use more specific resource", func() {
			writeFile(fsA, "en", ".png", pngA)
			writeFile(fsB, "en", ".png", pngB)

			asset, err := read(resource.EffectiveResource{
				DefaultTag:    "en",
				SupportedTags: []string{"en"},
			})
			So(err, ShouldBeNil)
			So(asset.Path, ShouldEqual, "static/en/myimage.png")
			So(asset.Data, ShouldResemble, pngB)
		})

		Convey("it should not fail when fallback is not en", func() {
			writeFile(fsA, "en", ".png", pngA)

			asset, err := read(resource.EffectiveResource{
				DefaultTag:    "zh",
				SupportedTags: []string{"zh"},
			})
			So(err, ShouldBeNil)
			So(asset.Path, ShouldEqual, "static/en/myimage.png")
			So(asset.Data, ShouldResemble, pngA)
		})

		Convey("it should use fallback language in the app fs first", func() {
			writeFile(fsA, "en", ".png", pngA)
			writeFile(fsB, "jp", ".png", pngB)

			asset, err := read(resource.EffectiveResource{
				SupportedTags: []string{"en", "jp"},
				DefaultTag:    "jp",
				PreferredTags: []string{"en"},
			})
			So(err, ShouldBeNil)
			So(asset.Path, ShouldEqual, "static/jp/myimage.png")
			So(asset.Data, ShouldResemble, pngB)

			asset, err = read(resource.EffectiveResource{
				SupportedTags: []string{"jp", "zh"},
				DefaultTag:    "jp",
				PreferredTags: []string{"fr"},
			})
			So(err, ShouldBeNil)
			So(asset.Path, ShouldEqual, "static/jp/myimage.png")
			So(asset.Data, ShouldResemble, pngB)
		})

		Convey("it should disallow duplicate resource", func() {
			writeFile(fsA, "en", ".png", pngA)
			writeFile(fsB, "en", ".png", pngB)
			writeFile(fsB, "en", ".jpg", pngB)

			_, err := read(resource.ValidateResource{})
			So(err, ShouldBeError, "duplicate resource: [static/en/myimage.jpg static/en/myimage.png]")
		})
	})

	Convey("LocaleAwareImageDescriptor EffectiveFile", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelApp},
		})

		img := web.LocaleAwareImageDescriptor{
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

	Convey("LocaleAwareImageDescriptor AppFile", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelApp},
		})

		img := web.LocaleAwareImageDescriptor{
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

func TestNonLocaleAwareImageDescriptor(t *testing.T) {
	Convey("NonLocaleAwareImageDescriptor EffectiveResource", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelApp},
		})

		img := web.NonLocaleAwareImageDescriptor{
			Name: "myimage",
		}
		r.Register(img)

		writeFile := func(fs afero.Fs, ext string, data []byte) {
			_ = fs.MkdirAll("static/", 0777)
			_ = afero.WriteFile(fs, "static/myimage"+ext, data, 0666)
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
			writeFile(fsA, ".png", pngA)

			asset, err := read(resource.EffectiveResource{
				PreferredTags: []string{"zh", "en"},
				DefaultTag:    "en",
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(asset.Path, ShouldEqual, "static/myimage.png")
			So(asset.Data, ShouldResemble, pngA)
		})

		Convey("it should use more specific resource", func() {
			writeFile(fsA, ".png", pngA)
			writeFile(fsB, ".png", pngB)

			asset, err := read(resource.EffectiveResource{
				DefaultTag:    "en",
				SupportedTags: []string{"en"},
			})
			So(err, ShouldBeNil)
			So(asset.Path, ShouldEqual, "static/myimage.png")
			So(asset.Data, ShouldResemble, pngB)
		})

		Convey("it should not fail when fallback is not en", func() {
			writeFile(fsA, ".png", pngA)

			asset, err := read(resource.EffectiveResource{
				DefaultTag:    "zh",
				SupportedTags: []string{"zh"},
			})
			So(err, ShouldBeNil)
			So(asset.Path, ShouldEqual, "static/myimage.png")
			So(asset.Data, ShouldResemble, pngA)
		})

		Convey("it should disallow duplicate resource", func() {
			writeFile(fsA, ".png", pngA)
			writeFile(fsB, ".png", pngB)
			writeFile(fsB, ".jpg", pngB)

			_, err := read(resource.ValidateResource{})
			So(err, ShouldBeError, "duplicate resource: [static/myimage.jpg static/myimage.png]")
		})
	})

	Convey("NonLocaleAwareImageDescriptor EffectiveFile", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelApp},
		})

		img := web.NonLocaleAwareImageDescriptor{
			Name: "myimage",
		}
		r.Register(img)

		writeFile := func(fs afero.Fs, ext string, data []byte) {
			_ = fs.MkdirAll("static/", 0777)
			_ = afero.WriteFile(fs, "static/myimage"+ext, data, 0666)
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
			writeFile(fsA, ".png", pngA)

			data, err := read(resource.EffectiveFile{
				Path: "static/myimage.png",
			})
			So(err, ShouldBeNil)
			So(data, ShouldResemble, pngA)

			_, err = read(resource.EffectiveFile{
				Path: "static/myimage.jpg",
			})
			So(err, ShouldBeError, "specified resource is not configured")
		})

		Convey("it should return resource with specific path", func() {
			writeFile(fsA, ".jpg", pngA)
			writeFile(fsA, ".png", pngB)

			data, err := read(resource.EffectiveFile{
				Path: "static/myimage.png",
			})
			So(err, ShouldBeNil)
			So(data, ShouldResemble, pngB)

			data, err = read(resource.EffectiveFile{
				Path: "static/myimage.jpg",
			})
			So(err, ShouldBeNil)
			So(data, ShouldResemble, pngA)
		})

		Convey("it should use more specific resource", func() {
			writeFile(fsA, ".png", pngA)
			writeFile(fsB, ".png", pngB)

			data, err := read(resource.EffectiveFile{
				Path: "static/myimage.png",
			})
			So(err, ShouldBeNil)
			So(data, ShouldResemble, pngB)
		})
	})

	Convey("NonLocaleAwareImageDescriptor AppFile", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelApp},
		})

		img := web.NonLocaleAwareImageDescriptor{
			Name: "myimage",
		}
		r.Register(img)

		writeFile := func(fs afero.Fs, ext string, data []byte) {
			_ = fs.MkdirAll("static/", 0777)
			_ = afero.WriteFile(fs, "static/myimage"+ext, data, 0666)
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
			writeFile(fsA, ".png", pngA)

			_, err := read(resource.AppFile{
				Path: "static/myimage.png",
			})
			So(err, ShouldBeError, "specified resource is not configured")
		})

		Convey("found", func() {
			writeFile(fsB, ".png", pngB)

			data, err := read(resource.AppFile{
				Path: "static/myimage.png",
			})
			So(err, ShouldBeNil)
			So(data, ShouldResemble, pngB)
		})
	})
}

func TestImageSizeLimit(t *testing.T) {
	Convey("LocaleAwareImageDescriptor size limit", t, func() {
		image := web.LocaleAwareImageDescriptor{
			Name:      "image",
			SizeLimit: 100 * 1024 * 1024,
		}
		So(image.GetSizeLimit(), ShouldEqual, 100*1024*1024)
	})
}
