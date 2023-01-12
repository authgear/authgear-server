package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebVerifyMagicLinkOTPHTML = template.RegisterHTML(
	"web/verify_magic_link.html",
	components...,
)

var VerifyMagicLinkOTPSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_oob_otp_target": { "type": "string" },
			"x_oob_otp_code": { "type": "string" },
			"x_oob_otp_redirect_to": { "type": "string" }
		},
		"required": ["x_oob_otp_target", "x_oob_otp_code"]
	}
`)

func ConfigureVerifyMagicLinkOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/verify_magic_link")
}

type VerifyMagicLinkOTPViewModel struct {
	Target     string
	Code       string
	RedirectTo string
}

func NewVerifyMagicLinkOTPViewModel(r *http.Request) VerifyMagicLinkOTPViewModel {
	target := r.URL.Query().Get("target")
	code := r.URL.Query().Get("token")
	redirectTo := r.URL.Query().Get("redirect_to")

	return VerifyMagicLinkOTPViewModel{
		Target:     target,
		Code:       code,
		RedirectTo: redirectTo,
	}
}

type VerifyMagicLinkOTPHandler struct {
	ControllerFactory       ControllerFactory
	BaseViewModel           *viewmodels.BaseViewModeler
	AuthenticationViewModel *viewmodels.AuthenticationViewModeler
	Renderer                Renderer
}

func (h *VerifyMagicLinkOTPHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, NewVerifyMagicLinkOTPViewModel(r))
	viewmodels.Embed(data, baseViewModel)
	return data, nil
}

func (h *VerifyMagicLinkOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebVerifyMagicLinkOTPHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			err = VerifyMagicLinkOTPSchema.Validator().ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			code := r.Form.Get("x_oob_otp_code")
			target := r.Form.Get("x_oob_otp_target")
			redirectTo := r.Form.Get("x_oob_otp_redirect_to")

			input = &InputVerifyMagicLinkOTP{
				Target:     target,
				Code:       code,
				RedirectTo: redirectTo,
			}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
}
