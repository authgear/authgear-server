package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/intents"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/core/phone"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/httputil"
	"github.com/authgear/authgear-server/pkg/template"
	"github.com/authgear/authgear-server/pkg/validation"
)

const (
	TemplateItemTypeAuthUILoginHTML config.TemplateItemType = "auth_ui_login.html"
)

var TemplateAuthUILoginHTML = template.Spec{
	Type:        TemplateItemTypeAuthUILoginHTML,
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
					{{ localize "sign-in-apple" }}
					{{- end -}}
					{{- if eq .provider_type "google" -}}
					{{ localize "sign-in-google" }}
					{{- end -}}
					{{- if eq .provider_type "facebook" -}}
					{{ localize "sign-in-facebook" }}
					{{- end -}}
					{{- if eq .provider_type "linkedin" -}}
					{{ localize "sign-in-linkedin" }}
					{{- end -}}
					{{- if eq .provider_type "azureadv2" -}}
					{{ localize "sign-in-azureadv2" }}
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

				{{ if $.x_login_id_input_type }}{{ if eq $.x_login_id_input_type "phone" }}{{ if $.LoginPageLoginIDHasPhone }}
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
				{{ end }}{{ end }}{{ end }}

				{{ if $.x_login_id_input_type }}{{ if not (eq $.x_login_id_input_type "phone") }}{{ if (not (eq $.LoginPageTextLoginIDVariant "none")) }}
				<input class="input text-input primary-txt" type="{{ $.LoginPageTextLoginIDInputType }}" name="x_login_id" placeholder="{{ localize "login-id-placeholder" $.LoginPageTextLoginIDVariant }}">
				{{ end }}{{ end }}{{ end }}

				{{ if $.x_login_id_input_type }}{{ if eq $.x_login_id_input_type "phone" }}{{ if (not (eq $.LoginPageTextLoginIDVariant "none")) }}
				<a class="link align-self-flex-start" href="{{ call $.MakeURLWithQuery "x_login_id_input_type" $.LoginPageTextLoginIDInputType }}">{{ localize "use-text-login-id-description" $.LoginPageTextLoginIDVariant }}</a>
				{{ end }}{{ end }}{{ end }}

				{{ if $.x_login_id_input_type }}{{ if not (eq $.x_login_id_input_type "phone") }}{{ if $.LoginPageLoginIDHasPhone }}
				<a class="link align-self-flex-start" href="{{ call $.MakeURLWithQuery "x_login_id_input_type" "phone" }}">{{ localize "use-phone-login-id-description" }}</a>
				{{ end }}{{ end }}{{ end }}

				<div class="link">
					<span class="primary-text">{{ localize "signup-button-hint" }}</span>
					<a href="{{ call $.MakeURLWithPathWithoutX "/signup" }}">{{ localize "signup-button-label" }}</a>
				</div>

				{{ if $.ForgotPasswordEnabled }}
				<a class="link align-self-flex-start" href="{{ call $.MakeURLWithPathWithoutX "/forgot_password" }}">{{ localize "forgot-password-button-label" }}</a>
				{{ end }}

				{{ if or $.LoginPageLoginIDHasPhone (not (eq $.LoginPageTextLoginIDVariant "none")) }}
				<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>
				{{ end }}
			</form>
		</div>
		{{ template "auth_ui_footer.html" . }}
	</div>
</body>
</html>
`,
}

const LoginWithLoginIDRequestSchema = "LoginWithLoginIDRequestSchema"

var LoginSchema = validation.NewMultipartSchema("").
	Add(LoginWithLoginIDRequestSchema, `
	{
		"type": "object",
		"properties": {
			"x_login_id_input_type": { "type": "string", "enum": ["email", "phone", "text"] },
			"x_calling_code": { "type": "string" },
			"x_national_number": { "type": "string" },
			"x_login_id": { "type": "string" }
		},
		"required": ["x_login_id_input_type"],
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

func ConfigureLoginRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/login")
}

type LoginHandler struct {
	ServerConfig  *config.ServerConfig
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	FormPrefiller *FormPrefiller
	Renderer      Renderer
	WebApp        WebAppService
}

func (h *LoginHandler) GetData(r *http.Request, state *webapp.State, graph *newinteraction.Graph, edges []newinteraction.Edge) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	var anyError interface{}
	if state != nil {
		anyError = state.Error
	}
	baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
	viewmodels.EmbedForm(data, r.Form)
	viewmodels.Embed(data, baseViewModel)
	authenticationViewModel := viewmodels.NewAuthenticationViewModel(edges)
	viewmodels.Embed(data, authenticationViewModel)
	return data, nil
}

