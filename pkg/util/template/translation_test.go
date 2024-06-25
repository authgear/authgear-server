package template_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
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
		fsC := afero.NewMemMapFs()

		r := &resource.Registry{}
		r.Register(template.TranslationJSON)

		manager := resource.NewManager(r, []resource.Fs{
			resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelCustom},
			resource.LeveledAferoFs{Fs: fsC, FsLevel: resource.FsLevelApp},
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
			writeFile(fsC, "en", `{
				"c": "en c in fs C"
			}`)
			writeFile(fsC, "zh", `{
				"b": "zh b in fs C",
				"c": "zh c in fs C"
			}`)

			data, err := read(resource.EffectiveResource{
				DefaultTag:    "en",
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "en a in fs A" },
				"b": { "LanguageTag": "en", "Value": "en b in fs A" },
				"c": { "LanguageTag": "en", "Value": "en c in fs C" }
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
				"c": { "LanguageTag": "en", "Value": "en c in fs C" }
			}`))

			data, err = read(resource.EffectiveResource{
				DefaultTag:    "en",
				PreferredTags: []string{"zh"},
				SupportedTags: []string{"zh", "en"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "en", "Value": "en a in fs A" },
				"b": { "LanguageTag": "zh", "Value": "zh b in fs C" },
				"c": { "LanguageTag": "zh", "Value": "zh c in fs C" }
			}`))
		})

		Convey("it should not fail when fallback is not en", func() {
			writeFile(fsA, "en", `{
				"a": "en a in fs A",
				"b": "en b in fs A",
				"c": "en c in fs A"
			}`)
			writeFile(fsC, "en", `{
				"b": "en b in fs C"
			}`)
			writeFile(fsC, "zh", `{
				"c": "zh c in fs C"
			}`)

			data, err := read(resource.EffectiveResource{
				DefaultTag:    "zh",
				SupportedTags: []string{"zh"},
			})
			So(err, ShouldBeNil)
			So(data, ShouldEqual, compact(`{
				"a": { "LanguageTag": "zh", "Value": "en a in fs A" },
				"b": { "LanguageTag": "zh", "Value": "en b in fs A" },
				"c": { "LanguageTag": "zh", "Value": "zh c in fs C" }
			}`))
		})

		Convey("it should resolve based on app agnostic / app specific keys", func() {
			writeFile(fsA, "en", `{
				"app.name": "en app.name in fs A",
				"email.default.sender": "no-reply+en@authgear.com",
				"some-key-1": "en some-key-1 in fs A",
				"some-key-2": "en some-key-2 in fs A"
			}`)
			writeFile(fsA, "zh-HK", `{
				"app.name": "zh-HK app.name in fs A",
				"email.default.sender": "no-reply+zh@authgear.com",
				"some-key-1": "zh-HK some-key-1 in fs A",
				"some-key-2": "zh-HK some-key-2 in fs A"
			}`)
			Convey("should resolve all keys when no keys are provided in higher fs level", func() {
				er := resource.EffectiveResource{
					DefaultTag:    "en",
					SupportedTags: []string{"en", "zh-HK", "ja"},
				}
				data, err := read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "en", "Value": "en app.name in fs A" },
					"email.default.sender": { "LanguageTag": "en", "Value": "no-reply+en@authgear.com" },
					"some-key-1": { "LanguageTag": "en", "Value": "en some-key-1 in fs A" },
					"some-key-2": { "LanguageTag": "en", "Value": "en some-key-2 in fs A" }
				}`))

				er.PreferredTags = []string{"zh"}
				data, err = read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "zh-HK", "Value": "zh-HK app.name in fs A" },
					"email.default.sender": { "LanguageTag": "zh-HK", "Value": "no-reply+zh@authgear.com" },
					"some-key-1": { "LanguageTag": "zh-HK", "Value": "zh-HK some-key-1 in fs A" },
					"some-key-2": { "LanguageTag": "zh-HK", "Value": "zh-HK some-key-2 in fs A" }
				}`))

				er.PreferredTags = []string{"ja"}
				data, err = read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "en", "Value": "en app.name in fs A" },
					"email.default.sender": { "LanguageTag": "en", "Value": "no-reply+en@authgear.com" },
					"some-key-1": { "LanguageTag": "en", "Value": "en some-key-1 in fs A" },
					"some-key-2": { "LanguageTag": "en", "Value": "en some-key-2 in fs A" }
				}`))
			})

			Convey("should resolve when keys are provided in custom fs level fallback language", func() {
				writeFile(fsB, "en", `{
					"email.default.sender": "no-reply@example.com"
				}`)
				er := resource.EffectiveResource{
					DefaultTag:    "en",
					SupportedTags: []string{"en", "zh-HK", "ja"},
				}
				er.PreferredTags = []string{"en"}
				data, err := read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "en", "Value": "en app.name in fs A" },
					"email.default.sender": { "LanguageTag": "en", "Value": "no-reply@example.com" },
					"some-key-1": { "LanguageTag": "en", "Value": "en some-key-1 in fs A" },
					"some-key-2": { "LanguageTag": "en", "Value": "en some-key-2 in fs A" }
				}`))

				er.PreferredTags = []string{"zh"}
				data, err = read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "zh-HK", "Value": "zh-HK app.name in fs A" },
					"email.default.sender": { "LanguageTag": "en", "Value": "no-reply@example.com" },
					"some-key-1": { "LanguageTag": "zh-HK", "Value": "zh-HK some-key-1 in fs A" },
					"some-key-2": { "LanguageTag": "zh-HK", "Value": "zh-HK some-key-2 in fs A" }
				}`))

				er.PreferredTags = []string{"ja"}
				data, err = read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "en", "Value": "en app.name in fs A" },
					"email.default.sender": { "LanguageTag": "en", "Value": "no-reply@example.com" },
					"some-key-1": { "LanguageTag": "en", "Value": "en some-key-1 in fs A" },
					"some-key-2": { "LanguageTag": "en", "Value": "en some-key-2 in fs A" }
				}`))

				er.PreferredTags = []string{"ko"}
				data, err = read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "en", "Value": "en app.name in fs A" },
					"email.default.sender": { "LanguageTag": "en", "Value": "no-reply@example.com" },
					"some-key-1": { "LanguageTag": "en", "Value": "en some-key-1 in fs A" },
					"some-key-2": { "LanguageTag": "en", "Value": "en some-key-2 in fs A" }
				}`))
			})

			Convey("should resolve when keys are provided in app fs level with non-English fallback language", func() {
				writeFile(fsB, "en", `{
					"email.default.sender": "no-reply+en@custom.com"
				}`)
				writeFile(fsC, "ja", `{
					"email.default.sender": "no-reply+ja@app.com",
					"some-key-1": "ja some-key-1 in fs C"
				}`)
				er := resource.EffectiveResource{
					DefaultTag:    "ja",
					SupportedTags: []string{"en", "zh-HK", "ja"},
				}
				er.PreferredTags = []string{"en"}
				data, err := read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "en", "Value": "en app.name in fs A" },
					"email.default.sender": { "LanguageTag": "ja", "Value": "no-reply+ja@app.com" },
					"some-key-1": { "LanguageTag": "en", "Value": "en some-key-1 in fs A" },
					"some-key-2": { "LanguageTag": "en", "Value": "en some-key-2 in fs A" }
				}`))

				er.PreferredTags = []string{"zh"}
				data, err = read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "zh-HK", "Value": "zh-HK app.name in fs A" },
					"email.default.sender": { "LanguageTag": "ja", "Value": "no-reply+ja@app.com" },
					"some-key-1": { "LanguageTag": "zh-HK", "Value": "zh-HK some-key-1 in fs A" },
					"some-key-2": { "LanguageTag": "zh-HK", "Value": "zh-HK some-key-2 in fs A" }
				}`))

				er.PreferredTags = []string{"ja"}
				data, err = read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "ja", "Value": "en app.name in fs A" },
					"email.default.sender": { "LanguageTag": "ja", "Value": "no-reply+ja@app.com" },
					"some-key-1": { "LanguageTag": "ja", "Value": "ja some-key-1 in fs C" },
					"some-key-2": { "LanguageTag": "ja", "Value": "en some-key-2 in fs A" }
				}`))

				er.PreferredTags = []string{"ko"}
				data, err = read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "ja", "Value": "en app.name in fs A" },
					"email.default.sender": { "LanguageTag": "ja", "Value": "no-reply+ja@app.com" },
					"some-key-1": { "LanguageTag": "ja", "Value": "ja some-key-1 in fs C" },
					"some-key-2": { "LanguageTag": "ja", "Value": "en some-key-2 in fs A" }
				}`))
			})

			Convey("should resolve when keys are provided in app fs level with non fallback language", func() {
				writeFile(fsB, "en", `{
					"email.default.sender": "no-reply+en@custom.com"
				}`)
				writeFile(fsC, "ja", `{
					"email.default.sender": "no-reply+ja@app.com",
					"some-key-1": "ja some-key-1 in fs C"
				}`)
				er := resource.EffectiveResource{
					DefaultTag:    "zh-HK",
					SupportedTags: []string{"en", "zh-HK", "ja"},
				}
				er.PreferredTags = []string{"en"}
				data, err := read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "en", "Value": "en app.name in fs A" },
					"email.default.sender": { "LanguageTag": "ja", "Value": "no-reply+ja@app.com" },
					"some-key-1": { "LanguageTag": "en", "Value": "en some-key-1 in fs A" },
					"some-key-2": { "LanguageTag": "en", "Value": "en some-key-2 in fs A" }
				}`))

				er.PreferredTags = []string{"zh"}
				data, err = read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "zh-HK", "Value": "zh-HK app.name in fs A" },
					"email.default.sender": { "LanguageTag": "ja", "Value": "no-reply+ja@app.com" },
					"some-key-1": { "LanguageTag": "zh-HK", "Value": "zh-HK some-key-1 in fs A" },
					"some-key-2": { "LanguageTag": "zh-HK", "Value": "zh-HK some-key-2 in fs A" }
				}`))

				er.PreferredTags = []string{"ja"}
				data, err = read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "zh-HK", "Value": "zh-HK app.name in fs A" },
					"email.default.sender": { "LanguageTag": "ja", "Value": "no-reply+ja@app.com" },
					"some-key-1": { "LanguageTag": "ja", "Value": "ja some-key-1 in fs C" },
					"some-key-2": { "LanguageTag": "zh-HK", "Value": "zh-HK some-key-2 in fs A" }
				}`))

				er.PreferredTags = []string{"ko"}
				data, err = read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "zh-HK", "Value": "zh-HK app.name in fs A" },
					"email.default.sender": { "LanguageTag": "ja", "Value": "no-reply+ja@app.com" },
					"some-key-1": { "LanguageTag": "zh-HK", "Value": "zh-HK some-key-1 in fs A" },
					"some-key-2": { "LanguageTag": "zh-HK", "Value": "zh-HK some-key-2 in fs A" }
				}`))
			})

			Convey("should resolve when keys are provided in custom fs level with not supported language", func() {
				writeFile(fsB, "en", `{
					"email.default.sender": "no-reply+en@custom.com"
				}`)
				er := resource.EffectiveResource{
					DefaultTag:    "ja-JP",
					SupportedTags: []string{"ja-JP"},
				}
				data, err := read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "ja-JP", "Value": "en app.name in fs A" },
					"email.default.sender": { "LanguageTag": "ja-JP", "Value": "no-reply+en@custom.com" },
					"some-key-1": { "LanguageTag": "ja-JP", "Value": "en some-key-1 in fs A" },
					"some-key-2": { "LanguageTag": "ja-JP", "Value": "en some-key-2 in fs A" }
				}`))
			})

			Convey("app fs has precedence over custom fs when the language is supported", func() {
				writeFile(fsB, "en", `{
					"email.default.sender": "no-reply+en@custom.com"
				}`)
				writeFile(fsC, "ja-JP", `{
					"email.default.sender": "no-reply+ja@app.com"
				}`)
				er := resource.EffectiveResource{
					DefaultTag:    "ja-JP",
					SupportedTags: []string{"ja-JP"},
				}
				data, err := read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "ja-JP", "Value": "en app.name in fs A" },
					"email.default.sender": { "LanguageTag": "ja-JP", "Value": "no-reply+ja@app.com" },
					"some-key-1": { "LanguageTag": "ja-JP", "Value": "en some-key-1 in fs A" },
					"some-key-2": { "LanguageTag": "ja-JP", "Value": "en some-key-2 in fs A" }
				}`))
			})

			Convey("app fs has NO precedence over custom fs when the language is unsupported", func() {
				writeFile(fsB, "en", `{
					"email.default.sender": "no-reply+en@custom.com"
				}`)
				writeFile(fsC, "en", `{
					"email.default.sender": "no-reply+en@app.com"
				}`)
				er := resource.EffectiveResource{
					DefaultTag:    "ja-JP",
					SupportedTags: []string{"ja-JP"},
				}
				data, err := read(er)
				So(err, ShouldBeNil)
				So(data, ShouldEqual, compact(`{
					"app.name": { "LanguageTag": "ja-JP", "Value": "en app.name in fs A" },
					"email.default.sender": { "LanguageTag": "ja-JP", "Value": "no-reply+en@custom.com" },
					"some-key-1": { "LanguageTag": "ja-JP", "Value": "en some-key-1 in fs A" },
					"some-key-2": { "LanguageTag": "ja-JP", "Value": "en some-key-2 in fs A" }
				}`))
			})
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

	Convey("TranslationJSON UpdateResource", t, func() {
		path := "templates/en/translation.json"
		builtin := resource.LeveledAferoFs{FsLevel: resource.FsLevelBuiltin}
		app := resource.LeveledAferoFs{FsLevel: resource.FsLevelApp}

		Convey("it should only write value that is not equal to default value", func() {
			featureConfig := config.NewEffectiveDefaultFeatureConfig()
			ctx := context.Background()
			ctx = context.WithValue(ctx, configsource.ContextKeyFeatureConfig, featureConfig)
			updated, err := template.TranslationJSON.UpdateResource(
				ctx,
				[]resource.ResourceFile{
					resource.ResourceFile{
						Location: resource.Location{
							Fs:   builtin,
							Path: path,
						},
						Data: []byte(`{
							"a": "default a",
							"b": "default b"
						}`),
					},
				},
				&resource.ResourceFile{
					Location: resource.Location{
						Fs:   app,
						Path: path,
					},
					Data: nil,
				},
				[]byte(`{
					"a": "default a",
					"b": "new b",
					"unknown": "key"
				}`),
			)

			So(err, ShouldBeNil)
			So(updated, ShouldResemble, &resource.ResourceFile{
				Location: resource.Location{
					Fs:   app,
					Path: path,
				},
				Data: []byte(`{"b":"new b","unknown":"key"}`),
			})
		})

		Convey("it should delete the file if the file is empty", func() {
			featureConfig := config.NewEffectiveDefaultFeatureConfig()
			ctx := context.Background()
			ctx = context.WithValue(ctx, configsource.ContextKeyFeatureConfig, featureConfig)
			updated, err := template.TranslationJSON.UpdateResource(
				ctx,
				[]resource.ResourceFile{
					resource.ResourceFile{
						Location: resource.Location{
							Fs:   builtin,
							Path: path,
						},
						Data: []byte(`{
							"a": "default a",
							"b": "default b"
						}`),
					},
				},
				&resource.ResourceFile{
					Location: resource.Location{
						Fs:   app,
						Path: path,
					},
					Data: nil,
				},
				[]byte(`{
					"a": "default a"
				}`),
			)

			So(err, ShouldBeNil)
			So(updated, ShouldResemble, &resource.ResourceFile{
				Location: resource.Location{
					Fs:   app,
					Path: path,
				},
				Data: nil,
			})
		})

		Convey("it should skip write value for email template if disallowed by feature flag", func() {
			featureConfig := config.NewEffectiveDefaultFeatureConfig()
			featureConfig.Messaging.TemplateCustomizationDisabled = true
			ctx := context.Background()
			ctx = context.WithValue(ctx, configsource.ContextKeyFeatureConfig, featureConfig)
			updated, err := template.TranslationJSON.UpdateResource(
				ctx,
				[]resource.ResourceFile{
					resource.ResourceFile{
						Location: resource.Location{
							Fs:   builtin,
							Path: path,
						},
						Data: []byte(`{
							"email.test.subject": "default title"
						}`),
					},
				},
				&resource.ResourceFile{
					Location: resource.Location{
						Fs:   app,
						Path: path,
					},
					Data: nil,
				},
				[]byte(`{
					"email.test.subject": "skip me",
					"another.test.subject": "foo",
					"email.test.another": "bar"
				}`),
			)

			So(err, ShouldBeNil)
			So(updated, ShouldResemble, &resource.ResourceFile{
				Location: resource.Location{
					Fs:   app,
					Path: path,
				},
				Data: []byte(`{"another.test.subject":"foo","email.test.another":"bar"}`),
			})
		})
	})

	Convey("TranslationJSON IsAppSpecificKey", t, func() {
		test := func(key string, result bool) {
			actual := template.TranslationJSON.(interface{ IsAppSpecificKey(key string) bool }).IsAppSpecificKey(key)
			So(actual, ShouldEqual, result)
		}

		test("app.name", true)
		test("email.default.sender", true)
		test("email.verification.sender", true)
		test("email.verification.reply-to", true)
		test("sms.default.sender", true)

		test("email.default.subject", false)
		test("email.verification.subject", false)
		test("settings-my-profile-title", false)
		test("any-key", false)

	})
}
