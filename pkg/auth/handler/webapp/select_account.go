package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
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

type SelectAccountHandler struct {
	ControllerFactory    ControllerFactory
	BaseViewModel        *viewmodels.BaseViewModeler
	Renderer             Renderer
	AuthenticationConfig *config.AuthenticationConfig
	SignedUpCookie       webapp.SignedUpCookieDef
}

func (h *SelectAccountHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.EmbedForm(data, r.Form)
	viewmodels.Embed(data, baseViewModel)
	return data, nil
}

func (h *SelectAccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	userID := session.GetUserID(r.Context())
	webSession := webapp.GetSession(r.Context())

	loginPrompt := false
	fromAuthzEndpoint := false
	if webSession != nil {
		// stay in the auth entry point if prompt = login
		loginPrompt = slice.ContainsString(webSession.Prompt, "login")
		fromAuthzEndpoint = webSession.ClientID != ""
	}

	if !fromAuthzEndpoint || userID == nil || loginPrompt {
		signedUpCookie, err := r.Cookie(h.SignedUpCookie.Def.Name)
		signedUp := (err == nil && signedUpCookie.Value == "true")
		path := GetAuthenticationEndpoint(signedUp, h.AuthenticationConfig.PublicSignupDisabled)
		http.Redirect(w, r, path, http.StatusFound)
	}

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}
		h.Renderer.RenderHTML(w, r, TemplateWebSelectAccountHTML, data)
		return nil
	})

	ctrl.PostAction("continue", func() error {
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
	})

	ctrl.PostAction("login", func() error {
		http.Redirect(w, r, "/login", http.StatusFound)
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
