package authflowv2

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/httproute"
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
	Get(userID string, role accesscontrol.Role) (*model.User, error)
}

type SelectAccountIdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type SelectAccountAuthenticationInfoService interface {
	Save(entry *authenticationinfo.Entry) error
}

type SelectAccountUIInfoResolver interface {
	SetAuthenticationInfoInQuery(redirectURI string, e *authenticationinfo.Entry) string
}

type SelectAccountViewModel struct {
	IdentityDisplayName string
	UserProfile         handlerwebapp.UserProfile
}

type AuthflowV2SelectAccountHandler struct {
	ControllerFactory         handlerwebapp.ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	Renderer                  handlerwebapp.Renderer
	AuthenticationConfig      *config.AuthenticationConfig
	SignedUpCookie            webapp.SignedUpCookieDef
	Users                     SelectAccountUserService
	Identities                SelectAccountIdentityService
	AuthenticationInfoService SelectAccountAuthenticationInfoService
	UIInfoResolver            SelectAccountUIInfoResolver
	Cookies                   handlerwebapp.CookieManager
	OAuthConfig               *config.OAuthConfig
	UIConfig                  *config.UIConfig
	OAuthClientResolver       handlerwebapp.WebappOAuthClientResolver
}

func (h *AuthflowV2SelectAccountHandler) GetData(r *http.Request, rw http.ResponseWriter, userID string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	identities, err := h.Identities.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	user, err := h.Users.Get(userID, accesscontrol.RoleGreatest)
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
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	session := session.GetSession(r.Context())
	webSession := webapp.GetSession(r.Context())

	oauthSessionID := ""
	loginPrompt := false
	userIDHint := ""
	canUseIntentReauthenticate := false
	suppressIDPSessionCookie := false
	oauthProviderAlias := ""

	if webSession != nil {
		oauthSessionID = webSession.OAuthSessionID
		loginPrompt = slice.ContainsString(webSession.Prompt, "login")
		userIDHint = webSession.UserIDHint
		canUseIntentReauthenticate = webSession.CanUseIntentReauthenticate
		suppressIDPSessionCookie = webSession.SuppressIDPSessionCookie
		oauthProviderAlias = webSession.OAuthProviderAlias
	}

	// When x_suppress_idp_session_cookie is true, ignore IDP session cookie.
	if suppressIDPSessionCookie {
		session = nil
	}
	// Ignore any session that is not allow to be used here
	if !oauth.ContainsAllScopes(oauth.SessionScopes(session), []string{oauth.PreAuthenticatedURLScope}) {
		session = nil
	}

	continueWithCurrentAccount := func() error {
		redirectURI := ""

		// Complete the web session and redirect to web session's RedirectURI
		if webSession != nil {
			redirectURI = webSession.RedirectURI
			if err := ctrl.DeleteSession(webSession.ID); err != nil {
				return err
			}
		}

		if redirectURI == "" {
			redirectURI = webapp.DerivePostLoginRedirectURIFromRequest(r, h.OAuthClientResolver, h.UIConfig)
		}

		// Write authentication info cookie
		if session != nil {
			info := session.CreateNewAuthenticationInfoByThisSession()
			entry := authenticationinfo.NewEntry(info, oauthSessionID)
			err := h.AuthenticationInfoService.Save(entry)
			if err != nil {
				return err
			}
			redirectURI = h.UIInfoResolver.SetAuthenticationInfoInQuery(redirectURI, entry)
		}

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

	// ctrl.Serve() always write response.
	// So we have to put http.Redirect before it.
	defer ctrl.Serve()

	ctrl.Get(func() error {
		// When promote anonymous user, the end-user should not see this page.
		if webSession != nil && webSession.LoginHint != "" {
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

				err := continueWithCurrentAccount()
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
		// - There is no session, so nothing to select.
		// - prompt=login, in this case, the end-user cannot select existing account.
		if oauthProviderAlias != "" {
			gotoLogin()
			return nil
		}

		fromAuthzEndpoint := oauthSessionID != ""
		if !fromAuthzEndpoint || session == nil || loginPrompt {
			gotoSignupOrLogin()
			return nil
		}

		data, err := h.GetData(r, w, session.GetAuthenticationInfo().UserID)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSelectAccountHTML, data)
		return nil
	})

	ctrl.PostAction("continue", func() error {
		return continueWithCurrentAccount()
	})

	ctrl.PostAction("login", func() error {
		gotoSignupOrLogin()
		return nil
	})
}

func (h *AuthflowV2SelectAccountHandler) continueFlow(w http.ResponseWriter, r *http.Request, path string) {
	// preserve query only when continuing the login flow
	u := webapp.MakeRelativeURL(path, webapp.PreserveQuery(r.URL.Query()))
	http.Redirect(w, r, u.String(), http.StatusFound)
}
