package template_test

import (
	"bytes"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func TestTranslationResource(t *testing.T) {
	Convey("TranslationJSON ValidateResource", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()

		r := &resource.Registry{}
		r.Register(template.TranslationJSON)

		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelApp},
		})

		compact := func(s string) string {
			buf := &bytes.Buffer{}
			_ = json.Compact(buf, []byte(s))
			return buf.String()
		}

		writeFile := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/translation.json", []byte(compact(data)), 0666)
		}

		read := func(view resource.View) (str string, err error) {
			result, err := manager.Read(template.TranslationJSON, view)
			if err != nil {
				return
			}

			translations := result.(map[string]template.Translation)

			bytes, err := json.Marshal(translations)
			if err != nil {
				return
			}

			return string(bytes), nil
		}

		Convey("it should validate", func() {
			writeFile(fsA, "en", `{
				"a": "{invalid",
				"b": "en b in fs A",
				"c": "en c in fs A"
			}`)

			_, err := read(resource.ValidateResource{})
			So(err, ShouldBeError, "translation `a` is invalid: unexpected token: <EOF>")
		})
	})

	Convey("TranslationJSON EffectiveResource", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()

		r := &resource.Registry{}
		r.Register(template.TranslationJSON)

		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelApp},
		})

		compact := func(s string) string {
			buf := &bytes.Buffer{}
			_ = json.Compact(buf, []byte(s))
			return buf.String()
		}

		writeFile := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/translation.json", []byte(compact(data)), 0666)
		}

		read := func(view resource.View) (str string, err error) {
			result, err := manager.Read(template.TranslationJSON, view)
			if err != nil {
				return
			}

			translations := result.(map[string]template.Translation)

			bytes, err := json.Marshal(translations)
			if err != nil {
				return
			}

			return string(bytes), nil
		}

		Convey("it should return single resource", func() {
			writeFile(fsA, "en", `{
				"a": "en a in fs A",
				"b": "en b in fs A",
				"c": "en c in fs A"
			}`)

			data, err := read(resource.EffectiveResource{
				DefaultTag:    "en",
				SupportedTags: []string{"en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "en a in fs A" },
				"b": { "LanguageTag": "en", "Value": "en b in fs A" },
				"c": { "LanguageTag": "en", "Value": "en c in fs A" }
			}`))
		})

		Convey("it should return resource with preferred language", func() {
			writeFile(fsA, "en", `{
				"a": "en a in fs A",
				"b": "en b in fs A",
				"c": "en c in fs A"
			}`)
			writeFile(fsA, "zh", `{
				"b": "zh b in fs A",
				"c": "zh c in fs A"
			}`)

			data, err := read(resource.EffectiveResource{
				DefaultTag:    "en",
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "en a in fs A" },
				"b": { "LanguageTag": "en", "Value": "en b in fs A" },
				"c": { "LanguageTag": "en", "Value": "en c in fs A" }
			}`))

			data, err = read(resource.EffectiveResource{
				DefaultTag:    "en",
				PreferredTags: []string{"en"},
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "en a in fs A" },
				"b": { "LanguageTag": "en", "Value": "en b in fs A" },
				"c": { "LanguageTag": "en", "Value": "en c in fs A" }
			}`))

			data, err = read(resource.EffectiveResource{
				DefaultTag:    "en",
				PreferredTags: []string{"zh"},
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "en a in fs A" },
				"b": { "LanguageTag": "zh", "Value": "zh b in fs A" },
				"c": { "LanguageTag": "zh", "Value": "zh c in fs A" }
			}`))
		})

		Convey("it should combine resources in different FS", func() {
			writeFile(fsA, "en", `{
				"a": "en a in fs A",
				"b": "en b in fs A",
				"c": "en c in fs A"
			}`)
			writeFile(fsB, "en", `{
				"c": "en c in fs B"
			}`)
			writeFile(fsB, "zh", `{
				"b": "zh b in fs B",
				"c": "zh c in fs B"
			}`)

			data, err := read(resource.EffectiveResource{
				DefaultTag:    "en",
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "en a in fs A" },
				"b": { "LanguageTag": "en", "Value": "en b in fs A" },
				"c": { "LanguageTag": "en", "Value": "en c in fs B" }
			}`))

			data, err = read(resource.EffectiveResource{
				DefaultTag:    "en",
				PreferredTags: []string{"en"},
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "en a in fs A" },
				"b": { "LanguageTag": "en", "Value": "en b in fs A" },
				"c": { "LanguageTag": "en", "Value": "en c in fs B" }
			}`))

			data, err = read(resource.EffectiveResource{
				DefaultTag:    "en",
				PreferredTags: []string{"zh"},
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "en a in fs A" },
				"b": { "LanguageTag": "zh", "Value": "zh b in fs B" },
				"c": { "LanguageTag": "zh", "Value": "zh c in fs B" }
			}`))
		})

		Convey("it should not fail when fallback is not en", func() {
			writeFile(fsA, "en", `{
				"a": "en a in fs A",
				"b": "en b in fs A",
				"c": "en c in fs A"
			}`)
			writeFile(fsB, "en", `{
				"b": "en b in fs B",
			}`)
			writeFile(fsB, "zh", `{
				"c": "zh c in fs B"
			}`)

			data, err := read(resource.EffectiveResource{
				DefaultTag:    "zh",
				SupportedTags: []string{"zh"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "zh", "Value": "en a in fs A" },
				"b": { "LanguageTag": "zh", "Value": "en b in fs A" },
				"c": { "LanguageTag": "zh", "Value": "zh c in fs B" }
			}`))
		})
	})

	Convey("TranslationJSON EffectiveFile", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()

		r := &resource.Registry{}
		r.Register(template.TranslationJSON)

		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelApp},
		})

		compact := func(s string) string {
			buf := &bytes.Buffer{}
			_ = json.Compact(buf, []byte(s))
			return buf.String()
		}

		writeFile := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/translation.json", []byte(compact(data)), 0666)
		}

		read := func(lang string) (str string, err error) {
			view := resource.EffectiveFile{
				Path: "templates/" + lang + "/translation.json",
			}
			result, err := manager.Read(template.TranslationJSON, view)
			if err != nil {
				return
			}

			bytes := result.([]byte)
			return string(bytes), nil
		}

		Convey("it should return single resource", func() {
			writeFile(fsA, "en", `{
				"a": "en a in fs A",
				"b": "en b in fs A",
				"c": "en c in fs A"
			}`)

			data, err := read("en")
			So(err, ShouldBeNil)
			So(compact(data), ShouldEqual, compact(`{
				"a": "en a in fs A",
				"b": "en b in fs A",
				"c": "en c in fs A"
			}`))
		})

		Convey("it should return resource with specific language", func() {
			writeFile(fsA, "en", `{
				"a": "en a in fs A",
				"b": "en b in fs A",
				"c": "en c in fs A"
			}`)
			writeFile(fsA, "zh", `{
				"b": "zh b in fs A",
				"c": "zh c in fs A"
			}`)

			data, err := read("en")
			So(err, ShouldBeNil)
			So(compact(data), ShouldEqual, compact(`{
				"a": "en a in fs A",
				"b": "en b in fs A",
				"c": "en c in fs A"
			}`))

			data, err = read("zh")
			So(err, ShouldBeNil)
			So(compact(data), ShouldEqual, compact(`{
				"b": "zh b in fs A",
				"c": "zh c in fs A"
			}`))
		})

		Convey("it should combine resources in different FS", func() {
			writeFile(fsA, "en", `{
				"a": "en a in fs A",
				"b": "en b in fs A",
				"c": "en c in fs A"
			}`)
			writeFile(fsB, "en", `{
				"c": "en c in fs B"
			}`)
			writeFile(fsB, "zh", `{
				"b": "zh b in fs B",
				"c": "zh c in fs B"
			}`)

			data, err := read("en")
			So(err, ShouldBeNil)
			So(compact(data), ShouldEqual, compact(`{
				"a": "en a in fs A",
				"b": "en b in fs A",
				"c": "en c in fs B"
			}`))

			data, err = read("zh")
			So(err, ShouldBeNil)
			So(compact(data), ShouldEqual, compact(`{
				"b": "zh b in fs B",
				"c": "zh c in fs B"
			}`))
		})
	})

	Convey("TranslationJSON AppFile", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()

		r := &resource.Registry{}
		r.Register(template.TranslationJSON)

		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelApp},
		})

		compact := func(s string) string {
			buf := &bytes.Buffer{}
			_ = json.Compact(buf, []byte(s))
			return buf.String()
		}

		writeFile := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/translation.json", []byte(compact(data)), 0666)
		}

		read := func(lang string) (str string, err error) {
			view := resource.AppFile{
				Path: "templates/" + lang + "/translation.json",
			}
			result, err := manager.Read(template.TranslationJSON, view)
			if err != nil {
				return
			}

			bytes := result.([]byte)
			return string(bytes), nil
		}

		Convey("not found", func() {
			_, err := read("en")
			So(err, ShouldBeError, "specified resource is not configured")
		})

		Convey("found", func() {
			writeFile(fsB, "en", `{
				"a": "en a in fs A",
				"b": "en b in fs A",
				"c": "en c in fs A"
			}`)

			data, err := read("en")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": "en a in fs A",
				"b": "en b in fs A",
				"c": "en c in fs A"
			}`))
		})

		Convey("it should return resource in app FS", func() {
			writeFile(fsA, "en", `{
				"a": "en a in fs A",
				"b": "en b in fs A",
				"c": "en c in fs A"
			}`)
			writeFile(fsB, "en", `{
				"a": "en a in fs B",
				"b": "en b in fs B",
				"c": "en c in fs B"
			}`)

			data, err := read("en")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": "en a in fs B",
				"b": "en b in fs B",
				"c": "en c in fs B"
			}`))
		})
	})
}
