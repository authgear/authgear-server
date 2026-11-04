package authflowv2

import (
	"context"
	"maps"
	"testing"

	"github.com/spf13/afero"

	runtimeresource "github.com/authgear/authgear-server"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"

	. "github.com/smartystreets/goconvey/convey"
)

const passkeyAutofillTestOptionsJSON = `{"publicKey":{"challenge":"AQID"}}`

func renderPasskeyAutofillPage(t *testing.T, tpl *template.HTML, overrides map[string]any) string {
	t.Helper()

	manager := resource.NewManagerWithDir(resource.NewManagerWithDirOptions{
		Registry:              resource.DefaultRegistry,
		BuiltinResourceFS:     runtimeresource.EmbedFS_resources_authgear,
		BuiltinResourceFSRoot: runtimeresource.RelativePath_resources_authgear,
	})

	// Render the page content only, without the page frame.
	fs := afero.NewMemMapFs()
	err := fs.MkdirAll("templates/en/web/authflowv2", 0o777)
	if err != nil {
		t.Fatal(err)
	}
	err = afero.WriteFile(fs, "templates/en/web/authflowv2/__page_frame.html", []byte(`{{ define "authflowv2/__page_frame.html" }}{{ template "page-content" . }}{{ end }}`), 0o666)
	if err != nil {
		t.Fatal(err)
	}
	manager = manager.Overlay(resource.LeveledAferoFs{
		Fs:      fs,
		FsLevel: resource.FsLevelCustom,
	})

	engine := &template.Engine{
		Resolver: &template.Resolver{
			Resources:             manager,
			DefaultLanguageTag:    template.DefaultLanguageTag("en"),
			SupportedLanguageTags: template.SupportedLanguageTags{"en"},
		},
	}

	data := map[string]any{
		"IdentityCandidates": []map[string]any{
			{"type": "login_id", "login_id_key": "email", "login_id_type": "email", "login_id_input_type": "email"},
		},
		"EmailLoginIDEnabled":       true,
		"LoginIDKey":                "email",
		"LoginIDInputType":          "email",
		"NonPhoneLoginIDInputType":  "email",
		"NonPhoneLoginIDType":       "email",
		"PasskeyEnabled":            true,
		"PasskeyRequestOptionsJSON": passkeyAutofillTestOptionsJSON,
		"FormsByIdentification":     map[string]FormViewModel{},
		"MakeURL":                   func(path string, pairs ...string) string { return path },
	}
	maps.Copy(data, overrides)

	result, err := engine.Render(context.Background(), tpl, []string{"en"}, data)
	if err != nil {
		t.Fatal(err)
	}
	return result.String
}

func TestPasskeyAutofillTemplates(t *testing.T) {
	autofillController := `data-controller="passkey-autofill"`
	autofillOptions := `data-passkey-autofill-options-value="{&#34;publicKey&#34;:{&#34;challenge&#34;:&#34;AQID&#34;}}"`

	Convey("login.html", t, func() {
		Convey("offers passkey autofill with the options of the flow", func() {
			html := renderPasskeyAutofillPage(t, TemplateWebAuthflowLoginHTML, nil)
			So(html, ShouldContainSubstring, autofillController)
			So(html, ShouldContainSubstring, autofillOptions)
			So(html, ShouldContainSubstring, `autocomplete="username webauthn"`)
		})

		Convey("does not offer passkey autofill without the passkey button", func() {
			for _, overrides := range []map[string]any{
				{"PasskeyEnabled": false},
				{"AlternativesDisabled": true},
				{"InlinePreview": true},
			} {
				html := renderPasskeyAutofillPage(t, TemplateWebAuthflowLoginHTML, overrides)
				So(html, ShouldNotContainSubstring, autofillController)
			}
		})
	})

	Convey("signup.html", t, func() {
		Convey("offers passkey autofill in the signup login variant", func() {
			html := renderPasskeyAutofillPage(t, TemplateWebAuthflowV2SignupHTML, map[string]any{
				"UIVariant": AuthflowV2SignupUIVariantSignupLogin,
			})
			So(html, ShouldContainSubstring, autofillController)
			So(html, ShouldContainSubstring, autofillOptions)
			So(html, ShouldContainSubstring, `autocomplete="username webauthn"`)
		})

		Convey("does not offer passkey autofill without the passkey button", func() {
			for _, overrides := range []map[string]any{
				{"PasskeyEnabled": false},
				{"AlternativesDisabled": true},
				{"InlinePreview": true},
			} {
				overrides["UIVariant"] = AuthflowV2SignupUIVariantSignupLogin
				html := renderPasskeyAutofillPage(t, TemplateWebAuthflowV2SignupHTML, overrides)
				So(html, ShouldNotContainSubstring, autofillController)
			}
		})

		Convey("does not offer passkey autofill on the signup page", func() {
			html := renderPasskeyAutofillPage(t, TemplateWebAuthflowV2SignupHTML, map[string]any{
				"UIVariant": AuthflowV2SignupUIVariantSignup,
			})
			So(html, ShouldNotContainSubstring, autofillController)
			So(html, ShouldNotContainSubstring, "webauthn")
			So(html, ShouldContainSubstring, `autocomplete="username"`)
		})
	})
}
