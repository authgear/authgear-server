package webapp

import (
	"net/http"
	"time"

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

var TemplateWebSettingsBiometricHTML = template.RegisterHTML(
	"web/settings_biometric.html",
	components...,
)

func ConfigureSettingsBiometricRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/biometric")
}

type BiometricIdentity struct {
	ID          string
	DisplayName string
	CreatedAt   time.Time
}

type SettingsBiometricViewModel struct {
	BiometricIdentities []*BiometricIdentity
}

type SettingsBiometricHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	Identities        SettingsIdentityService
}

func (h *SettingsBiometricHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	userID := session.GetUserID(r.Context())

	identityInfos, err := h.Identities.ListByUser(*userID)
	if err != nil {
		return nil, err
	}

	identityInfos = identity.ApplyFilters(
		identityInfos,
		identity.KeepType(authn.IdentityTypeBiometric),
	)

	viewModel := SettingsBiometricViewModel{}
	for _, info := range identityInfos {
		displayName := info.Claims[identity.IdentityClaimBiometricFormattedDeviceInfo].(string)
		viewModel.BiometricIdentities = append(viewModel.BiometricIdentities, &BiometricIdentity{
			ID:          info.ID,
			DisplayName: displayName,
			CreatedAt:   info.CreatedAt,
		})
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

func (h *SettingsBiometricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	userID := ctrl.RequireUserID()

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsBiometricHTML, data)
		return nil
	})

	redirectURI := httputil.HostRelative(r.URL).String()

	ctrl.PostAction("remove", func() error {
		opts := webapp.SessionOptions{
			RedirectURI: redirectURI,
		}
		identityID := r.Form.Get("x_identity_id")
		intent := intents.NewIntentRemoveIdentity(userID)

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			err = RemoveLoginIDSchema.Validator().ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return nil, err
			}

			input = &InputRemoveIdentity{
				Type: authn.IdentityTypeBiometric,
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
}
