package authflowv2

import (
	"context"
	"net/http"

	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
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
	Database                  *appdb.Handle
	ControllerFactory         handlerwebapp.ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	SettingsViewModel         *viewmodels.SettingsViewModeler
	Renderer                  handlerwebapp.Renderer
	Clock                     clock.Clock
	Cookies                   handlerwebapp.CookieManager
	Users                     handlerwebapp.SettingsDeleteAccountUserService
	Sessions                  handlerwebapp.SettingsDeleteAccountSessionStore
	OAuthSessions             handlerwebapp.SettingsDeleteAccountOAuthSessionService
	AccountDeletion           *config.AccountDeletionConfig
	AuthenticationInfoService handlerwebapp.SettingsDeleteAccountAuthenticationInfoService
}

func (h *AuthflowV2SettingsDeleteAccountHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	userID := session.GetUserID(ctx)

	// BaseViewModel
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	// SettingsViewModel
	settingsViewModel, err := h.SettingsViewModel.ViewModel(ctx, *userID)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, *settingsViewModel)

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
	defer ctrl.ServeWithoutDBTx(r.Context())

	currentSession := session.GetSession(r.Context())
	redirectURI := "/settings/delete_account/success"
	webSession := webapp.GetSession(r.Context())

	ctrl.Get(func(ctx context.Context) error {
		var data map[string]interface{}
		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			data, err = h.GetData(ctx, r, w)
			return err
		})
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsV2DeleteAccountHTML, data)

		return nil
	})

	ctrl.PostAction("delete", func(ctx context.Context) error {
		confirmation := r.Form.Get("delete")
		isConfirmed := confirmation == "DELETE"
		if !isConfirmed {
			return apierrors.NewInvalid("confirmation is required to delete account")
		}

		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			return h.Users.ScheduleDeletionByEndUser(ctx, currentSession.GetAuthenticationInfo().UserID)
		})
		if err != nil {
			return err
		}

		if webSession != nil && webSession.OAuthSessionID != "" {
			// delete account triggered by sdk via settings action
			// handle settings action result here

			authInfoEntry := authenticationinfo.NewEntry(currentSession.CreateNewAuthenticationInfoByThisSession(), webSession.OAuthSessionID, "")
			err := h.AuthenticationInfoService.Save(ctx, authInfoEntry)
			if err != nil {
				return err
			}
			webSession.Extra["authentication_info_id"] = authInfoEntry.ID
			err = h.Sessions.Update(ctx, webSession)
			if err != nil {
				return err
			}

			entry, err := h.OAuthSessions.Get(ctx, webSession.OAuthSessionID)
			if err != nil {
				return err
			}

			entry.T.SettingsActionResult = oauthsession.NewSettingsActionResult()
			err = h.OAuthSessions.Save(ctx, entry)
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
