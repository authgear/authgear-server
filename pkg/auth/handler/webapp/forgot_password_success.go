package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebForgotPasswordSuccessHTML = template.RegisterHTML(
	"web/forgot_password_success.html",
	components...,
)

func ConfigureForgotPasswordSuccessRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/forgot_password/success")
}

type ForgotPasswordSuccessViewModel struct {
	GivenLoginID string
	RedirectURI  string
}

type ForgotPasswordSuccessNode interface {
	GetLoginID() string
}

type ForgotPasswordSuccessHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
}

func (h *ForgotPasswordSuccessHandler) GetData(r *http.Request, state *webapp.State) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, state.Error)
	forgotPasswordSuccessViewModel := ForgotPasswordSuccessViewModel{}

	redirectURI := "/login"
	if rawRedirectURI, ok := state.Extra["redirect_uri"].(string); ok {
		u, err := url.Parse(rawRedirectURI)
		if err == nil {
			redirectURI = (&url.URL{
				Path:     u.Path,
				RawQuery: u.RawQuery,
			}).String()
		}
	}
	forgotPasswordSuccessViewModel.RedirectURI = redirectURI

	if loginID, ok := state.Extra["login_id"]; ok {
		forgotPasswordSuccessViewModel.GivenLoginID = loginID.(string)
	}
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, forgotPasswordSuccessViewModel)
	return data, nil
}

func (h *ForgotPasswordSuccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		err := h.Database.WithTx(func() error {
			state, err := h.WebApp.GetState(StateID(r))
			if err != nil {
				return err
			}

			data, err := h.GetData(r, state)
			if err != nil {
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateWebForgotPasswordSuccessHTML, data)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
}
