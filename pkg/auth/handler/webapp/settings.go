package webapp

import (
	"context"
	"errors"
	"net/http"

	"net/url"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/sessionlisting"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsHTML = template.RegisterHTML(
	"web/settings.html",
	Components...,
)

var TemplateWebSettingsAnonymousUserHTML = template.RegisterHTML(
	"web/settings_anonymous_user.html",
	Components...,
)

func ConfigureSettingsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET", "POST").
		WithPathPattern("/settings")
}

type SettingsEndpointsProvider interface {
	SSOCallbackURL(provider string) *url.URL
}

type SettingsOAuthStateStore interface {
	GenerateState(ctx context.Context, state *webappoauth.WebappOAuthState) (stateToken string, err error)
}

type SettingsAuthenticatorService interface {
	List(ctx context.Context, userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
}

type SettingsIdentityService interface {
	ListByUser(ctx context.Context, userID string) ([]*identity.Info, error)
	ListCandidates(ctx context.Context, userID string) ([]identity.Candidate, error)
}

type SettingsVerificationService interface {
	GetVerificationStatuses(ctx context.Context, is []*identity.Info) (map[string][]verification.ClaimStatus, error)
}

type SettingsSessionManager interface {
	List(ctx context.Context, userID string) ([]session.ListableSession, error)
	Get(ctx context.Context, id string) (session.ListableSession, error)
	RevokeWithEvent(ctx context.Context, s session.SessionBase, isTermination bool, isAdminAPI bool) error
	TerminateAllExcept(ctx context.Context, userID string, currentSession session.ResolvedSession, isAdminAPI bool) error
}

type SettingsAuthorizationService interface {
	GetByID(ctx context.Context, id string) (*oauth.Authorization, error)
	ListByUser(ctx context.Context, userID string, filters ...oauth.AuthorizationFilter) ([]*oauth.Authorization, error)
	Delete(ctx context.Context, a *oauth.Authorization) error
}

type SettingsSessionListingService interface {
	FilterForDisplay(ctx context.Context, sessions []session.ListableSession, currentSession session.ResolvedSession) ([]*sessionlisting.Session, error)
}

type SettingsHandler struct {
	ControllerFactory        ControllerFactory
	BaseViewModel            *viewmodels.BaseViewModeler
	AuthenticationViewModel  *viewmodels.AuthenticationViewModeler
	SettingsViewModel        *viewmodels.SettingsViewModeler
	SettingsProfileViewModel *viewmodels.SettingsProfileViewModeler
	Renderer                 Renderer
	Identities               SettingsIdentityService
	Verification             SettingsVerificationService
	AccountDeletion          *config.AccountDeletionConfig
	AccountAnonymization     *config.AccountAnonymizationConfig
	TutorialCookie           TutorialCookie
}

func (h *SettingsHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	userID := session.GetUserID(ctx)

	data := map[string]interface{}{}

	// BaseViewModel
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	if h.TutorialCookie.Pop(r, rw, httputil.SettingsTutorialCookieName) {
		baseViewModel.SetTutorial(httputil.SettingsTutorialCookieName)
	}
	viewmodels.Embed(data, baseViewModel)

	// SettingsViewModel
	viewModelPtr, err := h.SettingsViewModel.ViewModel(ctx, *userID)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, *viewModelPtr)

	// SettingsProfileViewModel
	profileViewModelPtr, err := h.SettingsProfileViewModel.ViewModel(ctx, *userID)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, *profileViewModelPtr)

	// Identity - Part 1
	candidates, err := h.Identities.ListCandidates(ctx, *userID)
	if err != nil {
		return nil, err
	}
	authenticationViewModel := h.AuthenticationViewModel.NewWithCandidates(candidates, r.Form)
	viewmodels.Embed(data, authenticationViewModel)

	// Identity - Part 2
	identities, err := h.Identities.ListByUser(ctx, *userID)
	if err != nil {
		return nil, err
	}
	identityViewModel := SettingsIdentityViewModel{
		AccountDeletionAllowed: h.AccountDeletion.ScheduledByEndUserEnabled,
	}
	identityViewModel.VerificationStatuses, err = h.Verification.GetVerificationStatuses(ctx, identities)
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
	defer ctrl.ServeWithDBTx(r.Context())

	redirectURI := httputil.HostRelative(r.URL).String()
	identityID := r.Form.Get("q_identity_id")
	userID := ctrl.RequireUserID(r.Context())

	// check if the user is anonymous user
	getIsAnonymous := func(ctx context.Context) (bool, error) {
		identities, err := h.Identities.ListByUser(ctx, userID)
		if err != nil {
			return false, err
		}
		for _, i := range identities {
			if i.Type == model.IdentityTypeAnonymous {
				return true, nil
			}
		}
		return false, nil
	}

	ctrl.Get(func(ctx context.Context) error {
		data, err := h.GetData(ctx, r, w)
		if err != nil {
			return err
		}
		isAnonymous, err := getIsAnonymous(ctx)
		if err != nil {
			return err
		}

		if isAnonymous {
			h.Renderer.RenderHTML(w, r, TemplateWebSettingsAnonymousUserHTML, data)
		} else {
			h.Renderer.RenderHTML(w, r, TemplateWebSettingsHTML, data)
		}
		return nil
	})

	ctrl.PostAction("unlink_oauth", func(ctx context.Context) error {
		isAnonymous, err := getIsAnonymous(ctx)
		if err != nil {
			return err
		}
		if isAnonymous {
			return errors.New("unexpected unlink oauth for anonymous user")
		}

		opts := webapp.SessionOptions{
			RedirectURI: redirectURI,
		}
		intent := intents.NewIntentRemoveIdentity(userID)

		result, err := ctrl.EntryPointPost(ctx, opts, intent, func() (input interface{}, err error) {
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

	ctrl.PostAction("verify_login_id", func(ctx context.Context) error {
		isAnonymous, err := getIsAnonymous(ctx)
		if err != nil {
			return err
		}
		if isAnonymous {
			return errors.New("unexpected verify login id for anonymous user")
		}

		opts := webapp.SessionOptions{
			RedirectURI:     redirectURI,
			KeepAfterFinish: true,
		}
		intent := intents.NewIntentVerifyIdentity(userID, model.IdentityTypeLoginID, identityID)

		result, err := ctrl.EntryPointPost(ctx, opts, intent, func() (input interface{}, err error) {
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
