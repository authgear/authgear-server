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
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
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

type SelectAccountViewModel struct {
	IdentityDisplayName string
	UserProfile         handlerwebapp.UserProfile
}

type AuthflowV2SelectAccountHandler struct {
	Controller           *handlerwebapp.AuthflowController
	BaseViewModel        *viewmodels.BaseViewModeler
	Renderer             handlerwebapp.Renderer
	AuthenticationConfig *config.AuthenticationConfig
	SignedUpCookie       webapp.SignedUpCookieDef
	Users                SelectAccountUserService
	UserFacade           SelectAccountUserFacade
	Cookies              handlerwebapp.CookieManager
	OAuthConfig          *config.OAuthConfig
	Database             *appdb.Handle
}

// GetData builds the render data for the select_account page using only the
// select_account option already read off the resolved login flow's
// identify-step response — display_name and user_id are exposed by the
// authentication-flow API precisely so this default UI can be implemented
// the same way a Custom UI would be, with no separate session lookup of its
// own. h.Users.Get is the one remaining direct dependency, since the API
// does not expose full profile data (avatar etc.), only display_name.
func (h *AuthflowV2SelectAccountHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter, option declarative.IdentificationOption) (map[string]any, error) {
	data := make(map[string]any)
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	var userProfile handlerwebapp.UserProfile
	err := h.Database.WithTx(ctx, func(ctx context.Context) error {
		user, err := h.Users.Get(ctx, option.UserID, accesscontrol.RoleGreatest)
		if err != nil {
			return err
		}
		userProfile = handlerwebapp.GetUserProfile(user)
		return nil
	})
	if err != nil {
		return nil, err
	}

	selectAccountViewModel := SelectAccountViewModel{
		IdentityDisplayName: option.DisplayName,
		UserProfile:         userProfile,
	}
	viewmodels.Embed(data, selectAccountViewModel)

	return data, nil
}

