package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebEnterRecoveryCodeHTML = template.RegisterHTML(
	"web/enter_recovery_code.html",
	components...,
)

const EnterRecoveryCodeRequestSchema = "EnterRecoveryCodeRequestSchema"

var EnterRecoveryCodeSchema = validation.NewMultipartSchema("").
	Add(EnterRecoveryCodeRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_code": { "type": "string" }
			},
			"required": ["x_code"]
		}
	`).Instantiate()

func ConfigureEnterRecoveryCodeRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/enter_recovery_code")
}

type EnterRecoveryCodeViewModel struct {
	AlternativeSteps []viewmodels.AlternativeStep
}

type EnterRecoveryCodeHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
}

func (h *EnterRecoveryCodeHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)

	alternatives := &viewmodels.AlternativeStepsViewModel{}
	err := alternatives.AddAuthenticationAlternatives(graph, webapp.SessionStepEnterRecoveryCode)
	if err != nil {
		return nil, err
	}

	viewModel := EnterRecoveryCodeViewModel{
		AlternativeSteps: alternatives.AlternativeSteps,
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

func (h *EnterRecoveryCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		h.Renderer.RenderHTML(w, r, TemplateWebEnterRecoveryCodeHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			err = EnterRecoveryCodeSchema.PartValidator(EnterRecoveryCodeRequestSchema).ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			code := r.Form.Get("x_code")

			input = &InputAuthRecoveryCode{
				Code: code,
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
