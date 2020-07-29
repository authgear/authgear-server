package template

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/config"
	. "github.com/authgear/authgear-server/pkg/core/skytest"
)

func TestResolveTemplateItem(t *testing.T) {
	const typeA config.TemplateItemType = "typeA"
	const typeB config.TemplateItemType = "typeB"
	specA := Spec{Type: typeA}
	Convey("resolveTemplateItem", t, func() {
		test := func(templateItems []config.TemplateItem, tags []string, expected *config.TemplateItem) {
			e := NewEngine(NewEngineOptions{
				TemplateItems: templateItems,
			})
			e = e.WithPreferredLanguageTags(tags)
			actual, err := e.resolveTemplateItem(specA)
			if expected == nil {
				So(err, ShouldBeError)
			} else {
				So(actual, ShouldResemble, expected)
			}
		}

		// No TemplateItems
		test(nil, nil, nil)

		// Select type
		test([]config.TemplateItem{
			config.TemplateItem{
				Type: typeA,
			},
			config.TemplateItem{
				Type: typeB,
			},
		}, nil, &config.TemplateItem{
			Type:        typeA,
			LanguageTag: "en",
		})

		// Select the empty language tag
		// If no preferred languages are given.
		test([]config.TemplateItem{
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
		}, nil, &config.TemplateItem{
			Type:        typeA,
			URI:         "default",
			LanguageTag: "en",
		})

		// Select the best language tag.
		test([]config.TemplateItem{
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
		}, []string{"en-US"}, &config.TemplateItem{
			Type:        typeA,
			LanguageTag: "en-US",
			URI:         "American English",
		})

		// Invalid fallback.
		test([]config.TemplateItem{
			config.TemplateItem{
				Type:        typeA,
				LanguageTag: "en-US",
				URI:         "American English",
			},
		}, []string{"zh-Hant-HK"}, nil)

		// Select fallback.
		test([]config.TemplateItem{
			config.TemplateItem{
				Type: typeA,
				URI:  "fallback",
			},
		}, []string{"zh-Hant-HK"}, &config.TemplateItem{
			Type:        typeA,
			URI:         "fallback",
			LanguageTag: "en",
		})
	})
}

type mockLoader struct{}

func (l *mockLoader) Load(s string) (string, error) {
	return s, nil
}

type mockDefaultLoader struct {
	Defaults map[config.TemplateItemType]string
}

func (l *mockDefaultLoader) LoadDefault(t config.TemplateItemType) (string, error) {
	return l.Defaults[t], nil
}

func TestResolveTranslations(t *testing.T) {
	const typeA config.TemplateItemType = "typeA"
	specA := Spec{Type: typeA}

	defaultLoader := &mockDefaultLoader{}
	defaultLoader.Defaults = map[config.TemplateItemType]string{
		typeA: `
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
			e.defaultLoader = defaultLoader

			actual, err := e.resolveTranslations(typeA)
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, expected)
		}

		// No provided translations
		test([]config.TemplateItem{}, map[string]map[string]string{
			"key1": map[string]string{
				"en": "Hello",
			},
			"key2": map[string]string{
				"en": "World",
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
				"en": "Hey",
				"zh": "你好",
				"ja": "こんにちは",
			},
			"key2": map[string]string{
				"en": "World",
				"zh": "世界",
				"ja": "世界",
			},
		})
	})
}

func TestResolveComponents(t *testing.T) {
	componentA := Spec{
		Type: "componentA",
	}
	componentB := Spec{
		Type: "componentB",
	}
	componentC := Spec{
		Type:       "componentC",
		Components: []config.TemplateItemType{"componentA", "componentB"},
	}
	pageA := Spec{
		Type: "pageA",
	}
	pageB := Spec{
		Type:       "pageB",
		Components: []config.TemplateItemType{"componentA"},
	}
	pageC := Spec{
		Type:       "pageC",
		Components: []config.TemplateItemType{"componentC"},
	}
	pageD := Spec{
		Type:       "pageD",
		Components: []config.TemplateItemType{"componentA", "componentC"},
	}

	defaultLoader := &mockDefaultLoader{
		Defaults: map[config.TemplateItemType]string{},
	}

	specs := []Spec{
		componentA,
		componentB,
		componentC,
		pageA,
		pageB,
		pageC,
		pageD,
	}
	for _, spec := range specs {
		defaultLoader.Defaults[spec.Type] = string(spec.Type)
	}

	test := func(spec Spec, expected []string) {
		e := NewEngine(NewEngineOptions{})
		for _, s := range specs {
			e.Register(s)
		}
		e.loader = &mockLoader{}
		e.defaultLoader = defaultLoader

		actual, err := e.resolveComponents(spec.Components)
		So(err, ShouldBeNil)
		So(actual, ShouldEqualStringSliceWithoutOrder, expected)
	}

	Convey("resolveComponents", t, func() {
		// No components
		test(pageA, nil)
		// Only one component
		test(pageB, []string{"componentA"})
		// Transitive components
		test(pageC, []string{"componentC", "componentA", "componentB"})
		// Duplicate transitive components
		test(pageD, []string{"componentA", "componentC", "componentB"})
	})
}

func TestMakeLocalize(t *testing.T) {
	key := "key1"
	Convey("makeLocalize", t, func() {
		test := func(m map[string]string, preferredLanguageTags []string, expected string) {
			localize := makeLocalize(preferredLanguageTags, "en", map[string]map[string]string{
				key: m,
			})
			actual, err := localize(key)
			So(err, ShouldBeNil)
			So(actual, ShouldEqual, expected)
		}

		// Select default if there is no preferred languages
		test(map[string]string{
			"en": "Hello from en",
			"ja": "Hello from ja",
			"zh": "Hello from zh",
		}, nil, "Hello from en")
		test(map[string]string{
			"en": "Hello from en",
			"ja": "Hello from ja",
			"zh": "Hello from zh",
		}, []string{}, "Hello from en")

		// Simply select japanese
		test(map[string]string{
			"en": "Hello from en",
			"ja": "Hello from ja",
			"zh": "Hello from zh",
		}, []string{"ja-JP", "en-US", "zh-Hant-HK"}, "Hello from ja")

		// Select the default because korean is not supported
		test(map[string]string{
			"en": "Hello from en",
			"ja": "Hello from ja",
			"zh": "Hello from zh",
		}, []string{"kr-KR"}, "Hello from en")

		// Return empty string if fallback is invalid
		test(map[string]string{
			"ja": "Hello from ja",
			"zh": "Hello from zh",
		}, nil, "")
	})
}

func TestLocalize(t *testing.T) {
	translations := map[string]map[string]string{
		"key": map[string]string{
			"en": "Hello {0}",
		},
	}
	Convey("localize", t, func() {
		test := func(key string, expected string, args ...interface{}) {
			localize := makeLocalize(nil, "en", translations)
			actual, err := localize(key, args...)
			So(err, ShouldBeNil)
			So(actual, ShouldEqual, expected)
		}
		test("key", "Hello John", "John")
	})
}
