package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

const (
	TemplateItemTypeAuthUIPromoteHTML string = "auth_ui_promote.html"
)

var TemplateAuthUIPromoteHTML = template.T{
	Type:                    TemplateItemTypeAuthUIPromoteHTML,
	IsHTML:                  true,
	TranslationTemplateType: TemplateItemTypeAuthUITranslationJSON,
	Defines:                 defines,
	ComponentTemplateTypes:  components,
}

const PromoteWithLoginIDRequestSchema = "PromoteWithLoginIDRequestSchema"

var PromoteSchema = validation.NewMultipartSchema("").
	Add(PromoteWithLoginIDRequestSchema, `
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

func ConfigurePromoteRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/promote_user")
}

type PromoteHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	FormPrefiller *FormPrefiller
	Renderer      Renderer
	WebApp        WebAppService
	CSRFCookie    webapp.CSRFCookieDef
}

func (h *PromoteHandler) GetData(r *http.Request, state *webapp.State, graph *interaction.Graph) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	var anyError interface{}
	if state != nil {
		anyError = state.Error
	}
	baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
	viewmodels.EmbedForm(data, r.Form)
	viewmodels.Embed(data, baseViewModel)
	authenticationViewModel := viewmodels.NewAuthenticationViewModelWithGraph(graph)
	viewmodels.Embed(data, authenticationViewModel)
	return data, nil
}

type PromoteOAuth struct {
	ProviderAlias    string
	NonceSource      *http.Cookie
	ErrorRedirectURI string
}

var _ nodes.InputUseIdentityOAuthProvider = &PromoteOAuth{}

func (i *PromoteOAuth) GetProviderAlias() string {
	return i.ProviderAlias
}

func (i *PromoteOAuth) GetNonceSource() *http.Cookie {
	return i.NonceSource
}

func (i *PromoteOAuth) GetErrorRedirectURI() string {
	return i.ErrorRedirectURI
}

type PromoteLoginID struct {
	LoginIDType  string
	LoginIDKey   string
	LoginIDValue string
}

var _ nodes.InputUseIdentityLoginID = &PromoteLoginID{}
var _ nodes.InputCreateAuthenticatorOOBSetup = &PromoteLoginID{}

// GetLoginIDKey implements InputCreateIdentityLoginID.
func (i *PromoteLoginID) GetLoginIDKey() string {
	return i.LoginIDKey
}

// GetLoginID implements InputCreateIdentityLoginID.
func (i *PromoteLoginID) GetLoginID() string {
	return i.LoginIDValue
}

func (i *PromoteLoginID) GetOOBChannel() authn.AuthenticatorOOBChannel {
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
func (i *PromoteLoginID) GetOOBTarget() string {
	return i.LoginIDValue
}

func (h *PromoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.FormPrefiller.Prefill(r.Form)

	if r.Method == "GET" {
		h.Database.WithTx(func() error {
			state, graph, err := h.WebApp.Get(StateID(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			data, err := h.GetData(r, state, graph)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUIPromoteHTML, data)
			return nil
		})
	}

	providerAlias := r.Form.Get("x_provider_alias")

	if r.Method == "POST" && providerAlias != "" {
		h.Database.WithTx(func() error {
			nonceSource, _ := r.Cookie(h.CSRFCookie.Name)
			stateID := StateID(r)
			result, err := h.WebApp.PostInput(stateID, func() (input interface{}, err error) {
				input = &PromoteOAuth{
					ProviderAlias:    providerAlias,
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
	}

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				err = PromoteSchema.PartValidator(PromoteWithLoginIDRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				loginIDValue, err := FormToLoginID(r.Form)
				if err != nil {
					return
				}

				loginIDKey := r.Form.Get("x_login_id_key")
				loginIDType := r.Form.Get("x_login_id_type")

				input = &PromoteLoginID{
					LoginIDType:  loginIDType,
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
