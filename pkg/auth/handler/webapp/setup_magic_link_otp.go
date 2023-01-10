package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSetupMagicLinkOTPHTML = template.RegisterHTML(
	"web/setup_magic_link_otp.html",
	components...,
)

var SetupMagicLinkOTPEmailSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_email": { "type": "string" }
		},
		"required": ["x_email"]
	}
`)

func ConfigureSetupMagicLinkOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/setup_magic_link_otp")
}

type SetupMagicLinkOTPNode interface {
}

type SetupMagicLinkOTPHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
}

func (h *SetupMagicLinkOTPHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)

	viewmodels.Embed(data, baseViewModel)
	return data, nil
}

func (h *SetupMagicLinkOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	ctrl.Get(func() error {
		session, err := ctrl.InteractionSession()
		if err != nil {
			return err
		}

		graph, err := ctrl.InteractionGet()
		if err != nil {
			return err
		}

		data, err := h.GetData(r, w, session, graph)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSetupMagicLinkOTPHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			err = SetupMagicLinkOTPEmailSchema.Validator().ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			target, err := FormToMagicLinkTarget(r.Form)
			if err != nil {
				return
			}

			input = &InputSetupMagicLinkOTP{
				InputType: "email",
				Target:    target,
			}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	handleAlternativeSteps(ctrl)
}

func FormToMagicLinkTarget(form url.Values) (target string, err error) {
	target = form.Get("x_email")
	return
}
