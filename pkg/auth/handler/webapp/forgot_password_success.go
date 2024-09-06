package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebForgotPasswordSuccessHTML = template.RegisterHTML(
	"web/forgot_password_success.html",
	Components...,
)

func ConfigureForgotPasswordSuccessRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/flows/forgot_password/success")
}

type ForgotPasswordSuccessViewModel struct {
	GivenLoginID string
}

type ForgotPasswordSuccessNode interface {
	GetLoginID() string
}

type ForgotPasswordSuccessHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
}

func (h *ForgotPasswordSuccessHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	forgotPasswordSuccessViewModel := ForgotPasswordSuccessViewModel{}

	if loginID, ok := session.Extra["login_id"]; ok {
		forgotPasswordSuccessViewModel.GivenLoginID = loginID.(string)
	}
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, forgotPasswordSuccessViewModel)
	return data, nil
}

func (h *ForgotPasswordSuccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

	ctrl.Get(func() error {
		session, err := ctrl.InteractionSession()
		if err != nil {
			return err
		}

		data, err := h.GetData(r, w, session)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebForgotPasswordSuccessHTML, data)
		return nil
	})
}
