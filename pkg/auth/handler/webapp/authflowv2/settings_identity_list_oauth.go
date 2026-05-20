package authflowv2

import (
	"context"
	"fmt"
	"net/http"

	"net/url"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/settingsaction"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsIdentityListOAuthHTML = template.RegisterHTML(
	"web/authflowv2/settings_identity_list_oauth.html",
	handlerwebapp.SettingsComponents...,
)

func ConfigureAuthflowV2SettingsIdentityListOAuthRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityListOAuth)
}

type AuthflowV2SettingsIdentityListOAuthViewModel struct {
	OAuthCandidates    []identity.Candidate
	OAuthIdentities    []*identity.OAuth
	Verifications      map[string][]verification.ClaimStatus
	IdentityCount      int
	CreateDisabled     bool
	IsInSettingsAction bool
	IsAlreadyLinked    bool
}

type AuthflowV2SettingsIdentityListOAuthHandler struct {
	AppID             config.AppID
	Database          *appdb.Handle
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          handlerwebapp.Renderer
	Identities        SettingsIdentityService
	Verification      SettingsVerificationService
	Endpoints         SettingsEndpointsProvider
	OAuthStateStore   SettingsOAuthStateStore
	AccountManagement accountmanagement.Service
}

func (h *AuthflowV2SettingsIdentityListOAuthHandler) getViewModel(ctx context.Context) (*AuthflowV2SettingsIdentityListOAuthViewModel, error) {
	userID := session.GetUserID(ctx)

	candidates, err := h.Identities.ListCandidates(ctx, *userID)
	if err != nil {
		return nil, err
	}

	var oauthCandidates []identity.Candidate
	for _, candidate := range candidates {
		typ, _ := candidate[identity.CandidateKeyType].(string)
		if typ == string(model.IdentityTypeOAuth) {
			oauthCandidates = append(oauthCandidates, candidate)
		}
	}

	identities, err := h.Identities.ListByUser(ctx, *userID)
	if err != nil {
		return nil, err
	}

	remaining := identity.ApplyFilters(
		identities,
		identity.KeepIdentifiable,
	)

	var oauthIdentities []*identity.OAuth
	var oauthInfos []*identity.Info
	for _, identity := range identities {
		if identity.Type != model.IdentityTypeOAuth {
			continue
		}

		oauthIdentities = append(oauthIdentities, identity.OAuth)
		oauthInfos = append(oauthInfos, identity.OAuth.ToInfo())
	}

	sort.Slice(oauthIdentities, func(i, j int) bool {
		return oauthIdentities[i].UpdatedAt.Before(oauthIdentities[j].UpdatedAt)
	})

	verifications, err := h.Verification.GetVerificationStatuses(ctx, oauthInfos)
	if err != nil {
		return nil, err
	}

	return &AuthflowV2SettingsIdentityListOAuthViewModel{
		OAuthCandidates: oauthCandidates,
		OAuthIdentities: oauthIdentities,
		Verifications:   verifications,
		IdentityCount:   len(remaining),
	}, nil
}

