package webapp

import (
	"context"
	"net/http"

	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/successpage"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsDeleteAccountHTML = template.RegisterHTML(
	"web/settings_delete_account.html",
	Components...,
)

func ConfigureSettingsDeleteAccountRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/delete_account")
}

type SettingsDeleteAccountViewModel struct {
	ExpectedAccountDeletionTime time.Time
}

type SettingsDeleteAccountUserService interface {
	ScheduleDeletionByEndUser(ctx context.Context, userID string) error
}

type SettingsDeleteAccountOAuthSessionService interface {
	Get(ctx context.Context, entryID string) (*oauthsession.Entry, error)
	Save(ctx context.Context, entry *oauthsession.Entry) error
}

type SettingsDeleteAccountSessionStore interface {
	Create(ctx context.Context, session *webapp.Session) (err error)
	Delete(ctx context.Context, id string) (err error)
	Update(ctx context.Context, session *webapp.Session) (err error)
}

type SettingsDeleteAccountAuthenticationInfoService interface {
	Save(ctx context.Context, entry *authenticationinfo.Entry) (err error)
}

type SettingsDeleteAccountHandler struct {
	ControllerFactory         ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	Renderer                  Renderer
	AccountDeletion           *config.AccountDeletionConfig
	Clock                     clock.Clock
	Users                     SettingsDeleteAccountUserService
	Cookies                   CookieManager
	OAuthSessions             SettingsDeleteAccountOAuthSessionService
	Sessions                  SettingsDeleteAccountSessionStore
	SessionCookie             webapp.SessionCookieDef
	AuthenticationInfoService SettingsDeleteAccountAuthenticationInfoService
}

func (h *SettingsDeleteAccountHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)

	now := h.Clock.NowUTC()
	deletionTime := now.Add(h.AccountDeletion.GracePeriod.Duration())
	viewModel := SettingsDeleteAccountViewModel{
		ExpectedAccountDeletionTime: deletionTime,
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

func (h *SettingsDeleteAccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx(r.Context())

	currentSession := session.GetSession(r.Context())
	redirectURI := "/settings/delete_account/success"
	webSession := webapp.GetSession(r.Context())

	ctrl.Get(func(ctx context.Context) error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		if !h.AccountDeletion.ScheduledByEndUserEnabled {
			http.Redirect(w, r, "/settings", http.StatusFound)
		} else {
			h.Renderer.RenderHTML(w, r, TemplateWebSettingsDeleteAccountHTML, data)
		}
		return nil
	})

	ctrl.PostAction("delete", func(ctx context.Context) error {
		confirmation := r.Form.Get("delete")
		isConfirmed := confirmation == "DELETE"
		if !isConfirmed {
			return apierrors.NewInvalid("confirmation is required to delete account")
		}

		err := h.Users.ScheduleDeletionByEndUser(ctx, currentSession.GetAuthenticationInfo().UserID)
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
