package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
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
		WithPathPattern("/settings/oob_otp")
}

type SettingsOOBOTPViewModel struct {
	Authenticators []*authenticator.Info
}

type SettingsOOBOTPHandler struct {
	Database       *db.Handle
	BaseViewModel  *viewmodels.BaseViewModeler
	Renderer       Renderer
	WebApp         WebAppService
	Authenticators SettingsAuthenticatorService
	CSRFCookie     webapp.CSRFCookieDef
}

func (h *SettingsOOBOTPHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	userID := session.GetUserID(r.Context())
	viewModel := SettingsOOBOTPViewModel{}
	authenticators, err := h.Authenticators.List(*userID,
		authenticator.KeepKind(authenticator.KindSecondary),
		authenticator.KeepType(authn.AuthenticatorTypeOOB),
	)
	if err != nil {
		return nil, err
	}
	viewModel.Authenticators = authenticators

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

func (h *SettingsOOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	redirectURI := httputil.HostRelative(r.URL).String()
	authenticatorID := r.Form.Get("x_authenticator_id")
	userID := session.GetUserID(r.Context())

	if r.Method == "GET" {
		err := h.Database.WithTx(func() error {
			data, err := h.GetData(r, w)
			if err != nil {
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateWebSettingsOOBOTPHTML, data)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	if r.Method == "POST" && r.Form.Get("x_action") == "remove" {
		err := h.Database.WithTx(func() error {
			intent := &webapp.Intent{
				RedirectURI: redirectURI,
				Intent:      intents.NewIntentRemoveAuthenticator(*userID),
			}
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = &InputRemoveAuthenticator{
					Type: authn.AuthenticatorTypeOOB,
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
		if err != nil {
			panic(err)
		}
	}

	if r.Method == "POST" && r.Form.Get("x_action") == "add" {
		err := h.Database.WithTx(func() error {
			intent := &webapp.Intent{
				RedirectURI: redirectURI,
				Intent: intents.NewIntentAddAuthenticator(
					*userID,
					interaction.AuthenticationStageSecondary,
					authn.AuthenticatorTypeOOB,
				),
			}
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = &InputCreateAuthenticator{}
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
