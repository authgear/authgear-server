package webapp

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

const (
	TemplateItemTypeAuthUIEnterOOBOTPHTML string = "auth_ui_enter_oob_otp.html"
)

var TemplateAuthUIEnterOOBOTPHTML = template.T{
	Type:                    TemplateItemTypeAuthUIEnterOOBOTPHTML,
	IsHTML:                  true,
	TranslationTemplateType: TemplateItemTypeAuthUITranslationJSON,
	Defines:                 defines,
	ComponentTemplateTypes:  components,
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
	OOBOTPTarget                    string
	OOBOTPCodeSendCooldown          int
	OOBOTPCodeLength                int
	OOBOTPChannel                   string
	AuthenticationAlternatives      []AuthenticationAlternative
	CreateAuthenticatorAlternatives []CreateAuthenticatorAlternative
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

func (h *EnterOOBOTPHandler) GetData(r *http.Request, state *webapp.State, graph *interaction.Graph) (map[string]interface{}, error) {
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

	currentNode := graph.CurrentNode()
	switch currentNode.(type) {
	case *nodes.NodeAuthenticationOOBTrigger:
		viewModel.AuthenticationAlternatives = DeriveAuthenticationAlternatives(
			// Use previous state ID because the current node is NodeAuthenticationOOBTrigger.
			state.PrevID,
			graph,
			AuthenticationTypeOOB,
			n.GetOOBOTPTarget(),
		)
	case *nodes.NodeCreateAuthenticatorOOBSetup:
		alternatives, err := DeriveCreateAuthenticatorAlternatives(
			// Use previous state ID because the current node is NodeCreateAuthenticatorOOBSetup.
			state.PrevID,
			graph,
			authn.AuthenticatorTypeOOB,
		)
		if err != nil {
			return nil, err
		}
		viewModel.CreateAuthenticatorAlternatives = alternatives
	default:
		panic(fmt.Errorf("enter_oob_otp: unexpected node: %T", currentNode))
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

type EnterOOBOTPResend struct{}

func (i *EnterOOBOTPResend) DoResend() {}

type EnterOOBOTPInput struct {
	Code        string
	DeviceToken bool
}

// GetOOBOTP implements InputAuthenticationOOB.
func (i *EnterOOBOTPInput) GetOOBOTP() string {
	return i.Code
}

// CreateDeviceToken implements InputCreateDeviceToken.
func (i *EnterOOBOTPInput) CreateDeviceToken() bool {
	return i.DeviceToken
}

func (h *EnterOOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		err := h.Database.WithTx(func() error {
			state, graph, err := h.WebApp.Get(StateID(r))
			if err != nil {
				return err
			}

			data, err := h.GetData(r, state, graph)
			if err != nil {
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUIEnterOOBOTPHTML, data)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	trigger := r.Form.Get("trigger") == "true"

	if r.Method == "POST" && trigger {
		err := h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				input = &EnterOOBOTPResend{}
				return
			})
			if err != nil {
				return err
			}
			result.WriteResponse(w, r)
			return nil
		})
		if err != nil {
			panic(err)
		}
		return
	}

	if r.Method == "POST" {
		err := h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				err = EnterOOBOTPSchema.PartValidator(EnterOOBOTPRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				code := r.Form.Get("x_password")
				deviceToken := r.Form.Get("x_device_token") == "true"

				input = &EnterOOBOTPInput{
					Code:        code,
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
		if err != nil {
			panic(err)
		}
	}
}
