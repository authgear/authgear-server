package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsPasskeyHTML = template.RegisterHTML(
	"web/settings_passkey.html",
	Components...,
)

func ConfigureSettingsPasskeyRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/passkey")
}

type SettingsPasskeyViewModel struct {
	PasskeyIdentities []*identity.Info
}

type SettingsPasskeyHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	Identities        SettingsIdentityService
}

func (h *SettingsPasskeyHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	userID := session.GetUserID(r.Context())

	identities, err := h.Identities.ListByUser(*userID)
	if err != nil {
		return nil, err
	}
	var passkeyIdentities []*identity.Info
	for _, i := range identities {
		if i.Type == model.IdentityTypePasskey {
			ii := i
			passkeyIdentities = append(passkeyIdentities, ii)
		}
	}

	viewModel := SettingsPasskeyViewModel{
		PasskeyIdentities: passkeyIdentities,
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

func (h *SettingsPasskeyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

	redirectURI := httputil.HostRelative(r.URL).String()
	userID := ctrl.RequireUserID()

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsPasskeyHTML, data)
		return nil
	})

	ctrl.PostAction("remove", func() error {
		opts := webapp.SessionOptions{
			RedirectURI: redirectURI,
		}
		identityID := r.Form.Get("q_identity_id")
		intent := intents.NewIntentRemoveIdentity(userID)
		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			input = &InputRemoveIdentity{
				Type: model.IdentityTypePasskey,
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

	ctrl.PostAction("add", func() error {
		opts := webapp.SessionOptions{
			RedirectURI: redirectURI,
		}
		intent := intents.NewIntentAddAuthenticator(
			userID,
			authn.AuthenticationStagePrimary,
			model.AuthenticatorTypePasskey,
		)

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			attestationResponseStr := r.Form.Get("x_attestation_response")
			attestationResponse := []byte(attestationResponseStr)

			return &InputPasskeyAttestationResponse{
				Stage:               string(authn.AuthenticationStagePrimary),
				AttestationResponse: attestationResponse,
			}, nil
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
}
