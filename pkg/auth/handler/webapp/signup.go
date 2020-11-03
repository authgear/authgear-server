package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSignupHTML = template.RegisterHTML(
	"web/signup.html",
	components...,
)

const SignupWithLoginIDRequestSchema = "SignupWithLoginIDRequestSchema"

var SignupSchema = validation.NewMultipartSchema("").
	Add(SignupWithLoginIDRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_login_id_key": { "type": "string" },
				"x_login_id_type": { "type": "string" },
				"x_login_id_input_type": { "type": "string", "enum": ["email", "phone", "text"] },
				"x_calling_code": { "type": "string" },
				"x_national_number": { "type": "string" },
				"x_login_id": { "type": "string" }
			},
			"required": ["x_login_id_key", "x_login_id_type", "x_login_id_input_type"],
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
	TrustProxy    config.TrustProxy
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	FormPrefiller *FormPrefiller
	Renderer      Renderer
	WebApp        WebAppService
	CSRFCookie    webapp.CSRFCookieDef
}

func (h *SignupHandler) GetData(r *http.Request, rw http.ResponseWriter, graph *interaction.Graph) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.EmbedForm(data, r.Form)
	viewmodels.Embed(data, baseViewModel)
	authenticationViewModel := viewmodels.NewAuthenticationViewModelWithGraph(graph)
	viewmodels.Embed(data, authenticationViewModel)
	return data, nil
}

type SignupOAuth struct {
	ProviderAlias    string
	NonceSource      *http.Cookie
	ErrorRedirectURI string
}

var _ nodes.InputUseIdentityOAuthProvider = &SignupOAuth{}

func (i *SignupOAuth) GetProviderAlias() string {
	return i.ProviderAlias
}

func (i *SignupOAuth) GetNonceSource() *http.Cookie {
	return i.NonceSource
}

func (i *SignupOAuth) GetErrorRedirectURI() string {
	return i.ErrorRedirectURI
}

type SignupLoginID struct {
	LoginIDType  string
	LoginIDKey   string
	LoginIDValue string
}

var _ nodes.InputUseIdentityLoginID = &SignupLoginID{}
var _ nodes.InputCreateAuthenticatorOOBSetup = &SignupLoginID{}

// GetLoginIDKey implements InputCreateIdentityLoginID.
func (i *SignupLoginID) GetLoginIDKey() string {
	return i.LoginIDKey
}

// GetLoginID implements InputCreateIdentityLoginID.
func (i *SignupLoginID) GetLoginID() string {
	return i.LoginIDValue
}

func (i *SignupLoginID) GetOOBChannel() authn.AuthenticatorOOBChannel {
	switch i.LoginIDType {
	case string(config.LoginIDKeyTypeEmail):
		return authn.AuthenticatorOOBChannelEmail
	case string(config.LoginIDKeyTypePhone):
		return authn.AuthenticatorOOBChannelSMS
	default:
		return ""
	}
}

// GetOOBTarget implements InputCreateAuthenticatorOOBSetup.
func (i *SignupLoginID) GetOOBTarget() string {
	return i.LoginIDValue
}

func (h *SignupHandler) MakeIntent(r *http.Request) *webapp.Intent {
	return &webapp.Intent{
		OldStateID:  StateID(r),
		RedirectURI: webapp.GetRedirectURI(r, bool(h.TrustProxy)),
		Intent:      intents.NewIntentSignup(),
	}
}

func (h *SignupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

			h.Renderer.RenderHTML(w, r, TemplateWebSignupHTML, data)
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
				input = &SignupOAuth{
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
				err = SignupSchema.PartValidator(SignupWithLoginIDRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				loginIDValue, err := FormToLoginID(r.Form)
				if err != nil {
					return
				}

				loginIDKey := r.Form.Get("x_login_id_key")
				loginIDType := r.Form.Get("x_login_id_type")

				input = &SignupLoginID{
					LoginIDType:  loginIDType,
					LoginIDKey:   loginIDKey,
					LoginIDValue: loginIDValue,
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
