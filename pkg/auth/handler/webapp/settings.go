package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsHTML = template.RegisterHTML(
	"web/settings.html",
	components...,
)

func ConfigureSettingsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET", "POST").
		WithPathPattern("/settings")
}

type SettingsAuthenticatorService interface {
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
}

type SettingsIdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
	ListCandidates(userID string) ([]identity.Candidate, error)
}

type SettingsVerificationService interface {
	GetVerificationStatuses(is []*identity.Info) (map[string][]verification.ClaimStatus, error)
}

type SettingsSessionManager interface {
	List(userID string) ([]session.Session, error)
	Get(id string) (session.Session, error)
	Revoke(s session.Session) error
}

type SettingsHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	Authentication    *config.AuthenticationConfig
	Identities        SettingsIdentityService
	Verification      SettingsVerificationService
	Authenticators    SettingsAuthenticatorService
	MFA               SettingsMFAService
	CSRFCookie        webapp.CSRFCookieDef
}

func (h *SettingsHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	userID := session.GetUserID(r.Context())

	data := map[string]interface{}{}

	// BaseViewModel
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	// MFA
	authenticators, err := h.Authenticators.List(*userID)
	if err != nil {
		return nil, err
	}
	totp := false
	oobotp := false
	password := false
	for _, typ := range h.Authentication.SecondaryAuthenticators {
		switch typ {
		case authn.AuthenticatorTypePassword:
			password = true
		case authn.AuthenticatorTypeTOTP:
			totp = true
		case authn.AuthenticatorTypeOOB:
			oobotp = true
		}
	}
	mfaViewModel := SettingsMFAViewModel{
		Authenticators:           authenticators,
		SecondaryTOTPAllowed:     totp,
		SecondaryOOBOTPAllowed:   oobotp,
		SecondaryPasswordAllowed: password,
	}
	viewmodels.Embed(data, mfaViewModel)

	// Identity - Part 1
	candidates, err := h.Identities.ListCandidates(*userID)
	if err != nil {
		return nil, err
	}
	authenticationViewModel := viewmodels.NewAuthenticationViewModelWithCandidates(candidates)
	viewmodels.Embed(data, authenticationViewModel)

	// Identity - Part 2
	identities, err := h.Identities.ListByUser(*userID)
	if err != nil {
		return nil, err
	}
	identityViewModel := SettingsIdentityViewModel{}
	identityViewModel.VerificationStatuses, err = h.Verification.GetVerificationStatuses(identities)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, identityViewModel)

	return data, nil
}

func (h *SettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	redirectURI := httputil.HostRelative(r.URL).String()
	identityID := r.Form.Get("x_identity_id")
	userID := ctrl.RequireUserID()

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsHTML, data)
		return nil
	})

	ctrl.PostAction("unlink_oauth", func() error {
		opts := webapp.SessionOptions{
			RedirectURI: redirectURI,
		}
		intent := intents.NewIntentRemoveIdentity(userID)

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			input = &InputRemoveIdentity{
				Type: authn.IdentityTypeOAuth,
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
		intent := intents.NewIntentVerifyIdentity(userID, authn.IdentityTypeLoginID, identityID)

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