func (h *AuthflowV2SelectAccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gotoSignupOrLogin := func(s *webapp.Session) {
		// Page has the highest precedence if it is specified.
		if s.Page != "" {
			var path string
			switch s.Page {
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
		if loginHint, ok := parseLoginHint(s); ok && loginHint.Type == oauth.LoginHintTypeLoginID {
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

	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		return h.get(ctx, w, r, s, screen, gotoSignupOrLogin, gotoLogin, gotoReauth)
	})

	handlers.PostAction("continue", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		// The resolved login flow's identify step may not declare a
		// select_account entry at all (a customized, non-generated flow
		// that predates this feature, e.g. one whose identify step only
		// has `identification: username`) — that project simply does not
		// support account continuation, so redirect the same way the GET
		// branch does rather than falling back to legacy completion.
		// Feeding {"identification":"select_account"} to such a flow would
		// be rejected by the input's JSON schema (built only from the
		// options this flow actually declares) BEFORE the flow engine's
		// ReactTo ever runs, surfacing a *validation.AggregatedError, NOT
		// authflow.ErrIncompatibleInput — check upfront instead of relying
		// on error matching.
		//
		// index is this option's actual position in the flow's one_of
		// list, not assumed to be 0: a hand-authored flow may declare
		// select_account anywhere, unlike the generated default flow,
		// which always prepends it first.
		_, index, ok := selectAccountOptionFromScreen(screen)
		if !ok {
			gotoSignupOrLogin(s)
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

	handlers.PostAction("login", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		gotoSignupOrLogin(s)
		return nil
	})

	opts := webapp.SessionOptions{
		RedirectURI: h.Controller.RedirectURI(r),
	}
	h.Controller.HandleStartOfFlow(r.Context(), w, r, opts, authflow.FlowTypeLogin, &handlers, nil)
}

func (h *AuthflowV2SelectAccountHandler) get(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	s *webapp.Session,
	screen *webapp.AuthflowScreenWithFlowResponse,
	gotoSignupOrLogin func(s *webapp.Session),
	gotoLogin func(),
	gotoReauth func(),
) error {
	loginHint, hasLoginHint := parseLoginHint(s)

	// When promote anonymous user, the end-user should not see this page.
	if hasLoginHint && loginHint.Type == oauth.LoginHintTypeAnonymous {
		h.continueFlow(w, r, "/flows/promote_user")
		return nil
	}

	loginPrompt := slice.ContainsString(s.Prompt, "login")

	// When UserIDHint is present, the end-user should never need to select anything in /select_account,
	// so this if block always ends with a return statement, and each branch must write response.
	if s.UserIDHint != "" {
		if loginPrompt && s.CanUseIntentReauthenticate {
			gotoReauth()
		} else if !loginPrompt {
			// Continue without user interaction:
			// 1. UserIDHint present
			// 2. The resolved login flow offers select_account for exactly
			//    this user (which already implies a usable, matching IDP
			//    session — see NewIdentificationOptionsSelectAccount)
			// 3. prompt!=login
			//
			// Resolved the same way as a normal "Continue as X" click —
			// submit select_account against the already-created login
			// flow, without ever rendering a page. If the flow doesn't
			// declare select_account, this project simply doesn't
			// support account continuation, so fall through to a
			// normal login instead of any engine-bypassing fallback.
			option, index, ok, err := h.resolveSelectAccountOption(ctx, screen, loginHint, hasLoginHint)
			if err != nil {
				return err
			}
			if !ok || option.UserID != s.UserIDHint {
				gotoLogin()
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
		} else {
			gotoLogin()
		}

		return nil
	}

	// If anything of the following condition holds,
	// the end-user does not need to select anything.
	// - If x_oauth_provider_alisa is provided via authorization endpoint
	// - The request is not from the authorization endpoint, e.g. /
	if s.OAuthProviderAlias != "" {
		gotoLogin()
		return nil
	}

	fromAuthzEndpoint := s.OAuthSessionID != "" || s.SAMLSessionID != ""
	if !fromAuthzEndpoint {
		gotoSignupOrLogin(s)
		return nil
	}

	// Whether to actually render "Continue as X" is read directly off the
	// already-created login flow's identify-step response, with login_hint
	// enforcement applied on top (the flow deliberately does not check
	// login_hint itself, see NewIdentificationOptionsSelectAccount) — if
	// select_account isn't offered, or the offered account doesn't match
	// login_hint, this project simply does not support account
	// continuation here, same as a Custom UI would see.
	option, _, ok, err := h.resolveSelectAccountOption(ctx, screen, loginHint, hasLoginHint)
	if err != nil {
		return err
	}
	if !ok {
		gotoSignupOrLogin(s)
		return nil
	}

	data, err := h.GetData(ctx, r, w, option)
	if err != nil {
		return err
	}
	h.Renderer.RenderHTML(w, r, TemplateWebSelectAccountHTML, data)
	return nil
}

// resolveSelectAccountOption reads the select_account option off the
// resolved login flow's identify-step response — the same data a Custom UI
// would see — and, if present, additionally enforces login_hint matching:
// the one piece of eligibility NewIdentificationOptionsSelectAccount
// deliberately does not implement itself. This is the only remaining
// select_account-specific policy layered on top of the flow's own answer;
// everything else (session validity, suppression, scope, prompt=login) is
// read straight from whether the flow offered the option at all, using only
// the option's exposed UserID — never this handler's own session lookup.
func (h *AuthflowV2SelectAccountHandler) resolveSelectAccountOption(
	ctx context.Context,
	screen *webapp.AuthflowScreenWithFlowResponse,
	loginHint *oauth.LoginHint,
	hasLoginHint bool,
) (option declarative.IdentificationOption, index int, ok bool, err error) {
	option, index, ok = selectAccountOptionFromScreen(screen)
	if !ok {
		return declarative.IdentificationOption{}, 0, false, nil
	}

	if hasLoginHint && loginHint.Type == oauth.LoginHintTypeLoginID {
		var hintUserIDs []string
		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			var err error
			hintUserIDs, err = h.UserFacade.GetUserIDsByLoginIDLoginHint(ctx, loginHint)
			return err
		})
		if err != nil {
			return declarative.IdentificationOption{}, 0, false, err
		}
		hintUserIDsSet := setutil.NewSetFromSlice(hintUserIDs, setutil.Identity[string])
		if !hintUserIDsSet.Has(option.UserID) {
			return declarative.IdentificationOption{}, 0, false, nil
		}
	}

	return option, index, true, nil
}

// parseLoginHint parses s.LoginHint, reporting ok == false if it is absent
// or not something we understand (in which case it should be ignored, not
// treated as an error).
func parseLoginHint(s *webapp.Session) (hint *oauth.LoginHint, ok bool) {
	if s.LoginHint == "" {
		return nil, false
	}
	l, err := oauth.ParseLoginHint(s.LoginHint)
	if err != nil {
		return nil, false
	}
	return l, true
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
