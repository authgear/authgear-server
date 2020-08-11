package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/core/phone"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/mail"
	"github.com/authgear/authgear-server/pkg/template"
	"github.com/authgear/authgear-server/pkg/validation"
)

const (
	TemplateItemTypeAuthUIEnterOOBOTPHTML config.TemplateItemType = "auth_ui_enter_oob_otp.html"
)

var TemplateAuthUIEnterOOBOTPHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIEnterOOBOTPHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
}

const EnterOOBOTPRequestSchema = "EnterOOBOTPRequestSchema"

var EnterOOBOTPSchema = validation.NewMultipartSchema("").
	Add(EnterOOBOTPRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_password": { "type": "string" }
			},
			"required": ["x_password"]
		}
	`).Instantiate()

func ConfigureEnterOOBOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/enter_oob_otp")
}

type EnterOOBOTPViewModel struct {
	OOBOTPTarget           string
	OOBOTPCodeSendCooldown int
	OOBOTPCodeLength       int
	OOBOTPChannel          string
	Alternatives           []AuthenticationAlternative
}

type EnterOOBOTPHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
}

type EnterOOBOTPNode interface {
	GetOOBOTPTarget() string
	GetOOBOTPChannel() string
	GetOOBOTPCodeSendCooldown() int
	GetOOBOTPCodeLength() int
}

func (h *EnterOOBOTPHandler) GetData(r *http.Request, state *webapp.State, graph *newinteraction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, state.Error)
	viewModel := EnterOOBOTPViewModel{}
	var n EnterOOBOTPNode
	if graph.FindLastNode(&n) {
		viewModel.OOBOTPCodeSendCooldown = n.GetOOBOTPCodeSendCooldown()
		viewModel.OOBOTPCodeLength = n.GetOOBOTPCodeLength()
		viewModel.OOBOTPChannel = n.GetOOBOTPChannel()

		switch authn.AuthenticatorOOBChannel(viewModel.OOBOTPChannel) {
		case authn.AuthenticatorOOBChannelEmail:
			viewModel.OOBOTPTarget = mail.MaskAddress(n.GetOOBOTPTarget())
		case authn.AuthenticatorOOBChannelSMS:
			viewModel.OOBOTPTarget = phone.Mask(n.GetOOBOTPTarget())
		}
	}

	viewModel.Alternatives = DeriveAuthenticationAlternatives(
		// Use previous state ID because the current node should be NodeAuthenticationOOBTrigger.
		state.PrevID,
		graph,
		authn.AuthenticatorTypeOOB,
		n.GetOOBOTPTarget(),
	)

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

type EnterOOBOTPResend struct{}

func (i *EnterOOBOTPResend) DoResend() {}

type EnterOOBOTPInput struct {
	Code string
}

// GetOOBOTP implements InputAuthenticationOOB.
func (i *EnterOOBOTPInput) GetOOBOTP() string {
	return i.Code
}

func (h *EnterOOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		h.Database.WithTx(func() error {
			state, graph, err := h.WebApp.Get(StateID(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			data, err := h.GetData(r, state, graph)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUIEnterOOBOTPHTML, data)
			return nil
		})
	}

	trigger := r.Form.Get("trigger") == "true"

	if r.Method == "POST" && trigger {
		h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				input = &EnterOOBOTPResend{}
				return
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}
			result.WriteResponse(w, r)
			return nil
		})
		return
	}

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				err = EnterOOBOTPSchema.PartValidator(EnterOOBOTPRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				code := r.Form.Get("x_password")

				input = &EnterOOBOTPInput{
					Code: code,
				}
				return
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}
			result.WriteResponse(w, r)
			return nil
		})
	}
}
