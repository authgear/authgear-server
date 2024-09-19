package authflowv2

import (
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
	List(userID string) ([]*identity.Biometric, error)
}

type AuthflowV2SettingsBiometricHandler struct {
	Database                 *appdb.Handle
	ControllerFactory        handlerwebapp.ControllerFactory
	BaseViewModel            *viewmodels.BaseViewModeler
	Renderer                 handlerwebapp.Renderer
	Identities               handlerwebapp.SettingsIdentityService
	BiometricProvider        BiometricIdentityProvider
	AccountManagementService *accountmanagement.Service
}

func (h *AuthflowV2SettingsBiometricHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	userID := session.GetUserID(r.Context())

	// BaseViewModel
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	// BiometricViewModel
	var biometricIdentityInfos []*identity.Biometric
	err := h.Database.WithTx(func() (err error) {
		biometricIdentityInfos, err = h.BiometricProvider.List(*userID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	biometricViewModel := SettingsBiometricViewModel{}

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
	defer ctrl.ServeWithoutDBTx()

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsV2BiometricHTML, data)

		return nil
	})

	ctrl.PostAction("remove", func() error {
		identityID := r.Form.Get("x_identity_id")

		s := session.GetSession(r.Context())

		input := &accountmanagement.RemoveIdentityBiometricInput{
			IdentityID: identityID,
		}
		_, err = h.AccountManagementService.RemoveIdentityBiometric(s, input)
		if err != nil {
			return err
		}

		redirectURI := httputil.HostRelative(r.URL).String()
		result := webapp.Result{RedirectURI: redirectURI}
		result.WriteResponse(w, r)

		return nil
	})
}
