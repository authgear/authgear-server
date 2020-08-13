package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/intents"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/template"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

const (
	TemplateItemTypeAuthUIEnterLoginIDHTML config.TemplateItemType = "auth_ui_enter_login_id.html"
)

var TemplateAuthUIEnterLoginIDHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIEnterLoginIDHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
}

type EnterLoginIDViewModel struct {
	LoginIDKey       string
	LoginIDType      string
	LoginIDInputType string
	IdentityID       string
}

func NewEnterLoginIDViewModel(r *http.Request) EnterLoginIDViewModel {
	loginIDKey := r.Form.Get("x_login_id_key")
	loginIDType := r.Form.Get("x_login_id_type")
	loginIDInputType := r.Form.Get("x_login_id_input_type")
	identityID := r.Form.Get("x_identity_id")

	return EnterLoginIDViewModel{
		LoginIDKey:       loginIDKey,
		LoginIDType:      loginIDType,
		LoginIDInputType: loginIDInputType,
		IdentityID:       identityID,
	}
}

const RemoveLoginIDRequest = "RemoveLoginIDRequest"

var EnterLoginIDSchema = validation.NewMultipartSchema("").
	Add(RemoveLoginIDRequest, `
		{
			"type": "object",
			"properties": {
				"x_login_id_key": { "type": "string" },
				"x_identity_id": { "type": "string" }
			},
			"required": ["x_login_id_key", "x_identity_id"]
		}
	`).Instantiate()

func ConfigureEnterLoginIDRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/enter_login_id")
}

type EnterLoginIDHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
}

func (h *EnterLoginIDHandler) GetData(r *http.Request, state *webapp.State) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	var anyError interface{}
	if state != nil {
		anyError = state.Error
	}

	baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
	enterLoginIDViewModel := NewEnterLoginIDViewModel(r)

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, enterLoginIDViewModel)
	return data, nil
}

type EnterLoginIDRemoveLoginID struct {
	IdentityID string
}

func (i *EnterLoginIDRemoveLoginID) GetIdentityType() authn.IdentityType {
	return authn.IdentityTypeLoginID
}

func (i *EnterLoginIDRemoveLoginID) GetIdentityID() string {
	return i.IdentityID
}

type EnterLoginIDLoginID struct {
	LoginIDType string
	LoginIDKey  string
	LoginID     string
}

var _ nodes.InputUseIdentityLoginID = &EnterLoginIDLoginID{}
var _ nodes.InputCreateAuthenticatorOOBSetup = &EnterLoginIDLoginID{}

// GetLoginIDKey implements InputUseIdentityLoginID.
func (i *EnterLoginIDLoginID) GetLoginIDKey() string {
	return i.LoginIDKey
}

// GetLoginIDKey implements InputUseIdentityLoginID.
func (i *EnterLoginIDLoginID) GetLoginID() string {
	return i.LoginID
}

func (i *EnterLoginIDLoginID) GetOOBChannel() authn.AuthenticatorOOBChannel {
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
func (i *EnterLoginIDLoginID) GetOOBTarget() string {
	return i.LoginID
}

func (h *EnterLoginIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID := session.GetUserID(r.Context())

	if r.Method == "GET" {
		h.Database.WithTx(func() error {
			state, err := h.WebApp.GetState(StateID(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			data, err := h.GetData(r, state)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUIEnterLoginIDHTML, data)
			return nil
		})
	}

	if r.Method == "POST" && r.Form.Get("x_action") == "remove" {
		h.Database.WithTx(func() error {
			enterLoginIDViewModel := NewEnterLoginIDViewModel(r)

			intent := &webapp.Intent{
				StateID:     StateID(r),
				RedirectURI: "/settings/identity",
				Intent:      intents.NewIntentRemoveIdentity(*userID),
			}

			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = &EnterLoginIDRemoveLoginID{
					IdentityID: enterLoginIDViewModel.IdentityID,
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

	if r.Method == "POST" && r.Form.Get("x_action") == "add_or_update" {
		h.Database.WithTx(func() error {
			enterLoginIDViewModel := NewEnterLoginIDViewModel(r)

			intent := &webapp.Intent{
				StateID:     StateID(r),
				RedirectURI: "/settings/identity",
			}

			if enterLoginIDViewModel.IdentityID != "" {
				intent.Intent = intents.NewIntentUpdateIdentity(*userID, enterLoginIDViewModel.IdentityID)
			} else {
				intent.Intent = intents.NewIntentAddIdentity(*userID)
			}

			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				newLoginID, err := FormToLoginID(r.Form)
				if err != nil {
					return nil, err
				}

				input = &EnterLoginIDLoginID{
					LoginIDType: enterLoginIDViewModel.LoginIDType,
					LoginIDKey:  enterLoginIDViewModel.LoginIDKey,
					LoginID:     newLoginID,
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
