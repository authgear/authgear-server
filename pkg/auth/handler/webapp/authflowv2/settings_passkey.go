package authflowv2

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateSettingsV2PasskeyHTML = template.RegisterHTML(
	"web/authflowv2/settings_passkey.html",
	handlerwebapp.SettingsComponents...,
)

func ConfigureAuthflowV2SettingsPasskeyRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/passkey")
}

type AuthflowV2SettingsPasskeyViewModel struct {
	PasskeyIdentities   []*identity.Info
	CreationOptionsJSON string
}

type PasskeyCreationOptionsService interface {
	MakeCreationOptions(ctx context.Context, userID string) (*model.WebAuthnCreationOptions, error)
}

type AuthflowV2SettingsChangePasskeyHandler struct {
	Database          *appdb.Handle
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	SettingsViewModel *viewmodels.SettingsViewModeler
	Renderer          handlerwebapp.Renderer
	Identities        identityservice.Service
	AccountManagement *accountmanagement.Service
	Passkey           PasskeyCreationOptionsService
}

func (h *AuthflowV2SettingsChangePasskeyHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	userID := session.GetUserID(ctx)

	// BaseViewModel
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	// SettingsViewModel
	settingsViewModel, err := h.SettingsViewModel.ViewModel(ctx, *userID)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, *settingsViewModel)

	// PasskeyViewModel
	var passkeyIdentities []*identity.Info
	var creationOptionsJSON string
	identities, err := h.Identities.Passkey.List(ctx, *userID)
	if err != nil {
		return nil, err
	}
	for _, i := range identities {
		passkeyIdentities = append(passkeyIdentities, i.ToInfo())
	}

	creationOptions, err := h.Passkey.MakeCreationOptions(ctx, *userID)
	if err != nil {
		return nil, err
	}
	creationOptionsJSONBytes, err := json.Marshal(creationOptions)
	if err != nil {
		return nil, err
	}
	creationOptionsJSON = string(creationOptionsJSONBytes)

	passkeyViewModel := AuthflowV2SettingsPasskeyViewModel{
		PasskeyIdentities:   passkeyIdentities,
		CreationOptionsJSON: creationOptionsJSON,
	}
	viewmodels.Embed(data, passkeyViewModel)

	return data, nil
}

func (h *AuthflowV2SettingsChangePasskeyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx(r.Context())

	ctrl.Get(func(ctx context.Context) error {
		var data map[string]interface{}
		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			data, err = h.GetData(ctx, r, w)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateSettingsV2PasskeyHTML, data)

		return nil
	})

	ctrl.PostAction("add", func(ctx context.Context) error {
		attestationResponseStr := r.Form.Get("x_attestation_response")

		var creationResponse protocol.CredentialCreationResponse
		err := json.Unmarshal([]byte(attestationResponseStr), &creationResponse)
		if err != nil {
			return err
		}

		s := session.GetSession(ctx)

		input := &accountmanagement.AddPasskeyInput{
			CreationResponse: &creationResponse,
		}

		_, err = h.AccountManagement.AddPasskey(ctx, s, input)
		if err != nil {
			return err
		}

		redirectURI := httputil.HostRelative(r.URL).String()
		result := webapp.Result{RedirectURI: redirectURI}
		result.WriteResponse(w, r)

		return nil

	})

	ctrl.PostAction("remove", func(ctx context.Context) error {
		identityID := r.Form.Get("x_identity_id")

		s := session.GetSession(ctx)

		input := &accountmanagement.DeletePasskeyInput{
			IdentityID: identityID,
		}

		_, err = h.AccountManagement.DeletePasskey(ctx, s, input)
		if err != nil {
			return err
		}

		redirectURI := httputil.HostRelative(r.URL).String()
		result := webapp.Result{RedirectURI: redirectURI}
		result.WriteResponse(w, r)

		return nil
	})
}