type LoginOAuth struct {
	ProviderAlias    string
	State            string
	NonceSource      *http.Cookie
	ErrorRedirectURI string
}

var _ nodes.InputSelectIdentityOAuthProvider = &LoginOAuth{}

func (i *LoginOAuth) GetProviderAlias() string {
	return i.ProviderAlias
}

func (i *LoginOAuth) GetState() string {
	return i.State
}

func (i *LoginOAuth) GetNonceSource() *http.Cookie {
	return i.NonceSource
}

func (i *LoginOAuth) GetErrorRedirectURI() string {
	return i.ErrorRedirectURI
}

type LoginLoginID struct {
	LoginIDKey string
	LoginID    string
}

var _ nodes.InputSelectIdentityLoginID = &LoginLoginID{}

// GetLoginIDKey implements InputSelectIdentityLoginID.
func (i *LoginLoginID) GetLoginIDKey() string {
	return i.LoginIDKey
}

// GetLoginID implements InputSelectIdentityLoginID.
func (i *LoginLoginID) GetLoginID() string {
	return i.LoginID
}

// GetOOBTarget implements InputAuthenticationOOBTrigger.
func (i *LoginLoginID) GetOOBTarget() string {
	return i.LoginID
}

func (h *LoginHandler) MakeIntent(r *http.Request) *webapp.Intent {
	return &webapp.Intent{
		RedirectURI: webapp.GetRedirectURI(r, h.ServerConfig.TrustProxy),
		Intent:      &intents.IntentLogin{},
	}
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	intent := h.MakeIntent(r)

	h.FormPrefiller.Prefill(r.Form)

	if r.Method == "GET" {
		h.Database.WithTx(func() error {
			state, graph, edges, err := h.WebApp.GetIntent(intent, StateID(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			data, err := h.GetData(r, state, graph, edges)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			h.Renderer.Render(w, r, TemplateItemTypeAuthUILoginHTML, data)
			return nil
		})
	}

	providerAlias := r.Form.Get("x_provider_alias")

	if r.Method == "POST" && providerAlias != "" {
		h.Database.WithTx(func() error {
			nonceSource, _ := r.Cookie(webapp.CSRFCookieName)
			stateID := webapp.NewID()
			intent.StateID = stateID
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = &LoginOAuth{
					ProviderAlias:    providerAlias,
					State:            stateID,
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
				err = LoginSchema.PartValidator(LoginWithLoginIDRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				loginID, err := FormToLoginID(r.Form)
				if err != nil {
					return
				}

				input = &LoginLoginID{
					LoginID: loginID,
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

	return
}

// FormToLoginID returns the raw login ID or the parsed phone number.
func FormToLoginID(form url.Values) (loginID string, err error) {
	if form.Get("x_login_id_input_type") == "phone" {
		nationalNumber := form.Get("x_national_number")
		countryCallingCode := form.Get("x_calling_code")
		var e164 string
		e164, err = phone.Parse(nationalNumber, countryCallingCode)
		if err != nil {
			err = &validation.AggregatedError{
				Errors: []validation.Error{{
					Keyword:  "format",
					Location: "/x_national_number",
					Info: map[string]interface{}{
						"format": "phone",
					},
				}},
			}
			return
		}
		loginID = e164
		return
	}

	loginID = form.Get("x_login_id")
	return
}
