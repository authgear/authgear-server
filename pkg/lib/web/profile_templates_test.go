package web_test

import (
	"context"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var profileTestPage = template.RegisterHTML("web/authflowv2/__profile_test_page.html", web.ComponentsHTML...)

// TestProfileTemplates renders the real templates, so that mistakes in them fail here instead of at runtime.
func TestProfileTemplates(t *testing.T) {
	Convey("profile templates", t, func() {
		builtin := afero.NewBasePathFs(afero.NewOsFs(), "../../../resources/authgear")
		overlay := afero.NewMemMapFs()
		_ = overlay.MkdirAll("templates/en/web/authflowv2", 0777)
		_ = afero.WriteFile(overlay, "templates/en/web/authflowv2/__profile_test_page.html", []byte(
			`{{ template "authflowv2/__string_constraints_hint.html" (dict "Constraints" .Constraints) }}`+
				`|ERROR:{{ include "authflowv2/__error.html" . }}`+
				`|LINK:{{ sanitizeLinksHTML (include "test-label" nil) }}`+
				`|TEXT:{{ linksToText (include "test-label" nil) }}`,
		), 0666)

		orig, err := afero.ReadFile(builtin, "templates/en/translation.json")
		So(err, ShouldBeNil)
		var translations map[string]string
		So(json.Unmarshal(orig, &translations), ShouldBeNil)
		translations["test-label"] = `I accept the <a href="https://example.com/t">Terms</a><script>x</script>`
		translations["standard-attribute-pattern-error-/given_name"] = "Letters only please"
		merged, err := json.Marshal(translations)
		So(err, ShouldBeNil)
		_ = afero.WriteFile(overlay, "templates/en/translation.json", merged, 0666)

		manager := resource.NewManager(resource.DefaultRegistry, []resource.Fs{
			resource.LeveledAferoFs{Fs: afero.NewCopyOnWriteFs(builtin, overlay), FsLevel: resource.FsLevelBuiltin},
		})
		engine := &template.Engine{Resolver: &template.Resolver{
			Resources:             manager,
			DefaultLanguageTag:    "en",
			SupportedLanguageTags: []string{"en"},
		}}
		ctx := context.Background()
		tr, err := engine.Translation(ctx, []string{"en"})
		So(err, ShouldBeNil)

		render := func(constraints *config.StringConstraints, causes ...map[string]any) string {
			errCauses := []any{}
			for _, c := range causes {
				errCauses = append(errCauses, c)
			}
			out, err := engine.Render(ctx, profileTestPage, []string{"en"}, map[string]any{
				"Constraints":  constraints,
				"Translations": tr,
				"Error": map[string]any{
					"reason": "ValidationFailed",
					"info":   map[string]any{"causes": errCauses},
				},
			})
			So(err, ShouldBeNil)
			return out.String
		}
		cause := func(location string, kind string, details map[string]any) map[string]any {
			return map[string]any{"location": location, "kind": kind, "details": details}
		}
		two, ten := 2, 10

		Convey("constraints hint", func() {
			out := render(&config.StringConstraints{
				MinLength:         &two,
				MaxLength:         &ten,
				AllowedCharacters: []config.AllowedCharacters{"letters", "digits"},
			})
			So(out, ShouldContainSubstring, "Enter 2 to 10 characters")
			So(out, ShouldContainSubstring, "Allowed characters: letters, digits")

			So(render(&config.StringConstraints{MinLength: &two}), ShouldContainSubstring, "Enter at least 2 characters")
			So(render(&config.StringConstraints{MaxLength: &ten}), ShouldContainSubstring, "Enter at most 10 characters")
			So(render(nil), ShouldNotContainSubstring, "characters")
		})

		Convey("constraint errors", func() {
			out := render(nil,
				cause("/a", "minLength", map[string]any{"expected": 2, "actual": 1}),
				cause("/a", "allowedCharacters", map[string]any{"expected": []any{"letters", "digits"}}),
				cause("/given_name", "pattern", map[string]any{"expected": "x"}),
				cause("/other", "pattern", map[string]any{"expected": "x"}),
			)
			So(out, ShouldContainSubstring, "greater than or equal to 2 characters")
			So(out, ShouldContainSubstring, "Only the following characters are allowed: letters, digits")
			So(out, ShouldContainSubstring, "Letters only please")
			So(out, ShouldContainSubstring, "The value is not in the expected format")
		})

		Convey("label links", func() {
			out := render(nil)
			So(out, ShouldContainSubstring, `I accept the <a href="https://example.com/t" target="_blank" rel="noopener noreferrer">Terms</a>&lt;script&gt;x&lt;/script&gt;`)
			So(out, ShouldContainSubstring, "|TEXT:I accept the Termsx")
		})
	})
}
