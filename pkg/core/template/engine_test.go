package template

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

func TestEngine(t *testing.T) {
	const typeA config.TemplateItemType = "typeA"
	const typeB config.TemplateItemType = "typeB"
	const keyA = "keyA"
	const keyB = "keyB"
	specA := Spec{Type: typeA, IsKeyed: true}

	Convey("Engine", t, func() {
		Convey("resolveTemplateItem", func() {
			cases := []struct {
				TemplateItems         []config.TemplateItem
				PreferredLanguageTags []string

				Spec Spec
				Key  string

				Expected *config.TemplateItem
			}{
				{
					// No TemplateItems
					Spec:     specA,
					Expected: nil,
				},
				{
					// Select type
					TemplateItems: []config.TemplateItem{
						config.TemplateItem{
							Type: typeA,
						},
						config.TemplateItem{
							Type: typeB,
						},
					},
					Spec: specA,
					Expected: &config.TemplateItem{
						Type: typeA,
					},
				},
				{
					// Select key
					TemplateItems: []config.TemplateItem{
						config.TemplateItem{
							Type: typeA,
						},
						config.TemplateItem{
							Type: typeB,
						},
						config.TemplateItem{
							Type: typeA,
							Key:  keyA,
						},
						config.TemplateItem{
							Type: typeA,
							Key:  keyB,
						},
					},
					Spec: specA,
					Key:  keyA,
					Expected: &config.TemplateItem{
						Type: typeA,
						Key:  keyA,
					},
				},
				{
					// Select the empty language tag
					// If no preferred languages are given.
					TemplateItems: []config.TemplateItem{
						config.TemplateItem{
							Type: typeA,
							URI:  "default",
						},
						config.TemplateItem{
							Type:        typeA,
							LanguageTag: "en-US",
							URI:         "American English",
						},
						config.TemplateItem{
							Type:        typeA,
							LanguageTag: "zh-Hant-HK",
							URI:         "Traditional Chinese in Hong Kong",
						},
					},
					Spec: specA,
					Expected: &config.TemplateItem{
						Type: typeA,
						URI:  "default",
					},
				},
				{
					// Select the empty language tag
					// If no preferred languages can be matched.
					TemplateItems: []config.TemplateItem{
						config.TemplateItem{
							Type: typeA,
							URI:  "default",
						},
						config.TemplateItem{
							Type:        typeA,
							LanguageTag: "en-US",
							URI:         "American English",
						},
						config.TemplateItem{
							Type:        typeA,
							LanguageTag: "zh-Hant-HK",
							URI:         "Traditional Chinese in Hong Kong",
						},
					},
					PreferredLanguageTags: []string{"ja-JP"},
					Spec:                  specA,
					Expected: &config.TemplateItem{
						Type: typeA,
						URI:  "default",
					},
				},
				{
					// Select the best language tag.
					TemplateItems: []config.TemplateItem{
						config.TemplateItem{
							Type: typeA,
							URI:  "default",
						},
						config.TemplateItem{
							Type:        typeA,
							LanguageTag: "en-US",
							URI:         "American English",
						},
						config.TemplateItem{
							Type:        typeA,
							LanguageTag: "zh-Hant-HK",
							URI:         "Traditional Chinese in Hong Kong",
						},
					},
					PreferredLanguageTags: []string{"en"},
					Spec:                  specA,
					Expected: &config.TemplateItem{
						Type:        typeA,
						LanguageTag: "en-US",
						URI:         "American English",
					},
				},
				{
					// Select the best language tag.
					TemplateItems: []config.TemplateItem{
						config.TemplateItem{
							Type: typeA,
							URI:  "default",
						},
						config.TemplateItem{
							Type:        typeA,
							LanguageTag: "en-US",
							URI:         "American English",
						},
						config.TemplateItem{
							Type:        typeA,
							LanguageTag: "zh-Hant-HK",
							URI:         "Traditional Chinese in Hong Kong",
						},
					},
					PreferredLanguageTags: []string{"ja-JP", "en-GB"},
					Spec:                  specA,
					Expected: &config.TemplateItem{
						Type:        typeA,
						LanguageTag: "en-US",
						URI:         "American English",
					},
				},
			}
			for _, c := range cases {
				e := NewEngine(NewEngineOptions{
					TemplateItems: c.TemplateItems,
				})
				e = e.WithPreferredLanguageTags(c.PreferredLanguageTags)

				actual, err := e.resolveTemplateItem(c.Spec, c.Key)
				if c.Expected == nil {
					So(err, ShouldBeError)
				} else {
					So(actual, ShouldResemble, c.Expected)
				}
			}
		})
	})
}

type mockLoader struct{}

func (l *mockLoader) Load(s string) (string, error) {
	return s, nil
}

func TestResolveTranslations(t *testing.T) {
	const typeA config.TemplateItemType = "typeA"
	specA := Spec{
		Type: typeA,
		Default: `
		{
			"key1": "Hello",
			"key2": "World"
		}
		`,
	}

	Convey("resolveTranslations", t, func() {
		test := func(items []config.TemplateItem, expected map[string]map[string]string) {
			e := NewEngine(NewEngineOptions{
				TemplateItems: items,
			})
			e.Register(specA)
			e.loader = &mockLoader{}

			actual, err := e.resolveTranslations(typeA)
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, expected)
		}

		// No provided translations
		test([]config.TemplateItem{}, map[string]map[string]string{
			"key1": map[string]string{
				"": "Hello",
			},
			"key2": map[string]string{
				"": "World",
			},
		})

		test([]config.TemplateItem{
			config.TemplateItem{
				Type:        typeA,
				LanguageTag: "zh",
				URI: `
				{
					"key1": "你好",
					"key2": "世界"
				}
				`,
			},
			config.TemplateItem{
				Type:        typeA,
				LanguageTag: "ja",
				URI: `
				{
					"key1": "こんにちは",
					"key2": "世界"
				}
				`,
			},
			config.TemplateItem{
				Type:        typeA,
				LanguageTag: "en",
				URI: `
				{
					"key1": "Hey"
				}
				`,
			},
		}, map[string]map[string]string{
			"key1": map[string]string{
				"":   "Hello",
				"en": "Hey",
				"zh": "你好",
				"ja": "こんにちは",
			},
			"key2": map[string]string{
				"":   "World",
				"zh": "世界",
				"ja": "世界",
			},
		})
	})
}
