package authflowv2

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/setutil"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSelectAccountHTML = template.RegisterHTML(
	"web/authflowv2/select_account.html",
	handlerwebapp.Components...,
)

func ConfigureAuthflowV2SelectAccountRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSelectAccount)
}

type SelectAccountUserService interface {
	Get(ctx context.Context, userID string, role accesscontrol.Role) (*model.User, error)
}

type SelectAccountUserFacade interface {
	GetUserIDsByLoginIDLoginHint(ctx context.Context, hint *oauth.LoginHint) ([]string, error)
}

type SelectAccountIdentityService interface {
	ListByUser(ctx context.Context, userID string) ([]*identity.Info, error)
}

type SelectAccountAuthenticationInfoService interface {
	Save(ctx context.Context, entry *authenticationinfo.Entry) error
}

type SelectAccountUIInfoResolver interface {
	SetAuthenticationInfoInQuery(redirectURI string, e *authenticationinfo.Entry) string
}

type SelectAccountViewModel struct {
	IdentityDisplayName string
	UserProfile         handlerwebapp.UserProfile
}

type AuthflowV2SelectAccountHandler struct {
	NonAuthflowControllerFactory handlerwebapp.ControllerFactory
	Controller                   *handlerwebapp.AuthflowController
	BaseViewModel                *viewmodels.BaseViewModeler
	Renderer                     handlerwebapp.Renderer
	AuthenticationConfig         *config.AuthenticationConfig
	SignedUpCookie               webapp.SignedUpCookieDef
	Users                        SelectAccountUserService
	UserFacade                   SelectAccountUserFacade
	Identities                   SelectAccountIdentityService
	AuthenticationInfoService    SelectAccountAuthenticationInfoService
	UIInfoResolver               SelectAccountUIInfoResolver
	Cookies                      handlerwebapp.CookieManager
	OAuthConfig                  *config.OAuthConfig
	UIConfig                     *config.UIConfig
	OAuthClientResolver          handlerwebapp.WebappOAuthClientResolver
}

func (h *AuthflowV2SelectAccountHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter, userID string) (map[string]any, error) {
	data := make(map[string]any)
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	identities, err := h.Identities.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	user, err := h.Users.Get(ctx, userID, accesscontrol.RoleGreatest)
	if err != nil {
		return nil, err
	}

	userProfile := handlerwebapp.GetUserProfile(user)
	displayID := handlerwebapp.IdentitiesDisplayName(identities)

	selectAccountViewModel := SelectAccountViewModel{
		IdentityDisplayName: displayID,
		UserProfile:         userProfile,
	}
	viewmodels.Embed(data, selectAccountViewModel)

	return data, nil
}

