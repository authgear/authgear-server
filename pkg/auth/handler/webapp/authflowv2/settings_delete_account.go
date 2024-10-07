package authflowv2

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/successpage"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsV2DeleteAccountHTML = template.RegisterHTML(
	"web/authflowv2/settings_delete_account.html",
	handlerwebapp.SettingsComponents...,
)

type AuthflowV2SettingsDeleteAccountViewModel struct {
	ExpectedAccountDeletionTime time.Time
}

type AuthflowV2SettingsDeleteAccountHandler struct {
	ControllerFactory         handlerwebapp.ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	Renderer                  handlerwebapp.Renderer
	Clock                     clock.Clock
	Cookies                   handlerwebapp.CookieManager
	Users                     handlerwebapp.SettingsDeleteAccountUserService
	Sessions                  handlerwebapp.SettingsDeleteAccountSessionStore
	OAuthSessions             handlerwebapp.SettingsDeleteAccountOAuthSessionService
	AccountDeletion           *config.AccountDeletionConfig
	AuthenticationInfoService handlerwebapp.SettingsDeleteAccountAuthenticationInfoService
}

func (h *AuthflowV2SettingsDeleteAccountHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	// BaseViewModel
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	now := h.Clock.NowUTC()
	deletionTime := now.Add(h.AccountDeletion.GracePeriod.Duration())
	deleteAccountViewModel := AuthflowV2SettingsDeleteAccountViewModel{
		ExpectedAccountDeletionTime: deletionTime,
	}
	viewmodels.Embed(data, deleteAccountViewModel)

	return data, nil
}

func (h *AuthflowV2SettingsDeleteAccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

	currentSession := session.GetSession(r.Context())
	redirectURI := "/settings/delete_account/success"
	webSession := webapp.GetSession(r.Context())

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsV2DeleteAccountHTML, data)

		return nil
	})

	ctrl.PostAction("delete", func() error {
		confirmation := r.Form.Get("delete")
		isConfirmed := confirmation == "DELETE"
		if !isConfirmed {
			return apierrors.NewInvalid("confirmation is required to delete account")
		}

		err := h.Users.ScheduleDeletionByEndUser(currentSession.GetAuthenticationInfo().UserID)
		if err != nil {
			return err
		}

		if webSession != nil && webSession.OAuthSessionID != "" {
			// delete account triggered by sdk via settings action
			// handle settings action result here

			authInfoEntry := authenticationinfo.NewEntry(currentSession.CreateNewAuthenticationInfoByThisSession(), webSession.OAuthSessionID, "")
			err := h.AuthenticationInfoService.Save(authInfoEntry)
			if err != nil {
				return err
			}
			webSession.Extra["authentication_info_id"] = authInfoEntry.ID
			err = h.Sessions.Update(webSession)
			if err != nil {
				return err
			}

			entry, err := h.OAuthSessions.Get(webSession.OAuthSessionID)
			if err != nil {
				return err
			}

			entry.T.SettingsActionResult = oauthsession.NewSettingsActionResult()
			err = h.OAuthSessions.Save(entry)
			if err != nil {
				return err
			}
		}

		// set success page path cookie before visiting success page
		result := webapp.Result{
			RedirectURI: redirectURI,
			Cookies: []*http.Cookie{
				h.Cookies.ValueCookie(successpage.PathCookieDef, redirectURI),
			},
		}
		result.WriteResponse(w, r)
		return nil
	})
}
