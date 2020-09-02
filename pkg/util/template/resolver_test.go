package template

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type defaultLoader struct {
	Map map[string]string
}

func (l *defaultLoader) LoadDefault(typ string) (string, error) {
	val := l.Map[typ]
	return val, nil
}

type uriLoader struct{}

func (l *uriLoader) Load(uri string) (string, error) {
	return uri, nil
}

func TestResolverResolve(t *testing.T) {
	translation := T{
		Type: "translation",
	}
	componentA := T{
		Type:                    "component-a",
		TranslationTemplateType: "translation",
	}
	componentB := T{
		Type:                    "component-b",
		TranslationTemplateType: "translation",
	}
	pageA := T{
		Type:                    "page-a",
		TranslationTemplateType: "translation",
		ComponentTemplateTypes:  []string{"component-a", "component-b"},
	}
	pageB := T{
		Type:                    "page-b",
		TranslationTemplateType: "translation",
		ComponentTemplateTypes:  []string{"component-a", "component-b"},
	}

	ts := []T{
		translation,
		componentA,
		componentB,
		pageA,
		pageB,
	}

	Convey("Resolver.Resolve", t, func() {
		loader := &defaultLoader{
			Map: map[string]string{
				translation.Type: `
				{
					"key1": "key1value",
					"key2": "key2value"
				}
				`,
				componentA.Type: "component-a-default",
				componentB.Type: "component-b-default",
				pageA.Type:      "page-a-default",
				pageB.Type:      "page-b-default",
			},
		}
		test := func(references []Reference, expected *Resolved) {
			resolver := NewResolver(NewResolverOptions{
				References:          references,
				FallbackLanguageTag: "en",
			})
			resolver.Registry = NewRegistry()
			resolver.Loader = &uriLoader{}
			resolver.DefaultLoader = loader
			for _, t := range ts {
				resolver.Registry.Register(t)
			}

			ctx := &ResolveContext{
				PreferredLanguageTags: []string{"zh", "en"},
			}

			actual, err := resolver.Resolve(ctx, "page-a")
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, expected)
		}

		// Nothing is customized.
		test([]Reference{}, &Resolved{
			T:       pageA,
			Content: "page-a-default",
			Translations: map[string]Translation{
				"key1": Translation{
					LanguageTag: "en",
					Value:       "key1value",
				},
				"key2": Translation{
					LanguageTag: "en",
					Value:       "key2value",
				},
			},
			ComponentContents: []string{
				"component-a-default",
				"component-b-default",
			},
		})

		// One translation key is customized.
		test([]Reference{
			Reference{
				Type:        "translation",
				LanguageTag: "zh",
				URI: `
				{
					"key1": "中文"
				}
				`,
			},
		}, &Resolved{
			T:       pageA,
			Content: "page-a-default",
			Translations: map[string]Translation{
				"key1": Translation{
					LanguageTag: "zh",
					Value:       "中文",
				},
				"key2": Translation{
					LanguageTag: "en",
					Value:       "key2value",
				},
			},
			ComponentContents: []string{
				"component-a-default",
				"component-b-default",
			},
		})

		// One component is customized.
		test([]Reference{
			Reference{
				Type:        "component-a",
				LanguageTag: "zh",
				URI:         "component-a-中文",
			},
		}, &Resolved{
			T:       pageA,
			Content: "page-a-default",
			Translations: map[string]Translation{
				"key1": Translation{
					LanguageTag: "en",
					Value:       "key1value",
				},
				"key2": Translation{
					LanguageTag: "en",
					Value:       "key2value",
				},
			},
			ComponentContents: []string{
				"component-a-中文",
				"component-b-default",
			},
		})

		// The page itself is customized.
		test([]Reference{
			Reference{
				Type:        "page-a",
				LanguageTag: "zh",
				URI:         "page-a-中文",
			},
		}, &Resolved{
			T:       pageA,
			Content: "page-a-中文",
			Translations: map[string]Translation{
				"key1": Translation{
					LanguageTag: "en",
					Value:       "key1value",
				},
				"key2": Translation{
					LanguageTag: "en",
					Value:       "key2value",
				},
			},
			ComponentContents: []string{
				"component-a-default",
				"component-b-default",
			},
		})
	})
}
