package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebEnterPasswordHTML = template.RegisterHTML(
	"web/enter_password.html",
	components...,
)

const EnterPasswordRequestSchema = "EnterPasswordRequestSchema"

var EnterPasswordSchema = validation.NewMultipartSchema("").
	Add(EnterPasswordRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_password": { "type": "string" }
			},
			"required": ["x_password"]
		}
	`).Instantiate()

func ConfigureEnterPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/enter_password")
}

type EnterPasswordViewModel struct {
	IdentityDisplayID string
	// ForgotPasswordInputType is either phone or email
	ForgotPasswordInputType   string
	ForgotPasswordLoginID     string
	ForgotPasswordCallingCode string
	ForgotPasswordNational    string
}

type EnterPasswordHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
}

func (h *EnterPasswordHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	identityInfo := graph.MustGetUserLastIdentity()

	alternatives := viewmodels.AlternativeStepsViewModel{}
	err := alternatives.AddAuthenticationAlternatives(graph, webapp.SessionStepEnterPassword)
	if err != nil {
		return nil, err
	}

	identityDisplayID := identityInfo.DisplayID()
	phoneFormat := validation.FormatPhone{}
	emailFormat := validation.FormatEmail{AllowName: false}

	forgotPasswordInputType := ""
	forgotPasswordLoginID := ""
	forgotPasswordCallingCode := ""
	forgotPasswordNational := ""
	// Instead of using the login id type, we parse the login id value for the type
	// So user cannot use this flow to check the identity type
	if err := phoneFormat.CheckFormat(identityDisplayID); err == nil {
		forgotPasswordInputType = "phone"
		forgotPasswordNational, forgotPasswordCallingCode, err = phone.ParseE164ToCallingCodeAndNumber(identityDisplayID)
		if err != nil {
			panic("enter_password: cannot parse number: " + err.Error())
		}
	} else if err := emailFormat.CheckFormat(identityDisplayID); err == nil {
		forgotPasswordInputType = "email"
		forgotPasswordLoginID = identityDisplayID
	}

	enterPasswordViewModel := EnterPasswordViewModel{
		IdentityDisplayID:         identityDisplayID,
		ForgotPasswordInputType:   forgotPasswordInputType,
		ForgotPasswordLoginID:     forgotPasswordLoginID,
		ForgotPasswordCallingCode: forgotPasswordCallingCode,
		ForgotPasswordNational:    forgotPasswordNational,
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, enterPasswordViewModel)
	viewmodels.Embed(data, alternatives)

	return data, nil
}

func (h *EnterPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		h.Renderer.RenderHTML(w, r, TemplateWebEnterPasswordHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			err = EnterPasswordSchema.PartValidator(EnterPasswordRequestSchema).ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			plainPassword := r.Form.Get("x_password")
			deviceToken := r.Form.Get("x_device_token") == "true"

			input = &InputAuthPassword{
				Password:    plainPassword,
				DeviceToken: deviceToken,
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
