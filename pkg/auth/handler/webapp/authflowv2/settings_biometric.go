package authflowv2

import (
	"context"
	"net/http"

	"time"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsV2BiometricHTML = template.RegisterHTML(
	"web/authflowv2/settings_biometric.html",
	handlerwebapp.SettingsComponents...,
)

type BiometricIdentity struct {
	ID          string
	DisplayName string
	CreatedAt   time.Time
}

type SettingsBiometricViewModel struct {
	BiometricIdentities []*BiometricIdentity
}

type BiometricIdentityProvider interface {
	List(ctx context.Context, userID string) ([]*identity.Biometric, error)
}

type AuthflowV2SettingsBiometricHandler struct {
	Database                 *appdb.Handle
	ControllerFactory        handlerwebapp.ControllerFactory
	BaseViewModel            *viewmodels.BaseViewModeler
	SettingsViewModel        *viewmodels.SettingsViewModeler
	Renderer                 handlerwebapp.Renderer
	Identities               handlerwebapp.SettingsIdentityService
	BiometricProvider        BiometricIdentityProvider
	AccountManagementService *accountmanagement.Service
}

func (h *AuthflowV2SettingsBiometricHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
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

	// BiometricViewModel
	biometricViewModel := SettingsBiometricViewModel{}

	var biometricIdentityInfos []*identity.Biometric
	biometricIdentityInfos, err = h.BiometricProvider.List(ctx, *userID)
	if err != nil {
		return nil, err
	}

	for _, biometricInfo := range biometricIdentityInfos {
		displayName := biometricInfo.FormattedDeviceInfo()
		biometricViewModel.BiometricIdentities = append(biometricViewModel.BiometricIdentities, &BiometricIdentity{
			ID:          biometricInfo.ID,
			DisplayName: displayName,
			CreatedAt:   biometricInfo.CreatedAt,
		})
	}

	viewmodels.Embed(data, biometricViewModel)

	return data, nil
}

func (h *AuthflowV2SettingsBiometricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
			return err
		})
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsV2BiometricHTML, data)

		return nil
	})

	ctrl.PostAction("remove", func(ctx context.Context) error {
		identityID := r.Form.Get("x_identity_id")

		s := session.GetSession(ctx)

		input := &accountmanagement.DeleteIdentityBiometricInput{
			IdentityID: identityID,
		}
		_, err = h.AccountManagementService.DeleteIdentityBiometric(ctx, s, input)
		if err != nil {
			return err
		}

		redirectURI := httputil.HostRelative(r.URL).String()
		result := webapp.Result{RedirectURI: redirectURI}
		result.WriteResponse(w, r)

		return nil
	})
}
