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

var passkeyAutofillTestPhoneData = map[string]any{
	"IdentityCandidates": []map[string]any{
		{"type": "login_id", "login_id_key": "phone", "login_id_type": "phone", "login_id_input_type": "phone"},
	},
	"EmailLoginIDEnabled": false,
	"PhoneLoginIDEnabled": true,
	"LoginIDKey":          "phone",
	"LoginIDInputType":    "phone",
}

func withPasskeyAutofillTestData(datas ...map[string]any) map[string]any {
	out := map[string]any{}
	for _, m := range datas {
		maps.Copy(out, m)
	}
	return out
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

		Convey("offers passkey autofill on the phone input", func() {
			html := renderPasskeyAutofillPage(t, TemplateWebAuthflowLoginHTML, passkeyAutofillTestPhoneData)
			So(html, ShouldContainSubstring, autofillController)
			So(html, ShouldContainSubstring, `inputmode="tel"`)
			So(html, ShouldContainSubstring, `autocomplete="username webauthn"`)
			So(html, ShouldNotContainSubstring, `autocomplete="username"`)
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

		Convey("offers passkey autofill on the phone input in the signup login variant", func() {
			html := renderPasskeyAutofillPage(t, TemplateWebAuthflowV2SignupHTML, withPasskeyAutofillTestData(passkeyAutofillTestPhoneData, map[string]any{
				"UIVariant": AuthflowV2SignupUIVariantSignupLogin,
			}))
			So(html, ShouldContainSubstring, autofillController)
			So(html, ShouldContainSubstring, `inputmode="tel"`)
			So(html, ShouldContainSubstring, `autocomplete="username webauthn"`)
			So(html, ShouldNotContainSubstring, `autocomplete="username"`)
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

		Convey("does not offer passkey autofill on the phone input of the signup page", func() {
			html := renderPasskeyAutofillPage(t, TemplateWebAuthflowV2SignupHTML, withPasskeyAutofillTestData(passkeyAutofillTestPhoneData, map[string]any{
				"UIVariant": AuthflowV2SignupUIVariantSignup,
			}))
			So(html, ShouldContainSubstring, `inputmode="tel"`)
			So(html, ShouldNotContainSubstring, "webauthn")
			So(html, ShouldContainSubstring, `autocomplete="username"`)
		})
	})
}

func TestPasskeyWebViewCheckTemplates(t *testing.T) {
	checkController := `data-controller="authflow-passkey-request passkey-webview-check"`
	checkOptions := `data-passkey-webview-check-options-value="{&#34;publicKey&#34;:{&#34;challenge&#34;:&#34;AQID&#34;}}"`
	checkInput := `name="x_assertion_response" data-authflow-passkey-request-target="input" data-passkey-webview-check-target="input"`
	checkSubmit := `name="x_action" value="passkey" data-authflow-passkey-request-target="submit" data-passkey-webview-check-target="submit"`

	shouldCheck := func(html string) {
		So(html, ShouldContainSubstring, checkController)
		So(html, ShouldContainSubstring, checkOptions)
		So(html, ShouldContainSubstring, checkInput)
		So(html, ShouldContainSubstring, checkSubmit)
	}
	shouldNotCheck := func(html string) {
		So(html, ShouldNotContainSubstring, "passkey-webview-check")
	}

	Convey("login.html", t, func() {
		Convey("asks the WebView for a passkey with the options of the flow", func() {
			html := renderPasskeyAutofillPage(t, TemplateWebAuthflowLoginHTML, nil)
			shouldCheck(html)
		})

		Convey("asks the WebView for a passkey on a page without a login ID input", func() {
			html := renderPasskeyAutofillPage(t, TemplateWebAuthflowLoginHTML, map[string]any{
				"IdentityCandidates": []map[string]any{
					{"type": "passkey"},
				},
				"EmailLoginIDEnabled": false,
			})
			shouldCheck(html)
			So(html, ShouldNotContainSubstring, `data-controller="passkey-autofill"`)
		})

		Convey("does not ask the WebView for a passkey in inline preview", func() {
			html := renderPasskeyAutofillPage(t, TemplateWebAuthflowLoginHTML, map[string]any{
				"InlinePreview": true,
			})
			So(html, ShouldContainSubstring, `data-controller="authflow-passkey-request"`)
			shouldNotCheck(html)
		})

		Convey("does not ask the WebView for a passkey without the passkey button", func() {
			html := renderPasskeyAutofillPage(t, TemplateWebAuthflowLoginHTML, map[string]any{
				"PasskeyEnabled": false,
			})
			shouldNotCheck(html)
		})
	})

	Convey("signup.html", t, func() {
		// signup.html shows the passkey button among other alternatives only.
		withAlternatives := func(overrides map[string]any) map[string]any {
			return withPasskeyAutofillTestData(map[string]any{
				"IdentityCandidates": []map[string]any{
					{"type": "login_id", "login_id_key": "email", "login_id_type": "email", "login_id_input_type": "email"},
					{"type": "login_id", "login_id_key": "username", "login_id_type": "username", "login_id_input_type": "text"},
				},
				"UsernameLoginIDEnabled": true,
			}, overrides)
		}

		Convey("asks the WebView for a passkey in the signup login variant", func() {
			html := renderPasskeyAutofillPage(t, TemplateWebAuthflowV2SignupHTML, withAlternatives(map[string]any{
				"UIVariant": AuthflowV2SignupUIVariantSignupLogin,
			}))
			shouldCheck(html)
		})

		Convey("does not ask the WebView for a passkey in inline preview", func() {
			html := renderPasskeyAutofillPage(t, TemplateWebAuthflowV2SignupHTML, withAlternatives(map[string]any{
				"UIVariant":     AuthflowV2SignupUIVariantSignupLogin,
				"InlinePreview": true,
			}))
			So(html, ShouldContainSubstring, `data-controller="authflow-passkey-request"`)
			shouldNotCheck(html)
		})

		Convey("does not ask the WebView for a passkey on the signup page", func() {
			html := renderPasskeyAutofillPage(t, TemplateWebAuthflowV2SignupHTML, withAlternatives(map[string]any{
				"UIVariant": AuthflowV2SignupUIVariantSignup,
			}))
			So(html, ShouldContainSubstring, `data-controller="authflow-passkey-request"`)
			shouldNotCheck(html)
		})
	})
}
