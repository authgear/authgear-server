package webapp

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsOOBOTPHTML = template.RegisterHTML(
	"web/settings_oob_otp.html",
	Components...,
)

func ConfigureSettingsOOBOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/mfa/oob_otp_:channel")
}

type SettingsOOBOTPViewModel struct {
	OOBAuthenticatorType model.AuthenticatorType
	Authenticators       []*authenticator.Info
}

type SettingsOOBOTPHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	Authenticators    SettingsAuthenticatorService
}

func (h *SettingsOOBOTPHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	userID := session.GetUserID(ctx)
	viewModel := SettingsOOBOTPViewModel{}
	oc := httproute.GetParam(r, "channel")
	t, err := model.Deprecated_GetOOBAuthenticatorType(model.AuthenticatorOOBChannel(oc))
	if err != nil {
		return nil, err
	}
	authenticators, err := h.Authenticators.List(ctx, *userID,
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
	oobAuthenticatorType, err := model.Deprecated_GetOOBAuthenticatorType(model.AuthenticatorOOBChannel(oc))
	if err != nil {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}
	defer ctrl.ServeWithDBTx(r.Context())

	redirectURI := httputil.HostRelative(r.URL).String()
	authenticatorID := r.Form.Get("x_authenticator_id")
	userID := ctrl.RequireUserID(r.Context())

	ctrl.Get(func(ctx context.Context) error {
		data, err := h.GetData(ctx, r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsOOBOTPHTML, data)
		return nil
	})

	ctrl.PostAction("remove", func(ctx context.Context) error {
		opts := webapp.SessionOptions{
			RedirectURI: redirectURI,
		}
		intent := intents.NewIntentRemoveAuthenticator(userID)

		result, err := ctrl.EntryPointPost(ctx, opts, intent, func() (input interface{}, err error) {
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

	ctrl.PostAction("add", func(ctx context.Context) error {
		opts := webapp.SessionOptions{
			RedirectURI: redirectURI,
		}
		intent := intents.NewIntentAddAuthenticator(
			userID,
			authn.AuthenticationStageSecondary,
			oobAuthenticatorType,
		)

		result, err := ctrl.EntryPointPost(ctx, opts, intent, func() (input interface{}, err error) {
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
