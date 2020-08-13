package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/template"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

const (
	// nolint: gosec
	TemplateItemTypeAuthUIForgotPasswordSuccessHTML config.TemplateItemType = "auth_ui_forgot_password_success.html"
)

var TemplateAuthUIForgotPasswordSuccessHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIForgotPasswordSuccessHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
}

func ConfigureForgotPasswordSuccessRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/forgot_password/success")
}

type ForgotPasswordSuccessViewModel struct {
	GivenLoginID string
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
	if loginID, ok := state.Extra["login_id"]; ok {
		forgotPasswordSuccessViewModel.GivenLoginID = loginID.(string)
	}
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, forgotPasswordSuccessViewModel)
	return data, nil
}

func (h *ForgotPasswordSuccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		h.Database.WithTx(func() error {
			state, err := h.WebApp.GetState(StateID(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			data, err := h.GetData(r, state)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUIForgotPasswordSuccessHTML, data)
			return nil
		})
	}
}
