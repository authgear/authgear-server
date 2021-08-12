package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
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
		WithPathPattern("/select_account")
}

type SelectAccountUserService interface {
	Get(userID string) (*model.User, error)
}

type SelectAccountIdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type SelectAccountViewModel struct {
	IdentityDisplayName string
}

type SelectAccountHandler struct {
	ControllerFactory    ControllerFactory
	BaseViewModel        *viewmodels.BaseViewModeler
	Renderer             Renderer
	AuthenticationConfig *config.AuthenticationConfig
	SignedUpCookie       webapp.SignedUpCookieDef
	Users                SelectAccountUserService
	Identities           SelectAccountIdentityService
	Cookies              CookieManager
}

func (h *SelectAccountHandler) GetData(r *http.Request, rw http.ResponseWriter, userID string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.EmbedForm(data, r.Form)
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

func (h *SelectAccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	webSession := webapp.GetSession(r.Context())
	continueWithCurrentAccount := func() error {
		redirectURI := "/settings"
		// continue to use the previous session
		// complete the web session and redirect to web session's RedirectURI
		if webSession != nil {
			redirectURI = webSession.RedirectURI
			if err := ctrl.DeleteSession(webSession.ID); err != nil {
				return err
			}
		}
		http.Redirect(w, r, redirectURI, http.StatusFound)
		return nil
	}
	gotoSignupOrLogin := func() {
		signedUpCookie, err := h.Cookies.GetCookie(r, h.SignedUpCookie.Def)
		signedUp := (err == nil && signedUpCookie.Value == "true")
		path := GetAuthenticationEndpoint(signedUp, h.AuthenticationConfig.PublicSignupDisabled)
		http.Redirect(w, r, path, http.StatusFound)
	}
	gotoLogin := func() {
		http.Redirect(w, r, "/login", http.StatusFound)
	}

	// ctrl.Serve() always write response.
	// So we have to put http.Redirect before it.
	defer ctrl.Serve()

	ctrl.Get(func() error {
		loginPrompt := false
		fromAuthzEndpoint := false
		userIDHint := ""
		canUseIntentReauthenticate := false

		if webSession != nil {
			loginPrompt = slice.ContainsString(webSession.Prompt, "login")
			fromAuthzEndpoint = webSession.ClientID != ""
			userIDHint = webSession.UserIDHint
			canUseIntentReauthenticate = webSession.CanUseIntentReauthenticate
		}

		opts := webapp.SessionOptions{
			RedirectURI: ctrl.RedirectURI(),
		}

		// When UserIDHint is present, the end-user should never need to select anything in /select_account,
		// so this if block always ends with a return statement, and each branch must write response.
		if userIDHint != "" {
			// When id_token_hint is present, we have a limitation that is not specified in the OIDC spec.
			// The limitation is that, when id_token_hint is present, an intention of reauthentication is assumed.
			// Therefore, the user indicated by the id_token_hint must be able to reauthenticate.
			user, err := h.Users.Get(userIDHint)
			if err != nil {
				return err
			}
			if !user.CanReauthenticate {
				return interaction.ErrNoAuthenticator
			}

			// The current session is the same user, reauthenticate the user if needed.
			if canUseIntentReauthenticate {
				if loginPrompt {
					intent := &intents.IntentReauthenticate{
						WebhookState: webSession.WebhookState,
						UserIDHint:   userIDHint,
					}
					result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
						return nil, nil
					})
					if err != nil {
						return err
					}
					result.WriteResponse(w, r)
				} else {
					// Otherwise, select the current account because this is the only
					// consequence that should happen.
					err := continueWithCurrentAccount()
					if err != nil {
						return err
					}
				}
			} else {
				// There is no session or the session is another user,
				// redirect to /login so that the end-user could reauthenticate as UserIDHint.
				gotoLogin()
			}
			return nil
		}

		sess := session.GetSession(r.Context())
		// If anything of the following condition holds,
		// the end-user does not need to select anything.
		// 1. The request is not from the authorization endpoint, e.g. /
		// 2. There is no session, so nothing to select.
		// 3. prompt=login, in this case, the end-user cannot select existing account.
		if !fromAuthzEndpoint || sess == nil || loginPrompt {
			gotoSignupOrLogin()
			return nil
		}

		data, err := h.GetData(r, w, sess.GetUserID())
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

func GetAuthenticationEndpoint(signedUp bool, publicSignupDisabled bool) string {
	path := "/signup"
	if publicSignupDisabled || signedUp {
		path = "/login"
	}

	return path
}
