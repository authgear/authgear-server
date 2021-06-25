package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
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

type IdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type SelectAccountViewModel struct {
	IdentityDisplayName string
}

type SelectAccountHandler struct {
	Database             *appdb.Handle
	ControllerFactory    ControllerFactory
	BaseViewModel        *viewmodels.BaseViewModeler
	Renderer             Renderer
	AuthenticationConfig *config.AuthenticationConfig
	SignedUpCookie       webapp.SignedUpCookieDef
	Identities           IdentityService
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

	sess := session.GetSession(r.Context())
	webSession := webapp.GetSession(r.Context())
	loginPrompt := false
	fromAuthzEndpoint := false
	userIDHint := ""

	gotoLogin := func() {
		http.Redirect(w, r, "/login", http.StatusFound)
	}

	gotoSignupOrLogin := func() {
		signedUpCookie, err := r.Cookie(h.SignedUpCookie.Def.Name)
		signedUp := (err == nil && signedUpCookie.Value == "true")
		path := GetAuthenticationEndpoint(signedUp, h.AuthenticationConfig.PublicSignupDisabled)
		http.Redirect(w, r, path, http.StatusFound)
	}

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

	if webSession != nil {
		loginPrompt = slice.ContainsString(webSession.Prompt, "login")
		fromAuthzEndpoint = webSession.ClientID != ""
		userIDHint = webSession.UserIDHint
	}

	opts := webapp.SessionOptions{
		RedirectURI: ctrl.RedirectURI(),
	}

	// When UserIDHint is present, the end-user should never need to select anything in /select_account,
	// so this if block always ends with a return statement, and each branch must write response.
	if userIDHint != "" {
		err := h.Database.WithTx(func() error {
			// The current session is the same user, reauthenticate the user if needed.
			if sess != nil && sess.GetUserID() == userIDHint {
				if loginPrompt {
					intent := &intents.IntentReauthenticate{
						WebhookState: webSession.WebhookState,
						UserIDHint:   userIDHint,
						IDPSessionID: sess.SessionID(),
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
		})
		if err != nil {
			panic(err)
		}
		return
	}

	// If anything of the following condition holds,
	// the end-user does not need to select anything.
	// 1. The request is not from the authorization endpoint, e.g. /
	// 2. There is no session, so nothing to select.
	// 3. prompt=login, in this case, the end-user cannot select existing account.
	if !fromAuthzEndpoint || sess == nil || loginPrompt {
		gotoSignupOrLogin()
		return
	}

	// ctrl.Serve() always write response.
	// So we have to put http.Redirect before it.
	defer ctrl.Serve()

	ctrl.Get(func() error {
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
