package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/template"
	"github.com/authgear/authgear-server/pkg/validation"
)

const (
	TemplateItemTypeAuthUISettingsIdentityHTML config.TemplateItemType = "auth_ui_settings_identity.html"
)

var TemplateAuthUISettingsIdentityHTML = template.Spec{
	Type:        TemplateItemTypeAuthUISettingsIdentityHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
	Default: `<!DOCTYPE html>
<html>
{{ template "auth_ui_html_head.html" . }}
<body class="page">
<div class="content">

{{ template "auth_ui_header.html" . }}

<div class="settings-identity">
  <h1 class="title primary-txt">{{ localize "settings-identity-title" }}</h1>

  {{ template "ERROR" . }}

  {{ range .x_identity_candidates }}
  <div class="identity">
    <div class="icon {{ .type }} {{ .provider_type }} {{ .login_id_type }}"></div>
    <div class="identity-info flex-child-no-overflow">
      <h2 class="identity-name primary-txt">
         {{ if eq .type "oauth" }}
           {{ if eq .provider_type "google" }}
           {{ localize "settings-identity-oauth-google" }}
           {{ end }}
           {{ if eq .provider_type "apple" }}
           {{ localize "settings-identity-oauth-apple" }}
           {{ end }}
           {{ if eq .provider_type "facebook" }}
           {{ localize "settings-identity-oauth-facebook" }}
           {{ end }}
           {{ if eq .provider_type "linkedin" }}
           {{ localize "settings-identity-oauth-linkedin" }}
           {{ end }}
           {{ if eq .provider_type "azureadv2" }}
           {{ localize "settings-identity-oauth-azureadv2" }}
           {{ end }}
         {{ end }}
         {{ if eq .type "login_id" }}
           {{ if eq .login_id_type "email" }}
           {{ localize "settings-identity-login-id-email" }}
           {{ end }}
           {{ if eq .login_id_type "phone" }}
           {{ localize "settings-identity-login-id-phone" }}
           {{ end }}
           {{ if eq .login_id_type "username" }}
           {{ localize "settings-identity-login-id-username" }}
           {{ end }}
           {{ if eq .login_id_type "raw" }}
           {{ localize "settings-identity-login-id-raw" }}
           {{ end }}
         {{ end }}
      </h2>

      {{ if eq .type "oauth" }}{{ if .email }}
      <h3 class="identity-claim secondary-txt text-ellipsis">
        {{ .email }}
      </h3>
      {{ end }}{{ end }}

      {{ if eq .type "login_id" }}{{ if .login_id_value }}
      <h3 class="identity-claim secondary-txt text-ellipsis">
        {{ .login_id_value }}
      </h3>
      {{ end }}{{ end }}
    </div>

    {{ if eq .type "oauth" }}
      <form method="post" novalidate>
      {{ $.csrfField }}
      <input type="hidden" name="x_idp_id" value="{{ .provider_alias }}">
      {{ if .provider_subject_id }}
      <button class="btn destructive-btn" type="submit" name="x_action" value="unlink">{{ localize "disconnect-button-label" }}</button>
      {{ else }}
      <button class="btn primary-btn" type="submit" name="x_action" value="link" data-form-xhr="false">{{ localize "connect-button-label" }}</button>
      {{ end }}
      </form>
    {{ end }}

    {{ if eq .type "login_id" }}
      <form method="post" novalidate>
      {{ $.csrfField }}
      <input type="hidden" name="x_login_id_key" value="{{ .login_id_key }}">
      <input type="hidden" name="x_login_id_type" value="{{ .login_id_type }}">
      {{ if eq .login_id_type "phone" }}
      <input type="hidden" name="x_login_id_input_type" value="phone">
      {{ else if eq .login_id_type "email" }}
      <input type="hidden" name="x_login_id_input_type" value="email">
      {{ else }}
      <input type="hidden" name="x_login_id_input_type" value="text">
      {{ end }}
      {{ if .login_id_value }}
      <input type="hidden" name="x_old_login_id_value" value="{{ .login_id_value }}">
      <button class="btn secondary-btn" type="submit" name="x_action" value="login_id">{{ localize "change-button-label" }}</a>
      {{ else }}
      <button class="btn primary-btn" type="submit" name="x_action" value="login_id">{{ localize "connect-button-label" }}</a>
      {{ end }}
      </form>
    {{ end }}
  </div>
  {{ end }}
</div>

{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

const AddOrChangeLoginIDRequest = "AddOrChangeLoginIDRequest"

var SettingsIdentitySchema = validation.NewMultipartSchema("").
	Add(AddOrChangeLoginIDRequest, `
		{
			"type": "object",
			"properties": {
				"x_login_id_key": { "type": "string" },
				"x_login_id_input_type": { "type": "string", "enum": ["email", "phone", "text"] }
			},
			"required": ["x_login_id_key", "x_login_id_input_type"]
		}
	`).Instantiate()

func ConfigureSettingsIdentityRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/identity")
}

type SettingsIdentityHandler struct {
	Database *db.Handle
}

func (h *SettingsIdentityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Database.WithTx(func() error {
		// FIXME(webapp): settings_identity
		// if r.Method == "GET" {
		// 	writeResponse, err := h.Provider.GetSettingsIdentity(w, r)
		// 	writeResponse(err)
		// 	return err
		// }

		// if r.Method == "POST" {
		// 	if r.Form.Get("x_action") == "link" {
		// 		writeResponse, err := h.Provider.LinkIdentityProvider(w, r, r.Form.Get("x_idp_id"))
		// 		writeResponse(err)
		// 		return err
		// 	}
		// 	if r.Form.Get("x_action") == "unlink" {
		// 		writeResponse, err := h.Provider.UnlinkIdentityProvider(w, r, r.Form.Get("x_idp_id"))
		// 		writeResponse(err)
		// 		return err
		// 	}
		// 	if r.Form.Get("x_action") == "login_id" {
		// 		writeResponse, err := h.Provider.AddOrChangeLoginID(w, r)
		// 		writeResponse(err)
		// 		return err
		// 	}
		// }

		return nil
	})
}
