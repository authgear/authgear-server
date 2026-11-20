package authflowv2

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPasskeyButtonTemplates(t *testing.T) {
	passkeyButton := `data-authgear-event="authgear.button.use_passkey"`

	Convey("login.html", t, func() {
		Convey("shows the passkey button with one login ID and passkey", func() {
			html := renderPasskeyAutofillPage(t, TemplateWebAuthflowLoginHTML, nil)
			So(html, ShouldContainSubstring, passkeyButton)
		})

		Convey("does not show the passkey button without passkey", func() {
			html := renderPasskeyAutofillPage(t, TemplateWebAuthflowLoginHTML, map[string]any{
				"PasskeyEnabled": false,
			})
			So(html, ShouldNotContainSubstring, passkeyButton)
		})
	})

	Convey("signup.html", t, func() {
		Convey("shows the passkey button in the signup login variant with one login ID and passkey", func() {
			html := renderPasskeyAutofillPage(t, TemplateWebAuthflowV2SignupHTML, map[string]any{
				"UIVariant": AuthflowV2SignupUIVariantSignupLogin,
			})
			So(html, ShouldContainSubstring, passkeyButton)
		})

		Convey("does not show the passkey button in the signup login variant when alternatives are disabled", func() {
			html := renderPasskeyAutofillPage(t, TemplateWebAuthflowV2SignupHTML, map[string]any{
				"UIVariant":            AuthflowV2SignupUIVariantSignupLogin,
				"AlternativesDisabled": true,
			})
			So(html, ShouldNotContainSubstring, passkeyButton)
		})

		Convey("does not show the passkey button in the signup variant", func() {
			for _, overrides := range []map[string]any{
				{},
				// An OAuth provider makes the alternatives section render.
				{"IdentityCandidates": []map[string]any{
					{"type": "login_id", "login_id_key": "email", "login_id_type": "email", "login_id_input_type": "email"},
					{"type": "oauth", "provider_type": "google", "provider_alias": "google", "provider_status": "active"},
				}},
			} {
				overrides["UIVariant"] = AuthflowV2SignupUIVariantSignup
				html := renderPasskeyAutofillPage(t, TemplateWebAuthflowV2SignupHTML, overrides)
				So(html, ShouldNotContainSubstring, passkeyButton)
			}
		})
	})
}
