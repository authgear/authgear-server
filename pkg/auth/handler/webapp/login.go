package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebLoginHTML = template.RegisterHTML(
	"web/login.html",
	components...,
)

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
	TrustProxy    config.TrustProxy
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	FormPrefiller *FormPrefiller
	Renderer      Renderer
	WebApp        WebAppService
	CSRFCookie    webapp.CSRFCookieDef
}

func (h *LoginHandler) GetData(r *http.Request, rw http.ResponseWriter, graph *interaction.Graph) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.EmbedForm(data, r.Form)
	viewmodels.Embed(data, baseViewModel)
	authenticationViewModel := viewmodels.NewAuthenticationViewModelWithGraph(graph)
	viewmodels.Embed(data, authenticationViewModel)
	return data, nil
}

func (h *LoginHandler) MakeIntent(r *http.Request) *webapp.Intent {
	return &webapp.Intent{
		OldStateID:  StateID(r),
		RedirectURI: webapp.GetRedirectURI(r, bool(h.TrustProxy)),
		Intent:      intents.NewIntentLogin(),
	}
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	intent := h.MakeIntent(r)

	h.FormPrefiller.Prefill(r.Form)

	if r.Method == "GET" {
		err := h.Database.WithTx(func() error {
			_, graph, err := h.WebApp.GetIntent(intent)
			if err != nil {
				return err
			}

			data, err := h.GetData(r, w, graph)
			if err != nil {
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateWebLoginHTML, data)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	providerAlias := r.Form.Get("x_provider_alias")

	if r.Method == "POST" && providerAlias != "" {
		err := h.Database.WithTx(func() error {
			nonceSource, _ := r.Cookie(h.CSRFCookie.Name)
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = &InputUseOAuth{
					ProviderAlias:    providerAlias,
					NonceSource:      nonceSource,
					ErrorRedirectURI: httputil.HostRelative(r.URL).String(),
				}
				return
			})
			if err != nil {
				return err
			}
			result.WriteResponse(w, r)
			return nil
		})
		if err != nil {
			panic(err)
		}
		return
	}

	if r.Method == "POST" {
		err := h.Database.WithTx(func() error {
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				err = LoginSchema.PartValidator(LoginWithLoginIDRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				loginID, err := FormToLoginID(r.Form)
				if err != nil {
					return
				}

				input = &InputUseLoginID{
					LoginID: loginID,
				}
				return
			})
			if err != nil {
				return err
			}
			result.WriteResponse(w, r)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
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
