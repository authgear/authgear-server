package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlslosession"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/setutil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebLogoutHTML = template.RegisterHTML(
	"web/logout.html",
	Components...,
)

func ConfigureLogoutRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/logout")
}

type LogoutSessionManager interface {
	Logout(session.SessionBase, http.ResponseWriter) ([]session.ListableSession, error)
}

type SAMLSLOSessionService interface {
	Get(sessionID string) (entry *samlslosession.SAMLSLOSession, err error)
	Save(session *samlslosession.SAMLSLOSession) (err error)
}

type SAMLSLOService interface {
	SendSLORequest(
		rw http.ResponseWriter,
		r *http.Request,
		sloSession *samlslosession.SAMLSLOSession,
		sp *config.SAMLServiceProviderConfig,
	) error
}

type LogoutHandler struct {
	ControllerFactory     ControllerFactory
	Database              *appdb.Handle
	TrustProxy            config.TrustProxy
	OAuth                 *config.OAuthConfig
	UIConfig              *config.UIConfig
	SAMLConfig            *config.SAMLConfig
	SessionManager        LogoutSessionManager
	BaseViewModel         *viewmodels.BaseViewModeler
	Renderer              Renderer
	OAuthClientResolver   WebappOAuthClientResolver
	SAMLSLOSessionService SAMLSLOSessionService
	SAMLSLOService        SAMLSLOService
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	ctrl.Get(func() error {
		baseViewModel := h.BaseViewModel.ViewModel(r, w)

		data := map[string]interface{}{}

		viewmodels.Embed(data, baseViewModel)

		h.Renderer.RenderHTML(w, r, TemplateWebLogoutHTML, data)
		return nil
	})

	ctrl.PostAction("logout", func() error {
		sess := session.GetSession(r.Context())
		invalidatedSessions, err := h.SessionManager.Logout(sess, w)
		if err != nil {
			return err
		}

		uiParam := uiparam.GetUIParam(r.Context())
		clientID := uiParam.ClientID
		client := h.OAuthClientResolver.ResolveClient(clientID)
		postLogoutRedirectURI := webapp.ResolvePostLogoutRedirectURI(client, r.FormValue("post_logout_redirect_uri"), h.UIConfig)
		redirectURI := webapp.GetRedirectURI(r, bool(h.TrustProxy), postLogoutRedirectURI)

		pendingLogoutServiceProviderIDs := setutil.Set[string]{}
		for _, s := range invalidatedSessions {
			pendingLogoutServiceProviderIDs = pendingLogoutServiceProviderIDs.Merge(s.GetParticipatedSAMLServiceProviderIDsSet())
		}
		if len(pendingLogoutServiceProviderIDs.Keys()) > 0 {
			sloSessionEntry := &samlslosession.SAMLSLOSessionEntry{
				PendingLogoutServiceProviderIDs: pendingLogoutServiceProviderIDs.Keys(),
				SID:                             oidc.EncodeSID(sess),
				UserID:                          sess.GetAuthenticationInfo().UserID,
				PostLogoutRedirectURI:           redirectURI,
			}
			sloSession := samlslosession.NewSAMLSLOSession(sloSessionEntry)
			err := h.SAMLSLOSessionService.Save(sloSession)
			if err != nil {
				return err
			}
			// Send the logout request to the first sp
			for _, spID := range pendingLogoutServiceProviderIDs.Keys() {
				sp, ok := h.SAMLConfig.ResolveProvider(spID)
				if ok && sp.SLOEnabled {
					return h.SAMLSLOService.SendSLORequest(w, r, sloSession, sp)
				}
			}
		}

		// If no saml service provider is pending logout
		http.Redirect(w, r, redirectURI, http.StatusFound)
		return nil

	})
}
