package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/intents"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/httputil"
	"github.com/authgear/authgear-server/pkg/template"
	"github.com/authgear/authgear-server/pkg/validation"
)

const (
	TemplateItemTypeAuthUISignupHTML config.TemplateItemType = "auth_ui_signup.html"
)

var TemplateAuthUISignupHTML = template.Spec{
	Type:        TemplateItemTypeAuthUISignupHTML,
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
				<input type="hidden" name="x_login_id_key" value="{{ $.x_login_id_key }}">

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

				{{ range .IdentityCandidates }}
				{{ if eq .type "login_id" }}{{ if not (eq .login_id_key $.x_login_id_key) }}
					<a class="link align-self-flex-start"
						href="{{ call $.MakeURLWithQuery "x_login_id_key" .login_id_key "x_login_id_input_type" .login_id_input_type}}">
						{{ localize "use-login-id-key" .login_id_key }}
					</a>
				{{ end }}{{ end }}
				{{ end }}

				<div class="link align-self-flex-start">
					<span class="primary-text">{{ localize "login-button-hint" }}</span>
					<a href="{{ call $.MakeURLWithPathWithoutX "/login" }}">{{ localize "login-button-label" }}</a>
				</div>

				{{ if .PasswordAuthenticatorEnabled }}
				<a class="link align-self-flex-start" href="{{ call $.MakeURLWithPathWithoutX "/forgot_password" }}">{{ localize "forgot-password-button-label" }}</a>
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

const SignupWithLoginIDRequestSchema = "SignupWithLoginIDRequestSchema"

var SignupSchema = validation.NewMultipartSchema("").
	Add(SignupWithLoginIDRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_login_id_key": { "type": "string" },
				"x_login_id_input_type": { "type": "string", "enum": ["email", "phone", "text"] },
				"x_calling_code": { "type": "string" },
				"x_national_number": { "type": "string" },
				"x_login_id": { "type": "string" }
			},
			"required": ["x_login_id_key", "x_login_id_input_type"],
			"allOf": [
				{
					"if": {
						"properties": {
							"x_login_id_input_type": { "type": "string", "const": "phone" }
						}
					},
					"then": {
						"required": ["x_calling_code", "x_national_number"]
					}
				},
				{
					"if": {
						"properties": {
							"x_login_id_input_type": { "type": "string", "enum": ["text", "email"] }
						}
					},
					"then": {
						"required": ["x_login_id"]
					}
				}
			]
		}
	`).Instantiate()

func ConfigureSignupRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/signup")
}

type SignupHandler struct {
	ServerConfig  *config.ServerConfig
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	FormPrefiller *FormPrefiller
	Renderer      Renderer
	WebApp        WebAppService
}

func (h *SignupHandler) GetData(r *http.Request, state *webapp.State, graph *newinteraction.Graph, edges []newinteraction.Edge) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	var anyError interface{}
	if state != nil {
		anyError = state.Error
	}
	baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
	viewmodels.EmbedForm(data, r.Form)
	viewmodels.Embed(data, baseViewModel)
	// FIXME(webapp): derive AuthenticationViewModel with graph and edges
	authenticationViewModel := viewmodels.AuthenticationViewModel{}
	viewmodels.Embed(data, authenticationViewModel)
	return data, nil
}

// FIXME(webapp): implement input interface
type SignupOAuth struct {
	ProviderAlias    string
	Action           string
	NonceSource      *http.Cookie
	ErrorRedirectURI string
}

// FIXME(webapp): implement input interface
type SignupLoginID struct {
	LoginIDKey   string
	LoginIDValue string
}

func (h *SignupHandler) MakeIntent(r *http.Request) *webapp.Intent {
	return &webapp.Intent{
		RedirectURI: webapp.GetRedirectURI(r, h.ServerConfig.TrustProxy),
		// FIXME(webapp): Use signup intent
		Intent: &intents.IntentLogin{},
	}
}

func (h *SignupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	intent := h.MakeIntent(r)

	h.FormPrefiller.Prefill(r.Form)

	if r.Method == "GET" {
		state, graph, edges, err := h.WebApp.GetIntent(intent, StateID(r))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := h.GetData(r, state, graph, edges)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		h.Renderer.Render(w, r, TemplateItemTypeAuthUISignupHTML, data)
		return
	}

	providerAlias := r.Form.Get("x_provider_alias")

	if r.Method == "POST" && providerAlias != "" {
		h.Database.WithTx(func() error {
			nonceSource, _ := r.Cookie(webapp.CSRFCookieName)
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = &SignupOAuth{
					ProviderAlias: providerAlias,
					// FIXME(webapp): Use constant
					Action:           "login",
					NonceSource:      nonceSource,
					ErrorRedirectURI: httputil.HostRelative(r.URL).String(),
				}
				return
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}
			result.WriteResponse(w, r)
			return nil
		})
		return
	}

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				err = SignupSchema.PartValidator(SignupWithLoginIDRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				loginIDValue, err := FormToLoginID(r.Form)
				if err != nil {
					return
				}

				loginIDKey := r.Form.Get("x_login_id_key")

				input = &SignupLoginID{
					LoginIDKey:   loginIDKey,
					LoginIDValue: loginIDValue,
				}
				return
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}
			result.WriteResponse(w, r)
			return nil
		})
	}
}
