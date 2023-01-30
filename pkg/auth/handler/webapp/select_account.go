package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSelectAccountHTML = template.RegisterHTML(
	"web/select_account.html",
	components...,
)

func ConfigureSelectAccountRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/select_account")
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

type SelectAccountViewModel struct {
	IdentityDisplayName string
}

type SelectAccountHandler struct {
	ControllerFactory         ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	Renderer                  Renderer
	AuthenticationConfig      *config.AuthenticationConfig
	SignedUpCookie            webapp.SignedUpCookieDef
	Users                     SelectAccountUserService
	Identities                SelectAccountIdentityService
	AuthenticationInfoService SelectAccountAuthenticationInfoService
	Cookies                   CookieManager
	OAuthConfig               *config.OAuthConfig
	UIConfig                  *config.UIConfig
}

func (h *SelectAccountHandler) GetData(r *http.Request, rw http.ResponseWriter, userID string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	identities, err := h.Identities.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	displayID := IdentitiesDisplayName(identities)

	selectAccountViewModel := SelectAccountViewModel{
		IdentityDisplayName: displayID,
	}
	viewmodels.Embed(data, selectAccountViewModel)

	return data, nil
}

// nolint: gocyclo
func (h *SelectAccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	idpSession := session.GetSession(r.Context())
	webSession := webapp.GetSession(r.Context())

	loginPrompt := false
	fromAuthzEndpoint := false
	userIDHint := ""
	canUseIntentReauthenticate := false
	suppressIDPSessionCookie := false
	oauthProviderAlias := ""

	if webSession != nil {
		loginPrompt = slice.ContainsString(webSession.Prompt, "login")
		fromAuthzEndpoint = webSession.FromAuthzEndpoint
		userIDHint = webSession.UserIDHint
		canUseIntentReauthenticate = webSession.CanUseIntentReauthenticate
		suppressIDPSessionCookie = webSession.SuppressIDPSessionCookie
		oauthProviderAlias = webSession.OAuthProviderAlias
	}

	// When x_suppress_idp_session_cookie is true, ignore IDP session cookie.
	if suppressIDPSessionCookie {
		idpSession = nil
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
			redirectURI = webapp.DerivePostLoginRedirectURIFromRequest(r, h.OAuthConfig, h.UIConfig)
		}

		// Write authentication info cookie
		if idpSession != nil {
			info := idpSession.GetAuthenticationInfo()
			entry := authenticationinfo.NewEntry(info)
			err := h.AuthenticationInfoService.Save(entry)
			if err != nil {
				return err
			}
			cookie := h.Cookies.ValueCookie(
				authenticationinfo.CookieDef,
				entry.ID,
			)
			httputil.UpdateCookie(w, cookie)
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
				h.continueLoginFlow(w, r, path)
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

		h.continueLoginFlow(w, r, path)
	}
	gotoLogin := func() {
		h.continueLoginFlow(w, r, "/login")
	}

	// ctrl.Serve() always write response.
	// So we have to put http.Redirect before it.
	defer ctrl.Serve()

	ctrl.Get(func() error {
		// When promote anonymous user, the end-user should not see this page.
		if webSession != nil && webSession.LoginHint != "" {
			h.continueLoginFlow(w, r, "/flows/promote_user")
			return nil
		}

		opts := webapp.SessionOptions{
			RedirectURI: ctrl.RedirectURI(),
		}

		// When UserIDHint is present, the end-user should never need to select anything in /select_account,
		// so this if block always ends with a return statement, and each branch must write response.
		if userIDHint != "" {
			if loginPrompt && canUseIntentReauthenticate {
				// Reauthentication
				// 1. UserIDHint present
				// 2. prompt=login
				// 3. canUseIntentReauthenticate
				// 4. user.CanReauthenticate

				user, err := h.Users.Get(userIDHint, accesscontrol.RoleGreatest)
				if err != nil {
					return err
				}

				if !user.CanReauthenticate {
					return interaction.ErrNoAuthenticator
				}

				intent := &intents.IntentReauthenticate{
					WebhookState:             webSession.WebhookState,
					UserIDHint:               userIDHint,
					SuppressIDPSessionCookie: suppressIDPSessionCookie,
				}
				result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
					return nil, nil
				})
				if err != nil {
					return err
				}
				result.WriteResponse(w, r)
			} else if !loginPrompt && idpSession != nil && idpSession.GetAuthenticationInfo().UserID == userIDHint {
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

		if !fromAuthzEndpoint || idpSession == nil || loginPrompt {
			gotoSignupOrLogin()
			return nil
		}

		data, err := h.GetData(r, w, idpSession.GetAuthenticationInfo().UserID)
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

func (h *SelectAccountHandler) continueLoginFlow(w http.ResponseWriter, r *http.Request, path string) {
	// preserve query only when continuing the login flow
	u := webapp.MakeRelativeURL(path, webapp.PreserveQuery(r.URL.Query()))
	http.Redirect(w, r, u.String(), http.StatusFound)
}
