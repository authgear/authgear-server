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
	Convey("TranslationJSON", t, func() {
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		r := &resource.Registry{}
		manager := resource.NewManager(r, []resource.Fs{
			resource.AferoFs{Fs: fsA},
			resource.AferoFs{Fs: fsB},
		})

		r.Register(template.TranslationJSON)

		args := map[string]interface{}{
			template.ResourceArgDefaultLanguageTag: "en",
		}

		writeFile := func(fs afero.Fs, lang string, data string) {
			_ = fs.MkdirAll("templates/"+lang, 0777)
			_ = afero.WriteFile(fs, "templates/"+lang+"/translation.json", []byte(data), 0666)
		}

		read := func() (string, error) {
			merged, err := manager.Read(template.TranslationJSON, args)
			if err != nil {
				return "", err
			}
			return string(merged.Data), nil
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

			args[template.ResourceArgPreferredLanguageTag] = []string{"en"}
			data, err = read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "default a in fs A" },
				"b": { "LanguageTag": "en", "Value": "en b in fs A" },
				"c": { "LanguageTag": "en", "Value": "default c in fs A" }
			}`))

			args[template.ResourceArgPreferredLanguageTag] = []string{"zh"}
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

			args[template.ResourceArgPreferredLanguageTag] = []string{"en"}
			data, err = read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "default a in fs A" },
				"b": { "LanguageTag": "en", "Value": "en b in fs A" },
				"c": { "LanguageTag": "en", "Value": "en c in fs B" }
			}`))

			args[template.ResourceArgPreferredLanguageTag] = []string{"zh"}
			data, err = read()
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "default a in fs A" },
				"b": { "LanguageTag": "zh", "Value": "zh b in fs B" },
				"c": { "LanguageTag": "zh", "Value": "zh c in fs B" }
			}`))
		})
	})
}
