package webapp

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
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
	ListCandidates(userID string) ([]identity.Candidate, error)
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

var RemoveLoginIDSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_identity_id": { "type": "string" }
		},
		"required": ["x_identity_id"]
	}
`)

var AddOrUpdateLoginIDSchema = validation.NewSimpleSchema(`
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
`)

func ConfigureEnterLoginIDRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/enter_login_id")
}

type EnterLoginIDHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	Identities        EnterLoginIDService
}

func (h *EnterLoginIDHandler) GetData(userID string, r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	identityID := r.Form.Get("x_identity_id")

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	var enterLoginIDViewModel EnterLoginIDViewModel
	if identityID != "" {
		idnInfo, err := h.Identities.Get(userID, authn.IdentityTypeLoginID, identityID)
		if errors.Is(err, identity.ErrIdentityNotFound) {
			return nil, webapp.ErrInvalidSession
		} else if err != nil {
			return nil, err
		}
		enterLoginIDViewModel = NewEnterLoginIDViewModel(r, idnInfo.DisplayID())
	} else {
		enterLoginIDViewModel = NewEnterLoginIDViewModel(r, "")
	}

	candidates, err := h.Identities.ListCandidates(userID)
	if err != nil {
		return nil, err
	}
	authenticationViewModel := viewmodels.NewAuthenticationViewModelWithCandidates(candidates)
	viewmodels.Embed(data, authenticationViewModel)

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, enterLoginIDViewModel)
	return data, nil
}

func (h *EnterLoginIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	userID := ctrl.RequireUserID()

	ctrl.Get(func() error {
		data, err := h.GetData(userID, r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebEnterLoginIDHTML, data)
		return nil
	})

	ctrl.PostAction("remove", func() error {
		opts := webapp.SessionOptions{
			RedirectURI: "/settings",
		}
		identityID := r.Form.Get("x_identity_id")
		intent := intents.NewIntentRemoveIdentity(userID)

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			err = RemoveLoginIDSchema.Validator().ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return nil, err
			}

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

	ctrl.PostAction("add_or_update", func() error {
		opts := webapp.SessionOptions{
			RedirectURI: "/settings",
		}
		loginIDKey := r.Form.Get("x_login_id_key")
		loginIDType := r.Form.Get("x_login_id_type")
		identityID := r.Form.Get("x_identity_id")
		var intent interaction.Intent
		if identityID != "" {
			intent = intents.NewIntentUpdateIdentity(userID, identityID)
		} else {
			intent = intents.NewIntentAddIdentity(userID)
		}

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			err = AddOrUpdateLoginIDSchema.Validator().ValidateValue(FormToJSON(r.Form))
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
}
