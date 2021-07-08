package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsOOBOTPHTML = template.RegisterHTML(
	"web/settings_oob_otp.html",
	components...,
)

func ConfigureSettingsOOBOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/mfa/oob_otp_:channel")
}

type SettingsOOBOTPViewModel struct {
	OOBAuthenticatorType authn.AuthenticatorType
	Authenticators       []*authenticator.Info
}

type SettingsOOBOTPHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	Authenticators    SettingsAuthenticatorService
}

func (h *SettingsOOBOTPHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	userID := session.GetUserID(r.Context())
	viewModel := SettingsOOBOTPViewModel{}
	oc := httproute.GetParam(r, "channel")
	t, err := authn.GetOOBAuthenticatorType(authn.AuthenticatorOOBChannel(oc))
	if err != nil {
		return nil, err
	}
	authenticators, err := h.Authenticators.List(*userID,
		authenticator.KeepKind(authenticator.KindSecondary),
		authenticator.KeepType(t),
	)
	if err != nil {
		return nil, err
	}
	viewModel.OOBAuthenticatorType = t
	viewModel.Authenticators = authenticators

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

func (h *SettingsOOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	oc := httproute.GetParam(r, "channel")
	oobAuthenticatorType, err := authn.GetOOBAuthenticatorType(authn.AuthenticatorOOBChannel(oc))
	if err != nil {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}
	defer ctrl.Serve()

	redirectURI := httputil.HostRelative(r.URL).String()
	authenticatorID := r.Form.Get("x_authenticator_id")
	userID := ctrl.RequireUserID()

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsOOBOTPHTML, data)
		return nil
	})

	ctrl.PostAction("remove", func() error {
		opts := webapp.SessionOptions{
			RedirectURI: redirectURI,
		}
		intent := intents.NewIntentRemoveAuthenticator(userID)

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interaction.Input, err error) {
			input = &InputRemoveAuthenticator{
				Type: oobAuthenticatorType,
				ID:   authenticatorID,
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
			authn.AuthenticationStageSecondary,
			oobAuthenticatorType,
		)

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interaction.Input, err error) {
			input = &InputCreateAuthenticator{}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
}