// nolint: gocognit
func (h *AuthflowV2SelectAccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.NonAuthflowControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	session := session.GetSession(r.Context())

	oauthSessionID := ""
	samlSessionID := ""
	loginPrompt := false
	userIDHint := ""
	canUseIntentReauthenticate := false
	suppressIDPSessionCookie := false
	oauthProviderAlias := ""
	var loginHint *oauth.LoginHint

	var webSession *webapp.Session
	ctrl.BeforeHandle(func(ctx context.Context) error {

		// Ensure webapp session exist
		ws, err := ctrl.InteractionSession(ctx)
		if err != nil {
			return err
		}
		webSession = ws

		oauthSessionID = webSession.OAuthSessionID
		samlSessionID = webSession.SAMLSessionID
		loginPrompt = slice.ContainsString(webSession.Prompt, "login")
		userIDHint = webSession.UserIDHint
		canUseIntentReauthenticate = webSession.CanUseIntentReauthenticate
		suppressIDPSessionCookie = webSession.SuppressIDPSessionCookie
		oauthProviderAlias = webSession.OAuthProviderAlias

		// When x_suppress_idp_session_cookie is true, ignore IDP session cookie.
		if suppressIDPSessionCookie {
			session = nil
		}
		// Ignore any session that is not allow to be used here
		if !oauth.ContainsAllScopes(oauth.SessionScopes(session), []string{oauth.PreAuthenticatedURLScope}) {
			session = nil
		}

		// Ignore any session that does not match login_hint
		if webSession.LoginHint != "" {
			l, err := oauth.ParseLoginHint(webSession.LoginHint)
			// Ignore the login_hint if it is not something we understand
			if err == nil {
				loginHint = l
			}
		}

		if loginHint != nil && session != nil && loginHint.Type == oauth.LoginHintTypeLoginID {
			hintUserIDs, err := h.UserFacade.GetUserIDsByLoginIDLoginHint(ctx, loginHint)
			if err != nil {
				return err
			}
			hintUserIDsSet := setutil.NewSetFromSlice(hintUserIDs, setutil.Identity[string])
			if !hintUserIDsSet.Has(session.GetAuthenticationInfo().UserID) {
				session = nil
			}
		}
		return nil
	})

	// continueWithCurrentAccountLegacy mints a session/authentication-info
	// entry directly, bypassing the authentication flow engine entirely.
	// Only called by the userIDHint continue-without-interaction branch
	// below: user_id_hint targets a specific already-known user, so there is
	// nothing to "select", and this predates the account-chooser UI this
	// file's other refactored branches now use. If a resolved login flow
	// does not declare a select_account entry, that project simply does not
	// support account continuation — the identify/continue branches below
	// redirect to signup/login in that case, they do not fall back here.
	continueWithCurrentAccountLegacy := func(ctx context.Context) error {
		redirectURI := ""

		// Complete the web session and redirect to web session's RedirectURI
		if webSession != nil {
			redirectURI = webSession.RedirectURI
			if err := ctrl.DeleteSession(ctx, webSession.ID); err != nil {
				return err
			}
		}

		if redirectURI == "" {
			redirectURI = webapp.DerivePostLoginRedirectURIFromRequest(r, h.OAuthClientResolver, h.UIConfig)
		}

		// Write authentication info cookie
		if session != nil {
			info := session.CreateNewAuthenticationInfoByThisSession()
			info.ShouldFireAuthenticatedEventWhenIssueOfflineGrant = true
			info.ContinueFromSessionType = string(session.SessionType())
			info.ContinueFromSessionID = session.SessionID()
			entry := authenticationinfo.NewEntry(info, oauthSessionID, samlSessionID)
			err := h.AuthenticationInfoService.Save(ctx, entry)
			if err != nil {
				return err
			}
			redirectURI = h.UIInfoResolver.SetAuthenticationInfoInQuery(redirectURI, entry)
		}

		// #nosec G710 -- redirectURI is either webSession.RedirectURI (set from an allow-listed/same-origin redirect_uri when the web session was created) or webapp.DerivePostLoginRedirectURIFromRequest, which allow-lists against the OAuth client's registered RedirectURIs.
		http.Redirect(w, r, redirectURI, http.StatusFound)
		return nil
	}
	gotoSignupOrLogin := func() {
		// Page has the highest precedence if it is specified.
		if webSession != nil && webSession.Page != "" {
			var path string
			switch webSession.Page {
			case "signup":
				if h.AuthenticationConfig.PublicSignupDisabled {
					path = "/login"
				} else {
					path = "/signup"
				}
			case "login":
				path = "/login"
			}
			if path != "" {
				h.continueFlow(w, r, path)
				return
			}
		}

		// If a login id login_hint exist, go to login
		if loginHint != nil && loginHint.Type == oauth.LoginHintTypeLoginID {
			h.continueFlow(w, r, "/login")
			return
		}

		// Page is something that we do not understand or it is absent.
		// In this case, we look at the cookie.
		signedUpCookie, err := h.Cookies.GetCookie(r, h.SignedUpCookie.Def)
		signedUp := (err == nil && signedUpCookie.Value == "true")
		path := "/signup"
		if h.AuthenticationConfig.PublicSignupDisabled || signedUp {
			path = "/login"
		}

		h.continueFlow(w, r, path)
	}
	gotoLogin := func() {
		h.continueFlow(w, r, "/login")
	}
	gotoReauth := func() {
		h.continueFlow(w, r, "/reauth")
	}

	// ctrl.ServeWithDBTx() always write response.
	// So we have to put http.Redirect before it.
	defer ctrl.ServeWithDBTx(r.Context())

	ctrl.Get(func(ctx context.Context) error {
		// When promote anonymous user, the end-user should not see this page.
		if loginHint != nil && loginHint.Type == oauth.LoginHintTypeAnonymous {
			h.continueFlow(w, r, "/flows/promote_user")
			return nil
		}

		// When UserIDHint is present, the end-user should never need to select anything in /select_account,
		// so this if block always ends with a return statement, and each branch must write response.
		if userIDHint != "" {
			if loginPrompt && canUseIntentReauthenticate {
				gotoReauth()
			} else if !loginPrompt && session != nil && session.GetAuthenticationInfo().UserID == userIDHint {
				// Continue without user interaction
				// 1. UserIDHint present
				// 2. IDP session present and the same as UserIDHint
				// 3. prompt!=login

				err := continueWithCurrentAccountLegacy(ctx)
				if err != nil {
					return err
				}
			} else {
				gotoLogin()
			}

			return nil
		}

		// If anything of the following condition holds,
		// the end-user does not need to select anything.
		// - If x_oauth_provider_alisa is provided via authorization endpoint
		// - The request is not from the authorization endpoint, e.g. /
		if oauthProviderAlias != "" {
			gotoLogin()
			return nil
		}

		fromAuthzEndpoint := oauthSessionID != "" || samlSessionID != ""
		if !fromAuthzEndpoint {
			gotoSignupOrLogin()
			return nil
		}

		// session == nil / loginPrompt reflect webapp-specific reasons to
		// distrust the session (suppressed, out-of-scope, or a mismatched
		// login_hint — the latter is deliberately NOT checked by
		// NewIdentificationOptionsSelectAccount, see Part 1's spec) that the
		// resolved login flow does not know about, so they must gate here,
		// before a flow is even created — never deferred to the flow's own
		// answer.
		if session == nil || loginPrompt {
			gotoSignupOrLogin()
			return nil
		}

		// The resolved login flow is created here so the POST "continue"
		// action below can advance the same flow instance. Whether to
		// actually render "Continue as X" is read directly off that flow's
		// identify-step response — if it doesn't declare select_account,
		// this project simply does not support account continuation, same
		// as a Custom UI would see.
		var getHandlers handlerwebapp.AuthflowControllerHandlers
		getHandlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
			if _, _, ok := selectAccountOptionFromScreen(screen); !ok {
				gotoSignupOrLogin()
				return nil
			}

			data, err := h.GetData(ctx, r, w, session.GetAuthenticationInfo().UserID)
			if err != nil {
				return err
			}
			h.Renderer.RenderHTML(w, r, TemplateWebSelectAccountHTML, data)
			return nil
		})

		opts := webapp.SessionOptions{
			OAuthSessionID: oauthSessionID,
			SAMLSessionID:  samlSessionID,
		}
		h.Controller.HandleStartOfFlow(ctx, w, r.WithContext(ctx), opts, authflow.FlowTypeLogin, &getHandlers, nil)
		return nil
	})

	ctrl.PostAction("continue", func(ctx context.Context) error {
		var postHandlers handlerwebapp.AuthflowControllerHandlers
		postHandlers.PostAction("continue", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
			// The resolved login flow's identify step may not declare a
			// select_account entry at all (a customized, non-generated flow
			// that predates this feature, e.g. one whose identify step only
			// has `identification: username`) — that project simply does
			// not support account continuation, so redirect the same way
			// the GET branch does rather than falling back to legacy
			// completion. Feeding {"identification":"select_account"} to
			// such a flow would be rejected by the input's JSON schema
			// (built only from the options this flow actually declares)
			// BEFORE the flow engine's ReactTo ever runs, surfacing a
			// *validation.AggregatedError, NOT authflow.ErrIncompatibleInput
			// — check upfront instead of relying on error matching.
			//
			// index is this option's actual position in the flow's one_of
			// list, not assumed to be 0: a hand-authored flow may declare
			// select_account anywhere, unlike the generated default flow,
			// which always prepends it first.
			_, index, ok := selectAccountOptionFromScreen(screen)
			if !ok {
				gotoSignupOrLogin()
				return nil
			}

			result, err := h.Controller.AdvanceWithInput(ctx, r, s, screen, map[string]any{
				"identification": "select_account",
				"index":          index,
			}, nil)
			if err != nil {
				return err
			}
			result.WriteResponse(w, r)
			return nil
		})

		opts := webapp.SessionOptions{
			OAuthSessionID: oauthSessionID,
			SAMLSessionID:  samlSessionID,
		}
		h.Controller.HandleStartOfFlow(ctx, w, r.WithContext(ctx), opts, authflow.FlowTypeLogin, &postHandlers, nil)
		return nil
	})

	ctrl.PostAction("login", func(ctx context.Context) error {
		gotoSignupOrLogin()
		return nil
	})
}

