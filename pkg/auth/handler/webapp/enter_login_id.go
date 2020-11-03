package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebEnterLoginIDHTML = template.RegisterHTML(
	"web/enter_login_id.html",
	components...,
)

type EnterLoginIDViewModel struct {
	LoginIDKey       string
	LoginIDType      string
	LoginIDInputType string
	IdentityID       string
	DisplayID        string
}

type EnterLoginIDService interface {
	Get(userID string, typ authn.IdentityType, id string) (*identity.Info, error)
}

func NewEnterLoginIDViewModel(r *http.Request, displayID string) EnterLoginIDViewModel {
	loginIDKey := r.Form.Get("x_login_id_key")
	loginIDType := r.Form.Get("x_login_id_type")
	loginIDInputType := r.Form.Get("x_login_id_input_type")
	identityID := r.Form.Get("x_identity_id")

	return EnterLoginIDViewModel{
		LoginIDKey:       loginIDKey,
		LoginIDType:      loginIDType,
		LoginIDInputType: loginIDInputType,
		IdentityID:       identityID,
		DisplayID:        displayID,
	}
}

const RemoveLoginIDRequest = "RemoveLoginIDRequest"
const AddOrUpdateLoginIDRequest = "AddOrUpdateLoginIDRequest"

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
	`).
	Add(AddOrUpdateLoginIDRequest, `
		{
			"type": "object",
			"properties": {
				"x_login_id_input_type": { "type": "string" },
				"x_login_id_key": { "type": "string" },
				"x_login_id_type": { "type": "string" },
				"x_calling_code": { "type": "string" },
				"x_national_number": { "type": "string" },
				"x_login_id": { "type": "string" }
			},
			"required": ["x_login_id_input_type", "x_login_id_key", "x_login_id_type"],
			"allOf": [
				{
					"if": {
						"properties": {
							"x_login_id_key": { "type": "string", "const": "phone" }
						}
					},
					"then": {
						"required": ["x_calling_code", "x_national_number"]
					}
				},
				{
					"if": {
						"properties": {
							"x_login_id_key": { "type": "string", "enum": ["username", "email"] }
						}
					},
					"then": {
						"required": ["x_login_id"]
					}
				}
			]
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
	Identities    EnterLoginIDService
}

func (h *EnterLoginIDHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	userID := session.GetUserID(r.Context())
	identityID := r.Form.Get("x_identity_id")

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	var enterLoginIDViewModel EnterLoginIDViewModel
	if identityID != "" {
		idnInfo, err := h.Identities.Get(*userID, authn.IdentityTypeLoginID, identityID)
		if err != nil {
			return nil, err
		}
		enterLoginIDViewModel = NewEnterLoginIDViewModel(r, idnInfo.DisplayID())
	} else {
		enterLoginIDViewModel = NewEnterLoginIDViewModel(r, "")
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, enterLoginIDViewModel)
	return data, nil
}

func (h *EnterLoginIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := session.GetUserID(r.Context())

	if r.Method == "GET" {
		err := h.Database.WithTx(func() error {
			data, err := h.GetData(r, w)
			if err != nil {
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateWebEnterLoginIDHTML, data)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	if r.Method == "POST" && r.Form.Get("x_action") == "remove" {
		err := h.Database.WithTx(func() error {
			identityID := r.Form.Get("x_identity_id")

			intent := &webapp.Intent{
				RedirectURI: "/settings/identity",
				Intent:      intents.NewIntentRemoveIdentity(*userID),
			}

			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = &InputRemoveIdentity{
					Type: authn.IdentityTypeLoginID,
					ID:   identityID,
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

	if r.Method == "POST" && r.Form.Get("x_action") == "add_or_update" {
		err := h.Database.WithTx(func() error {
			loginIDKey := r.Form.Get("x_login_id_key")
			loginIDType := r.Form.Get("x_login_id_type")
			identityID := r.Form.Get("x_identity_id")

			intent := &webapp.Intent{
				RedirectURI: "/settings/identity",
			}

			if identityID != "" {
				intent.Intent = intents.NewIntentUpdateIdentity(*userID, identityID)
			} else {
				intent.Intent = intents.NewIntentAddIdentity(*userID)
			}

			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				err = EnterLoginIDSchema.PartValidator(AddOrUpdateLoginIDRequest).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return nil, err
				}

				newLoginID, err := FormToLoginID(r.Form)
				if err != nil {
					return nil, err
				}

				input = &InputNewLoginID{
					LoginIDType:  loginIDType,
					LoginIDKey:   loginIDKey,
					LoginIDValue: newLoginID,
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
