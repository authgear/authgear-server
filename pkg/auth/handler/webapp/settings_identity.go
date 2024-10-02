package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsIdentityHTML = template.RegisterHTML(
	"web/settings_identity.html",
	Components...,
)

func ConfigureSettingsIdentityRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/identity")
}

type SettingsIdentityViewModel struct {
	VerificationStatuses   map[string][]verification.ClaimStatus
	AccountDeletionAllowed bool
}

type SettingsIdentityHandler struct {
	ControllerFactory       ControllerFactory
	BaseViewModel           *viewmodels.BaseViewModeler
	AuthenticationViewModel *viewmodels.AuthenticationViewModeler
	Renderer                Renderer
	Identities              SettingsIdentityService
	Verification            SettingsVerificationService
	AccountDeletion         *config.AccountDeletionConfig
}

func (h *SettingsIdentityHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	userID := session.GetUserID(r.Context())

	candidates, err := h.Identities.ListCandidates(*userID)
	if err != nil {
		return nil, err
	}
	authenticationViewModel := h.AuthenticationViewModel.NewWithCandidates(candidates, r.Form)

	viewModel := SettingsIdentityViewModel{
		AccountDeletionAllowed: h.AccountDeletion.ScheduledByEndUserEnabled,
	}
	identities, err := h.Identities.ListByUser(*userID)
	if err != nil {
		return nil, err
	}
	viewModel.VerificationStatuses, err = h.Verification.GetVerificationStatuses(identities)
	if err != nil {
		return nil, err
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, authenticationViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

func (h *SettingsIdentityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

	redirectURI := httputil.HostRelative(r.URL).String()
	providerAlias := r.Form.Get("x_provider_alias")
	identityID := r.Form.Get("q_identity_id")
	userID := ctrl.RequireUserID()

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsIdentityHTML, data)
		return nil
	})

	ctrl.PostAction("link_oauth", func() error {
		opts := webapp.SessionOptions{
			RedirectURI: redirectURI,
		}
		intent := intents.NewIntentAddIdentity(userID)

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			// FIXME(settings): support prompt parameters for connecting oauth
			input = &InputUseOAuth{
				ProviderAlias:    providerAlias,
				ErrorRedirectURI: httputil.HostRelative(r.URL).String(),
			}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("unlink_oauth", func() error {
		opts := webapp.SessionOptions{
			RedirectURI: redirectURI,
		}
		intent := intents.NewIntentRemoveIdentity(userID)

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			input = &InputRemoveIdentity{
				Type: model.IdentityTypeOAuth,
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

	ctrl.PostAction("verify_login_id", func() error {
		opts := webapp.SessionOptions{
			RedirectURI:     redirectURI,
			KeepAfterFinish: true,
		}
		intent := intents.NewIntentVerifyIdentity(userID, model.IdentityTypeLoginID, identityID)

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			input = nil
			return
		})
		if err != nil {
			return err
		}
		result.WriteResponse(w, r)
		return nil
	})
}
