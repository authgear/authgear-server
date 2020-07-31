package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/intents"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/template"
	"github.com/authgear/authgear-server/pkg/validation"
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

func NewEnterLoginIDViewModel(state *webapp.State) EnterLoginIDViewModel {
	loginIDKey, _ := state.Extra["x_login_id_key"].(string)
	loginIDType, _ := state.Extra["x_login_id_type"].(string)
	loginIDInputType, _ := state.Extra["x_login_id_input_type"].(string)
	identityID, _ := state.Extra["x_identity_id"].(string)

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

func (h *EnterLoginIDHandler) GetData(r *http.Request, state *webapp.State, graph *newinteraction.Graph, edges []newinteraction.Edge) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, state.Error)
	enterLoginIDViewModel := NewEnterLoginIDViewModel(state)

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

// FIXME(webapp): implement input interface
type EnterLoginIDUpdateLoginID struct {
	IdentityID string
	LoginIDKey string
	LoginID    string
}

type EnterLoginIDAddLoginID struct {
	LoginIDKey string
	LoginID    string
}

// GetLoginIDKey implements InputUseIdentityLoginID.
func (i *EnterLoginIDAddLoginID) GetLoginIDKey() string {
	return i.LoginIDKey
}

// GetLoginIDKey implements InputUseIdentityLoginID.
func (i *EnterLoginIDAddLoginID) GetLoginID() string {
	return i.LoginID
}

func (h *EnterLoginIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID := auth.GetSession(r.Context()).AuthnAttrs().UserID

	if r.Method == "GET" {
		h.Database.WithTx(func() error {
			state, graph, edges, err := h.WebApp.Get(StateID(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			data, err := h.GetData(r, state, graph, edges)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			h.Renderer.Render(w, r, TemplateItemTypeAuthUIEnterLoginIDHTML, data)
			return nil
		})
	}

	if r.Method == "POST" && r.Form.Get("x_action") == "remove" {
		h.Database.WithTx(func() error {
			state, _, _, err := h.WebApp.Get(StateID(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			enterLoginIDViewModel := NewEnterLoginIDViewModel(state)

			intent := state.NewIntent()
			intent.Intent = intents.NewIntentRemoveIdentity(userID)

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
			state, _, _, err := h.WebApp.Get(StateID(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			enterLoginIDViewModel := NewEnterLoginIDViewModel(state)

			intent := state.NewIntent()
			intent.Intent = intents.NewIntentAddIdentity(userID)

			newLoginID := r.Form.Get("x_login_id")

			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				if enterLoginIDViewModel.IdentityID != "" {
					input = &EnterLoginIDUpdateLoginID{
						IdentityID: enterLoginIDViewModel.IdentityID,
						LoginIDKey: enterLoginIDViewModel.LoginIDKey,
						LoginID:    newLoginID,
					}
				} else {
					input = &EnterLoginIDAddLoginID{
						LoginIDKey: enterLoginIDViewModel.LoginIDKey,
						LoginID:    newLoginID,
					}
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
