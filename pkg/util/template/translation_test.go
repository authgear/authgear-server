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
	Convey("TranslationJSON EffectiveResource", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()

		r := &resource.Registry{}
		r.Register(template.TranslationJSON)

		manager := resource.NewManager(r, []resource.Fs{
			resource.AferoFs{Fs: fsA},
			resource.AferoFs{Fs: fsB},
		})

		view := resource.EffectiveResource{
			DefaultTag: "en",
		}

		writeFile := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/translation.json", []byte(data), 0666)
		}

		read := func() (str string, err error) {
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

		compact := func(s string) string {
			buf := &bytes.Buffer{}
			_ = json.Compact(buf, []byte(s))
			return buf.String()
		}

		Convey("it should return single resource", func() {
			writeFile(fsA, "__default__", `{
				"a": "default a in fs A",
				"b": "default b in fs A",
				"c": "default c in fs A"
			}`)

			data, err := read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "default a in fs A" },
				"b": { "LanguageTag": "en", "Value": "default b in fs A" },
				"c": { "LanguageTag": "en", "Value": "default c in fs A" }
			}`))
		})

		Convey("it should return resource with preferred language", func() {
			writeFile(fsA, "__default__", `{
				"a": "default a in fs A",
				"b": "default b in fs A",
				"c": "default c in fs A"
			}`)
			writeFile(fsA, "en", `{
				"b": "en b in fs A"
			}`)
			writeFile(fsA, "zh", `{
				"b": "zh b in fs A",
				"c": "zh c in fs A"
			}`)

			data, err := read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "default a in fs A" },
				"b": { "LanguageTag": "en", "Value": "en b in fs A" },
				"c": { "LanguageTag": "en", "Value": "default c in fs A" }
			}`))

			view.PreferredTags = []string{"en"}
			data, err = read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "default a in fs A" },
				"b": { "LanguageTag": "en", "Value": "en b in fs A" },
				"c": { "LanguageTag": "en", "Value": "default c in fs A" }
			}`))

			view.PreferredTags = []string{"zh"}
			data, err = read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "default a in fs A" },
				"b": { "LanguageTag": "zh", "Value": "zh b in fs A" },
				"c": { "LanguageTag": "zh", "Value": "zh c in fs A" }
			}`))
		})

		Convey("it should combine resources in different FS", func() {
			writeFile(fsA, "__default__", `{
				"a": "default a in fs A",
				"b": "default b in fs A",
				"c": "default c in fs A"
			}`)
			writeFile(fsA, "en", `{
				"b": "en b in fs A"
			}`)
			writeFile(fsB, "en", `{
				"c": "en c in fs B"
			}`)
			writeFile(fsB, "zh", `{
				"b": "zh b in fs B",
				"c": "zh c in fs B"
			}`)

			data, err := read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "default a in fs A" },
				"b": { "LanguageTag": "en", "Value": "en b in fs A" },
				"c": { "LanguageTag": "en", "Value": "en c in fs B" }
			}`))

			view.PreferredTags = []string{"en"}
			data, err = read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "default a in fs A" },
				"b": { "LanguageTag": "en", "Value": "en b in fs A" },
				"c": { "LanguageTag": "en", "Value": "en c in fs B" }
			}`))

			view.PreferredTags = []string{"zh"}
			data, err = read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "default a in fs A" },
				"b": { "LanguageTag": "zh", "Value": "zh b in fs B" },
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
			resource.AferoFs{Fs: fsA},
			resource.AferoFs{Fs: fsB},
		})

		writeFile := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/translation.json", []byte(data), 0666)
		}

		read := func(lang string) (str string, err error) {
			view := resource.EffectiveFile{
				Path:       "templates/" + lang + "/translation.json",
				DefaultTag: "en",
			}
			result, err := manager.Read(template.TranslationJSON, view)
			if err != nil {
				return
			}

			translations := result.(map[string]string)

			bytes, err := json.Marshal(translations)
			if err != nil {
				return
			}

			return string(bytes), nil
		}

		compact := func(s string) string {
			buf := &bytes.Buffer{}
			_ = json.Compact(buf, []byte(s))
			return buf.String()
		}

		Convey("it should return single resource", func() {
			writeFile(fsA, "__default__", `{
				"a": "default a in fs A",
				"b": "default b in fs A",
				"c": "default c in fs A"
			}`)

			data, err := read("__default__")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": "default a in fs A",
				"b": "default b in fs A",
				"c": "default c in fs A"
			}`))
		})

		Convey("it should return resource with specific language", func() {
			writeFile(fsA, "__default__", `{
				"a": "default a in fs A",
				"b": "default b in fs A",
				"c": "default c in fs A"
			}`)
			writeFile(fsA, "en", `{
				"b": "en b in fs A"
			}`)
			writeFile(fsA, "zh", `{
				"b": "zh b in fs A",
				"c": "zh c in fs A"
			}`)

			data, err := read("en")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": "default a in fs A",
				"b": "en b in fs A",
				"c": "default c in fs A"
			}`))

			data, err = read("zh")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"b": "zh b in fs A",
				"c": "zh c in fs A"
			}`))
		})

		Convey("it should combine resources in different FS", func() {
			writeFile(fsA, "__default__", `{
				"a": "default a in fs A",
				"b": "default b in fs A",
				"c": "default c in fs A"
			}`)
			writeFile(fsA, "en", `{
				"b": "en b in fs A"
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
			So(data, ShouldEqual, compact(`{
				"a": "default a in fs A",
				"b": "en b in fs A",
				"c": "en c in fs B"
			}`))

			data, err = read("zh")
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"b": "zh b in fs B",
				"c": "zh c in fs B"
			}`))
		})
	})
}
