package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/httputil"
	"github.com/authgear/authgear-server/pkg/template"
)

const (
	TemplateItemTypeAuthUIPromoteHTML config.TemplateItemType = "auth_ui_promote.html"
)

var TemplateAuthUIPromoteHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIPromoteHTML,
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
		<div class="authorize-form">
			<div class="authorize-idp-section">
				{{ range $.IdentityCandidates }}
				{{ if eq .type "oauth" }}
				<form class="authorize-idp-form" method="post" novalidate>
				{{ $.CSRFField }}
				<button class="btn sso-btn {{ .provider_type }}" type="submit" name="x_provider_alias" value="{{ .provider_alias }}" data-form-xhr="false">
					{{- if eq .provider_type "apple" -}}
					{{ localize "sign-up-apple" }}
					{{- end -}}
					{{- if eq .provider_type "google" -}}
					{{ localize "sign-up-google" }}
					{{- end -}}
					{{- if eq .provider_type "facebook" -}}
					{{ localize "sign-up-facebook" }}
					{{- end -}}
					{{- if eq .provider_type "linkedin" -}}
					{{ localize "sign-up-linkedin" }}
					{{- end -}}
					{{- if eq .provider_type "azureadv2" -}}
					{{ localize "sign-up-azureadv2" }}
					{{- end -}}
				</button>
				</form>
				{{ end }}
				{{ end }}
			</div>

			{{ $has_oauth := false }}
			{{ $has_login_id := false }}
			{{ range $.IdentityCandidates }}
				{{ if eq .type "oauth" }}
				{{ $has_oauth = true }}
				{{ end }}
				{{ if eq .type "login_id" }}
				{{ $has_login_id = true }}
				{{ end }}
			{{ end }}
			{{ if $has_oauth }}{{ if $has_login_id }}
			<div class="primary-txt sso-loginid-separator">{{ localize "sso-login-id-separator" }}</div>
			{{ end }}{{ end }}

			{{ template "ERROR" . }}

			<form class="authorize-loginid-form" method="post" novalidate>
				{{ $.CSRFField }}
				<input type="hidden" name="x_login_id_key" value="{{ .x_login_id_key }}">

				{{ range $.IdentityCandidates }}
				{{ if eq .type "login_id" }}{{ if eq .login_id_key $.x_login_id_key }}
				{{ if eq .login_id_type "phone" }}
					<div class="phone-input">
						<select class="input select primary-txt" name="x_calling_code">
							{{ range $.CountryCallingCodes }}
							<option
								value="{{ . }}"
								{{ if $.x_calling_code }}{{ if eq $.x_calling_code . }}
								selected
								{{ end }}{{ end }}
								>
								+{{ . }}
							</option>
							{{ end }}
						</select>
						<input class="input text-input primary-txt" type="text" inputmode="numeric" pattern="[0-9]*" name="x_national_number" placeholder="{{ localize "phone-number-placeholder" }}">
					</div>
				{{ else }}
					<input class="input text-input primary-txt" type="{{ $.x_login_id_input_type }}" name="x_login_id" placeholder="{{ .login_id_type }}">
				{{ end }}
				{{ end }}{{ end }}
				{{ end }}

				{{ range $.IdentityCandidates }}
				{{ if eq .type "login_id" }}{{ if not (eq .login_id_key $.x_login_id_key) }}
					<a class="link align-self-flex-start"
						href="{{ call $.MakeURLWithQuery "x_login_id_key" .login_id_key "x_login_id_input_type" .login_id_input_type}}">
						{{ localize "use-login-id-key" .login_id_key }}
					</a>
				{{ end }}{{ end }}
				{{ end }}

				<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">
					{{ localize "next-button-label" }}
				</button>
			</form>
		</div>
		{{ template "auth_ui_footer.html" . }}
	</div>
</body>
</html>
`,
}

func ConfigurePromoteRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/promote_user")
}

type PromoteInteractions interface {
	BeginOAuth(state *interactionflows.State, opts interactionflows.BeginOAuthOptions) (*interactionflows.WebAppResult, error)
	PromoteWithLoginID(state *interactionflows.State, loginIDKey string, loginID string, userID string) (*interactionflows.WebAppResult, error)
}

type PromoteHandler struct {
	Database                *db.Handle
	State                   StateService
	BaseViewModel           *BaseViewModeler
	AuthenticationViewModel *AuthenticationViewModeler
	FormPrefiller           *FormPrefiller
	Renderer                Renderer
	Interactions            PromoteInteractions
	Responder               Responder
}

func (h *PromoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.FormPrefiller.Prefill(r.Form)

	if r.Method == "GET" {
		state, err := h.State.RestoreReadOnlyState(r, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		baseViewModel := h.BaseViewModel.ViewModel(r, state.Error)
		authenticationViewModel := h.AuthenticationViewModel.ViewModel(r)

		data := map[string]interface{}{}

		EmbedForm(data, r.Form)
		Embed(data, baseViewModel)
		Embed(data, authenticationViewModel)

		h.Renderer.Render(w, r, TemplateItemTypeAuthUIPromoteHTML, data)
		return
	}

	providerAlias := r.Form.Get("x_provider_alias")

	if r.Method == "POST" && providerAlias != "" {
		h.Database.WithTx(func() error {
			state, err := h.State.CloneState(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			var result *interactionflows.WebAppResult
			defer func() {
				h.State.UpdateState(state, result, err)
				h.Responder.Respond(w, r, state, result, err)
			}()

			userID, _ := state.Extra[interactionflows.ExtraAnonymousUserID].(string)

			nonceSource, _ := r.Cookie(webapp.CSRFCookieName)
			result, err = h.Interactions.BeginOAuth(state, interactionflows.BeginOAuthOptions{
				ProviderAlias:    providerAlias,
				Action:           interaction.OAuthActionPromote,
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

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			state, err := h.State.CloneState(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			var result *interactionflows.WebAppResult
			defer func() {
				h.State.UpdateState(state, result, err)
				h.Responder.Respond(w, r, state, result, err)
			}()

			userID, _ := state.Extra[interactionflows.ExtraAnonymousUserID].(string)
			loginID, err := FormToLoginID(r.Form)
			if err != nil {
				return err
			}

			loginIDKey := r.Form.Get("x_login_id_key")
			result, err = h.Interactions.PromoteWithLoginID(state, loginIDKey, loginID, userID)
			if err != nil {
				return err
			}

			return nil
		})
	}
}
