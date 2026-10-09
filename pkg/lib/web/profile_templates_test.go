package web_test

import (
	"context"
	"encoding/json"
	htmltemplate "html/template"
	"maps"
	"strings"
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

var formFieldsTestPage = template.RegisterHTML("web/authflowv2/__form_fields_test_page.html", web.ComponentsHTML...)

func TestFormFieldsTemplate(t *testing.T) {
	Convey("__form_fields.html", t, func() {
		builtin := afero.NewBasePathFs(afero.NewOsFs(), "../../../resources/authgear")
		overlay := afero.NewMemMapFs()
		_ = overlay.MkdirAll("templates/en/web/authflowv2", 0777)
		_ = afero.WriteFile(overlay, "templates/en/web/authflowv2/__form_fields_test_page.html", []byte(
			`{{ template "authflowv2/__alert_message.html" (dict "Type" "error" "Message" (include "authflowv2/__form_error_message.html" (merge (dict "Form" $.Form) $))) }}`+
				`|FIELDS:{{ template "authflowv2/__form_fields.html" (merge (dict "Form" $.Form) $) }}`), 0666)
		orig, err := afero.ReadFile(builtin, "templates/en/translation.json")
		So(err, ShouldBeNil)
		var translations map[string]string
		So(json.Unmarshal(orig, &translations), ShouldBeNil)
		translations["standard-attribute-pattern-error-/address/postal_code"] = "Postal code format"
		translations["custom-attribute-pattern-error-/x_employee_id"] = "Employee ID format"
		merged, err := json.Marshal(translations)
		So(err, ShouldBeNil)
		_ = afero.WriteFile(overlay, "templates/en/translation.json", merged, 0666)
		manager := resource.NewManager(resource.DefaultRegistry, []resource.Fs{
			resource.LeveledAferoFs{Fs: afero.NewCopyOnWriteFs(builtin, overlay), FsLevel: resource.FsLevelBuiltin},
		})
		engine := &template.Engine{Resolver: &template.Resolver{
			Resources: manager, DefaultLanguageTag: "en", SupportedLanguageTags: []string{"en"},
		}}
		ctx := context.Background()
		tr, err := engine.Translation(ctx, []string{"en"})
		So(err, ShouldBeNil)

		field := func(key string, typ string, extra map[string]any) map[string]any {
			f := map[string]any{
				"Key": key, "Type": typ, "Label": "L " + key, "Required": false,
				"Value": nil, "Checked": false, "Address": map[string]string{},
				"Options": nil, "Suggestions": nil, "Constraints": nil,
				"Error": nil, "MissingRequired": false, "HookMessage": "",
			}
			maps.Copy(f, extra)
			return f
		}
		renderForm := func(form map[string]any) string {
			out, err := engine.Render(ctx, formFieldsTestPage, []string{"en"}, map[string]any{
				"Translations": tr,
				"Form":         form,
			})
			So(err, ShouldBeNil)
			return out.String
		}
		render := func(fields ...map[string]any) string {
			return renderForm(map[string]any{
				"Fields":    fields,
				"Timezones": nil,
				"Alpha2":    []string{"HK"},
			})
		}

		Convey("every type renders", func() {
			two := 2
			out := render(
				field("/given_name", "string", map[string]any{"Required": true, "Value": "John", "Suggestions": []string{"a"},
					"Constraints": &config.StringConstraints{MinLength: &two}}),
				field("/x_email", "email", nil),
				field("/website", "url", nil),
				field("/x_age", "integer", nil),
				field("/x_score", "number", nil),
				field("/x_phone", "phone_number", nil),
				field("/birthdate", "date", nil),
				field("/x_plan", "enum", map[string]any{"Options": []map[string]any{{"Value": "free", "Label": "Free"}}}),
				field("/x_country", "country_code", nil),
				field("/zoneinfo", "timezone", nil),
				field("/address", "address", map[string]any{"Address": map[string]string{"locality": "London"}}),
				field("/x_terms", "boolean", map[string]any{"Checked": true, "Required": true, "LabelHTML": htmltemplate.HTML(`I accept <a href="https://example.com">Terms</a>`)}),
			)
			So(out, ShouldContainSubstring, `name="/given_name"`)
			So(out, ShouldContainSubstring, `value="John"`)
			So(out, ShouldContainSubstring, "Enter at least 2 characters")
			So(out, ShouldContainSubstring, `name="/address/locality" value="London"`)
			So(out, ShouldContainSubstring, `name="/x_terms"`)
			So(out, ShouldContainSubstring, `<a href="https://example.com">Terms</a>`)
			So(out, ShouldContainSubstring, `checked`)
			So(out, ShouldContainSubstring, `value="false"`)

			// Every input has an accessible name.
			So(out, ShouldContainSubstring, `id="form-field-label-/given_name"`)
			So(out, ShouldContainSubstring, `name="/given_name"
            aria-labelledby="form-field-label-/given_name"`)
			So(out, ShouldContainSubstring, `aria-labelledby="form-field-label-/x_phone"`)
			So(out, ShouldContainSubstring, `aria-labelledby="form-field-label-/birthdate"`)
			So(out, ShouldContainSubstring, `role="group" aria-labelledby="form-field-label-/x_plan"`)
			So(out, ShouldContainSubstring, `aria-label="Postal code"`)
		})

		Convey("errors", func() {
			out := render(
				field("/a", "string", map[string]any{"HookMessage": "From the hook"}),
				field("/b", "string", map[string]any{"MissingRequired": true, "Required": true}),
				field("/c", "string", map[string]any{"Error": map[string]any{
					"reason": "ValidationFailed",
					"info": map[string]any{"causes": []any{
						map[string]any{"location": "/fields/0/value", "kind": "minLength", "details": map[string]any{"expected": 2, "actual": 1}},
					}},
				}}),
			)
			So(out, ShouldContainSubstring, "From the hook")
			So(out, ShouldContainSubstring, "This field is required")
			So(out, ShouldContainSubstring, "greater than or equal to 2 characters")
		})

		Convey("the form error keeps the hook error as HTML and escapes unmatched messages", func() {
			out := renderForm(map[string]any{
				"Fields": nil,
				"UnknownError": map[string]any{
					"reason": "HookDisallowed",
					"info":   map[string]any{"reasons": []any{map[string]any{"title": "Hook title", "reason": "Hook reason"}}},
				},
				"HookUnmatchedMessages": []string{"<b>Bad other</b>"},
			})
			// The visible message, not the data-error-message attribute, where escaping is expected.
			_, message, found := strings.Cut(out, `data-alert-message-target="message">`)
			So(found, ShouldBeTrue)
			So(message, ShouldContainSubstring, "Hook title")
			So(message, ShouldContainSubstring, "<br />")
			So(message, ShouldContainSubstring, "<span>&lt;b&gt;Bad other&lt;/b&gt;</span>")
			So(message, ShouldNotContainSubstring, "&lt;span&gt;")
			So(message, ShouldNotContainSubstring, "&lt;br")
		})

		Convey("pattern errors use the message of the attribute or sub-attribute", func() {
			patternError := func(location string) map[string]any {
				return map[string]any{
					"reason": "ValidationFailed",
					"info": map[string]any{"causes": []any{
						map[string]any{"location": location, "kind": "pattern", "details": map[string]any{}},
					}},
				}
			}
			out := render(
				field("/x_employee_id", "string", map[string]any{"Error": patternError("/fields/0/value")}),
				field("/address", "address", map[string]any{"Error": patternError("/fields/1/value/postal_code")}),
			)
			So(out, ShouldContainSubstring, "Employee ID format")
			So(out, ShouldContainSubstring, "Postal code format")
		})
	})
}
