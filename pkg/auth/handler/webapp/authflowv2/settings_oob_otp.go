package authflowv2

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	authenticatorservice "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsOOBOTPHTML = template.RegisterHTML(
	"web/authflowv2/settings_oob_otp.html",
	handlerwebapp.SettingsComponents...,
)

type AuthflowV2SettingsOOBOTPViewModel struct {
	OOBOTPType           string
	OOBOTPAuthenticators []*authenticator.OOBOTP
}

type AuthflowV2SettingsOOBOTPHandler struct {
	Database          *appdb.Handle
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	SettingsViewModel *viewmodels.SettingsViewModeler
	Renderer          handlerwebapp.Renderer
	AccountManagement *accountmanagement.Service
	Authenticators    authenticatorservice.Service
}

func (h *AuthflowV2SettingsOOBOTPHandler) GetData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	userID := session.GetUserID(r.Context())

	// BaseViewModel
	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)

	// SettingsViewModel
	settingsViewModel, err := h.SettingsViewModel.ViewModel(*userID)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, *settingsViewModel)

	oc := httproute.GetParam(r, "channel")

	t, err := model.GetOOBAuthenticatorType(model.AuthenticatorOOBChannel(oc))
	if err != nil {
		return nil, err
	}
	authenticators, err := h.Authenticators.List(
		*userID,
		authenticator.KeepKind(authenticator.KindSecondary),
		authenticator.KeepType(t),
	)
	if err != nil {
		return nil, err
	}

	var OOBOTPAuthenticators []*authenticator.OOBOTP
	for _, a := range authenticators {
		OOBOTPAuthenticators = append(OOBOTPAuthenticators, a.OOBOTP)
	}
	vm := AuthflowV2SettingsOOBOTPViewModel{
		OOBOTPType:           oc,
		OOBOTPAuthenticators: OOBOTPAuthenticators,
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsOOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx()

	ctrl.Get(func() error {
		var data map[string]interface{}
		err := h.Database.WithTx(func() error {
			data, err = h.GetData(w, r)
			return err
		})
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsOOBOTPHTML, data)
		return nil
	})

	ctrl.PostAction("remove", func() error {
		authenticatorID := r.Form.Get("x_authenticator_id")

		s := session.GetSession(r.Context())

		input := &accountmanagement.DeleteOOBOTPAuthenticatorInput{
			AuthenticatorID: authenticatorID,
		}
		_, err = h.AccountManagement.DeleteOOBOTPAuthenticator(s, input)
		if err != nil {
			return err
		}

		redirectURI := httputil.HostRelative(r.URL).String()
		result := webapp.Result{RedirectURI: redirectURI}
		result.WriteResponse(w, r)

		return nil
	})
}