func (h *AuthflowV2SettingsIdentityListOAuthHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter) (map[string]any, error) {
	data := map[string]any{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	vm, err := h.getViewModel(ctx)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func generateAuthorizationURLWithState(authorizationURLString string, stateToken string) (string, error) {
	authorizationURL, err := url.Parse(authorizationURLString)
	if err != nil {
		return "", err
	}

	q := authorizationURL.Query()
	q.Set("state", stateToken)

	authorizationURL.RawQuery = q.Encode()

	return authorizationURL.String(), nil
}

// startOAuthFlow resolves the SSO callback URL for the given candidate,
// starts the account-management OAuth flow, and issues a redirect to the
// external provider. It is shared by the auto-trigger (Branch A) and the
// user-initiated POST "add" action so that both paths stay in sync.
func (h *AuthflowV2SettingsIdentityListOAuthHandler) startOAuthFlow(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	s session.ResolvedSession,
	alias string,
	candidate identity.Candidate,
) error {
	var redirectURI string
	if status, ok := candidate[identity.CandidateKeyProviderStatus].(string); ok && status == string(config.OAuthProviderStatusUsingDemoCredentials) {
		redirectURI = h.Endpoints.SharedSSOCallbackURL().String()
	} else {
		redirectURI = h.Endpoints.SSOCallbackURL(alias).String()
	}

	output, err := h.AccountManagement.StartAddIdentityOAuth(ctx, s, &accountmanagement.StartAddIdentityOAuthInput{
		Alias:       alias,
		RedirectURI: redirectURI,
	})
	if err != nil {
		return err
	}

	state := &webappoauth.WebappOAuthState{
		AppID:                  string(h.AppID),
		AccountManagementToken: output.Token,
		ProviderAlias:          alias,
		SettingsActionID:       settingsaction.GetSettingsActionID(r),
	}
	stateToken, err := h.OAuthStateStore.GenerateState(ctx, state)
	if err != nil {
		return err
	}

	authorizationURLString, err := generateAuthorizationURLWithState(output.AuthorizationURL, stateToken)
	if err != nil {
		return err
	}

	http.Redirect(w, r, authorizationURLString, http.StatusFound)
	return nil
}

// findCandidate returns the OAuth candidate matching alias, or nil.
func findCandidate(candidates []identity.Candidate, alias string) identity.Candidate {
	for _, c := range candidates {
		if c[identity.CandidateKeyProviderAlias] == alias {
			return c
		}
	}
	return nil
}

// autoTriggerOAuth looks up the candidate for providerAlias and, if not yet
// linked, starts the OAuth flow and returns handled=true. If the provider is
// already linked it returns handled=false so Branch C can show the error banner.
func (h *AuthflowV2SettingsIdentityListOAuthHandler) autoTriggerOAuth(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	s session.ResolvedSession,
	providerAlias string,
) (handled bool, err error) {
	var vm *AuthflowV2SettingsIdentityListOAuthViewModel
	err = h.Database.WithTx(ctx, func(ctx context.Context) error {
		var e error
		vm, e = h.getViewModel(ctx)
		return e
	})
	if err != nil {
		return false, err
	}

	candidate := findCandidate(vm.OAuthCandidates, providerAlias)
	if candidate == nil {
		return false, fmt.Errorf("unknown provider alias: %s", providerAlias)
	}

	if identityID, _ := candidate[identity.CandidateKeyIdentityID].(string); identityID == "" {
		return true, h.startOAuthFlow(ctx, w, r, s, providerAlias, candidate)
	}
	return false, nil
}

func (h *AuthflowV2SettingsIdentityListOAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx(r.Context())

	ctrl.GetWithSettingsActionWebSession(r, func(ctx context.Context, webappSession *webapp.Session) error {
		s := session.GetSession(ctx)
		providerAlias := r.URL.Query().Get("x_provider_alias")
		oauthConnected := r.URL.Query().Get("x_oauth_linked")

		// Branch A: in settings action, auto-trigger link when not yet linked.
		// Falls through to Branch C when the provider is already linked.
		// q_sso_error is set by the SSO callback on error to prevent an infinite
		// redirect loop: without it, every page load would trigger a new OAuth redirect.
		ssoError := r.URL.Query().Get("q_sso_error")
		if ctrl.IsInSettingsAction(s, webappSession) && providerAlias != "" && oauthConnected == "" && ssoError == "" {
			if handled, err := h.autoTriggerOAuth(ctx, w, r, s, providerAlias); err != nil || handled {
				return err
			}
			// Already linked: fall through to Branch C to show filtered list with error.
		}

		// Branch B: in settings action, finish after link
		if ctrl.IsInSettingsAction(s, webappSession) && oauthConnected == "1" {
			settingsActionResult, err := ctrl.FinishSettingsActionWithResult(ctx, s, webappSession)
			if err != nil {
				return err
			}
			settingsActionResult.WriteResponse(w, r)
			return nil
		}

		// Branch C: render list (filtered to single provider in settings-action mode)
		var vm *AuthflowV2SettingsIdentityListOAuthViewModel
		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			var e error
			vm, e = h.getViewModel(ctx)
			return e
		})
		if err != nil {
			return err
		}

		if ctrl.IsInSettingsAction(s, webappSession) && providerAlias != "" {
			filtered := []identity.Candidate{}
			if c := findCandidate(vm.OAuthCandidates, providerAlias); c != nil {
				filtered = append(filtered, c)
				identityID, _ := c[identity.CandidateKeyIdentityID].(string)
				vm.IsAlreadyLinked = identityID != ""
			}
			vm.OAuthCandidates = filtered
			vm.IsInSettingsAction = true
		}

		data := map[string]any{}
		viewmodels.Embed(data, h.BaseViewModel.ViewModel(r, w))
		viewmodels.Embed(data, vm)
		h.Renderer.RenderHTML(w, r, TemplateWebSettingsIdentityListOAuthHTML, data)
		return nil
	})

	ctrl.PostAction("add", func(ctx context.Context) error {
		s := session.GetSession(ctx)

		var vm *AuthflowV2SettingsIdentityListOAuthViewModel
		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			var e error
			vm, e = h.getViewModel(ctx)
			return e
		})
		if err != nil {
			return err
		}

		alias := r.Form.Get("x_provider_alias")
		candidate := findCandidate(vm.OAuthCandidates, alias)
		if candidate == nil {
			return fmt.Errorf("unknown provider alias: %s", alias)
		}

		return h.startOAuthFlow(ctx, w, r, s, alias, candidate)
	})

	ctrl.PostAction("remove", func(ctx context.Context) error {
		s := session.GetSession(ctx)

		identityID := r.Form.Get("q_identity_id")

		_, err := h.AccountManagement.DeleteIdentityOAuth(ctx, s, &accountmanagement.DeleteIdentityOAuthInput{
			IdentityID: identityID,
		})
		if err != nil {
			return err
		}

		redirectURI := httputil.HostRelative(r.URL).String()
		result := webapp.Result{RedirectURI: redirectURI}
		result.WriteResponse(w, r)

		return nil
	})
}
