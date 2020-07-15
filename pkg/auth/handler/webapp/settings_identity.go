package webapp

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/httputil"
	"github.com/authgear/authgear-server/pkg/template"
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

  {{ range .IdentityCandidates }}
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
      {{ $.CSRFField }}
      <input type="hidden" name="x_provider_alias" value="{{ .provider_alias }}">
      {{ if .provider_subject_id }}
      <button class="btn destructive-btn" type="submit" name="x_action" value="unlink_oauth">{{ localize "disconnect-button-label" }}</button>
      {{ else }}
      <button class="btn primary-btn" type="submit" name="x_action" value="link_oauth" data-form-xhr="false">{{ localize "connect-button-label" }}</button>
      {{ end }}
      </form>
    {{ end }}

    {{ if eq .type "login_id" }}
      <form method="post" novalidate>
      {{ $.CSRFField }}
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

func ConfigureSettingsIdentityRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/identity")
}

type SettingsIdentityInteractions interface {
	BeginOAuth(state *interactionflows.State, opts interactionflows.BeginOAuthOptions) (*interactionflows.WebAppResult, error)
	UnlinkOAuthProvider(state *interactionflows.State, providerAlias string, userID string) (*interactionflows.WebAppResult, error)
}

type SettingsIdentityHandler struct {
	Database                *db.Handle
	State                   StateService
	BaseViewModel           *BaseViewModeler
	AuthenticationViewModel *AuthenticationViewModeler
	Renderer                Renderer
	Interactions            SettingsIdentityInteractions
	Responder               Responder
}

func (h *SettingsIdentityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		state, err := h.State.RestoreReadOnlyState(r, true)
		if errors.Is(err, interactionflows.ErrStateNotFound) {
			err = nil
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var anyError interface{}
		if state != nil {
			anyError = state.Error
		}

		baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
		authenticationViewModel := h.AuthenticationViewModel.ViewModel(r)

		data := map[string]interface{}{}

		Embed(data, baseViewModel)
		Embed(data, authenticationViewModel)

		h.Renderer.Render(w, r, TemplateItemTypeAuthUISettingsIdentityHTML, data)
		return
	}

	providerAlias := r.Form.Get("x_provider_alias")

	if r.Method == "POST" && r.Form.Get("x_action") == "link_oauth" {
		h.Database.WithTx(func() error {
			state := interactionflows.NewState()
			state = h.State.CreateState(state, "/settings/identity")
			var result *interactionflows.WebAppResult
			var err error

			defer func() {
				h.State.UpdateState(state, result, err)
				h.Responder.Respond(w, r, state, result, err)
			}()

			sess := auth.GetSession(r.Context())
			userID := sess.AuthnAttrs().UserID

			nonceSource, _ := r.Cookie(webapp.CSRFCookieName)
			result, err = h.Interactions.BeginOAuth(state, interactionflows.BeginOAuthOptions{
				ProviderAlias:    providerAlias,
				Action:           interactionflows.OAuthActionLink,
				UserID:           userID,
				NonceSource:      nonceSource,
				ErrorRedirectURI: httputil.HostRelative(r.URL).String(),
			})
			if err != nil {
				return err
			}

			return nil
		})
	}

	if r.Method == "POST" && r.Form.Get("x_action") == "unlink_oauth" {
		h.Database.WithTx(func() error {
			state := interactionflows.NewState()
			state = h.State.CreateState(state, "/settings/identity")
			var result *interactionflows.WebAppResult
			var err error

			defer func() {
				h.State.UpdateState(state, result, err)
				h.Responder.Respond(w, r, state, result, err)
			}()

			sess := auth.GetSession(r.Context())
			userID := sess.AuthnAttrs().UserID

			result, err = h.Interactions.UnlinkOAuthProvider(state, providerAlias, userID)
			if err != nil {
				return err
			}

			return nil
		})
	}

	loginIDKey := r.Form.Get("x_login_id_key")
	loginIDType := r.Form.Get("x_login_id_type")
	loginIDInputType := r.Form.Get("x_login_id_input_type")
	oldLoginIDValue := r.Form.Get("x_old_login_id_value")

	if r.Method == "POST" && r.Form.Get("x_action") == "login_id" {
		state := interactionflows.NewState()
		state = h.State.CreateState(state, "/settings/identity")
		state.Extra[interactionflows.ExtraLoginIDKey] = loginIDKey
		state.Extra[interactionflows.ExtraLoginIDType] = loginIDType
		state.Extra[interactionflows.ExtraLoginIDInputType] = loginIDInputType
		state.Extra[interactionflows.ExtraOldLoginID] = oldLoginIDValue
		h.State.UpdateState(state, nil, nil)

		redirectURI := state.RedirectURI(r.URL)
		redirectURI.Path = "/enter_login_id"
		http.Redirect(w, r, redirectURI.String(), http.StatusFound)
	}
}
