package template_test

import (
	htmltemplate "html/template"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func TestTemplateResource(t *testing.T) {
	Convey("HTML EffectiveResource", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelCustom},
		})

		txt := &template.HTML{Name: "resource.txt"}
		r.Register(txt)

		writeFile := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/resource.txt", []byte(data), 0666)
		}

		read := func(view resource.View) (str string, err error) {
			result, err := manager.Read(txt, view)
			if err != nil {
				return
			}

			tpl := result.(*htmltemplate.Template)
			var out strings.Builder
			err = tpl.Execute(&out, nil)
			if err != nil {
				return
			}

			str = out.String()
			return
		}

		Convey("it should return single resource", func() {
			writeFile(fsA, "en", "en in fs A")

			data, err := read(resource.EffectiveResource{
				DefaultTag:    "en",
				SupportedTags: []string{"en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")
		})

		Convey("it should return resource with preferred language", func() {
			writeFile(fsA, "en", "en in fs A")
			writeFile(fsA, "zh", "zh in fs A")

			data, err := read(resource.EffectiveResource{
				DefaultTag:    "en",
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")

			data, err = read(resource.EffectiveResource{
				DefaultTag:    "en",
				PreferredTags: []string{"en"},
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")

			data, err = read(resource.EffectiveResource{
				DefaultTag:    "en",
				PreferredTags: []string{"zh"},
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "zh in fs A")

		})

		Convey("it should return something if default language is not found", func() {
			writeFile(fsA, "en", "en in fs A")
			writeFile(fsA, "zh", "zh in fs A")

			data, err := read(resource.EffectiveResource{
				DefaultTag:    "en",
				PreferredTags: []string{"ja"},
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")
		})

		Convey("it should combine resources in different FS", func() {
			writeFile(fsB, "zh", "zh in fs B")
			writeFile(fsA, "en", "en in fs A")

			data, err := read(resource.EffectiveResource{
				DefaultTag:    "en",
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")

			data, err = read(resource.EffectiveResource{
				DefaultTag:    "en",
				PreferredTags: []string{"en"},
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")

			data, err = read(resource.EffectiveResource{
				DefaultTag:    "en",
				PreferredTags: []string{"zh"},
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "zh in fs B")
		})

		Convey("it should not fail when fallback is not en", func() {
			writeFile(fsA, "en", "en in fs A")

			data, err := read(resource.EffectiveResource{
				DefaultTag:    "zh",
				SupportedTags: []string{"zh"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")
		})
	})

	Convey("HTML EffectiveFile", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelCustom},
		})

		txt := &template.HTML{Name: "resource.txt"}
		r.Register(txt)

		writeFile := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/resource.txt", []byte(data), 0666)
		}

		read := func(lang string) (str string, err error) {
			view := resource.EffectiveFile{
				Path: "templates/" + lang + "/resource.txt",
			}
			result, err := manager.Read(txt, view)
			if err != nil {
				return
			}

			str = string(result.([]byte))
			return
		}

		Convey("it should return single resource", func() {
			writeFile(fsA, "en", "en in fs A")

			data, err := read("en")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")

			_, err = read("zh")
			So(err, ShouldBeError, "specified resource is not configured")
		})

		Convey("it should return resource with preferred language", func() {
			writeFile(fsA, "en", "en in fs A")
			writeFile(fsA, "zh", "zh in fs A")

			data, err := read("en")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")

			data, err = read("zh")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "zh in fs A")
		})

		Convey("it should combine resources in different FS", func() {
			writeFile(fsB, "zh", "zh in fs B")
			writeFile(fsA, "en", "en in fs A")

			data, err := read("en")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs A")

			data, err = read("zh")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "zh in fs B")
		})
	})

	Convey("HTML AppFile", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelApp},
		})

		txt := &template.MessageHTML{Name: "messages/resource.txt"}
		r.Register(txt)

		writeFile := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang+"/messages", 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/messages/resource.txt", []byte(data), 0666)
		}

		read := func(lang string) (str string, err error) {
			view := resource.AppFile{
				Path: "templates/" + lang + "/messages/resource.txt",
			}
			result, err := manager.Read(txt, view)
			if err != nil {
				return
			}

			str = string(result.([]byte))
			return
		}

		Convey("not found", func() {
			writeFile(fsA, "en", "en in fs A")

			_, err := read("en")
			So(err, ShouldBeError, "specified resource is not configured")
		})

		Convey("found", func() {
			writeFile(fsB, "en", "en in fs B")

			data, err := read("en")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs B")
		})

		Convey("it should return resource in app FS", func() {
			writeFile(fsA, "en", "en in fs A")
			writeFile(fsB, "en", "en in fs B")

			data, err := read("en")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, "en in fs B")
		})
	})

	Convey("matchTemplatePath", t, func() {
		// expected path "templates/{{ localeKey }}/.../{{ templateName }}"
		mockRes := &template.HTML{
			Name: "messages/dummy.html",
		}

		Convey("path with less than 3 segments", func() {
			_, result := mockRes.MatchResource(
				"templates/dummy.html",
			)
			So(result, ShouldBeFalse)
		})

		Convey("first segment mismatch", func() {
			_, result := mockRes.MatchResource(
				"wrong/en/messages/dummy.html",
			)
			So(result, ShouldBeFalse)
		})

		Convey("invalid locale key", func() {
			_, result := mockRes.MatchResource(
				"templates/abc/messages/dummy.html",
			)
			So(result, ShouldBeFalse)
		})

		Convey("template name mismatch", func() {
			_, result := mockRes.MatchResource(
				"templates/en/messages/wrong_name.html",
			)
			So(result, ShouldBeFalse)
		})

		Convey("valid path", func() {
			match, result := mockRes.MatchResource(
				"templates/en/messages/dummy.html",
			)
			So(result, ShouldBeTrue)
			So(match.LanguageTag, ShouldEqual, "en")
		})
	})
}