func (h *AuthflowV2SelectAccountHandler) continueFlow(w http.ResponseWriter, r *http.Request, path string) {
	// preserve query only when continuing the login flow
	u := webapp.MakeRelativeURL(path, webapp.PreserveQuery(r.URL.Query()))
	// #nosec G710 -- webapp.MakeRelativeURL only ever sets Path and RawQuery, never Scheme/Host, so u is always relative to the current origin.
	http.Redirect(w, r, u.String(), http.StatusFound)
}

// selectAccountOptionFromScreen reports whether the login flow's current
// identify-step response offers a select_account option, and returns it
// (with its position in the options array, needed to submit a matching
// "index" — a hand-authored flow may declare select_account anywhere in its
// one_of list, not necessarily first, unlike the generated default flow)
// if so. Mirrors the read pattern already used by every other
// AuthflowController-backed screen (e.g. reset_password.go's
// declarative.NewPasswordData type assertion) — returns ok == false for any
// other action type, not just a missing select_account entry, since a
// customized flow's identify step may not even be the current action.
func selectAccountOptionFromScreen(screen *webapp.AuthflowScreenWithFlowResponse) (option declarative.IdentificationOption, index int, ok bool) {
	data, ok := screen.StateTokenFlowResponse.Action.Data.(declarative.IdentificationData)
	if !ok {
		return declarative.IdentificationOption{}, 0, false
	}
	for i, o := range data.Options {
		if o.Identification == model.AuthenticationFlowIdentificationSelectAccount {
			return o, i, true
		}
	}
	return declarative.IdentificationOption{}, 0, false
}
