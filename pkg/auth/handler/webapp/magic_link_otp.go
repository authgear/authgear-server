package webapp

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/setutil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebMagicLinkHTML = template.RegisterHTML(
	"web/magic_link_otp.html",
	components...,
)

func ConfigureMagicLinkOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/magic_link_otp")
}

type MagicLinkOTPNode interface {
	GetMagicLinkOTPTarget() string
	GetMagicLinkOTPOOBType() interaction.OOBType
}

type MagicLinkOTPViewModel struct {
	Target              string
	OTPCodeSendCooldown int
	StateQuery          MagicLinkOTPPageQueryState
}

type MagicLinkOTPHandler struct {
	ControllerFactory         ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	AlternativeStepsViewModel *viewmodels.AlternativeStepsViewModeler
	Renderer                  Renderer
	RateLimiter               RateLimiter
	FlashMessage              FlashMessage
}

func (h *MagicLinkOTPHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewModel := MagicLinkOTPViewModel{
		StateQuery: GetMagicLinkStateFromQuery(r),
	}
	var alternatives *viewmodels.AlternativeStepsViewModel

	var n MagicLinkOTPNode
	if graph.FindLastNode(&n) {
		oobType := n.GetMagicLinkOTPOOBType()
		target := n.GetMagicLinkOTPTarget()
		bucket := interaction.AntiSpamSendOOBCodeBucket(oobType, target)
		pass, resetDuration, err := h.RateLimiter.CheckToken(bucket)
		if err != nil {
			return nil, err
		}
		if pass {
			// allow sending immediately
			viewModel.OTPCodeSendCooldown = 0
		} else {
			viewModel.OTPCodeSendCooldown = int(resetDuration.Seconds())
		}
		viewModel.Target = mail.MaskAddress(n.GetMagicLinkOTPTarget())
	}

	currentNode := graph.CurrentNode()
	switch currentNode.(type) {
	case *nodes.NodeAuthenticationMagicLinkTrigger:
		var err error
		alternatives, err = h.AlternativeStepsViewModel.AuthenticationAlternatives(graph, webapp.SessionStepVerifyMagicLinkOTPAuthn)
		if err != nil {
			return nil, err
		}
	case *nodes.NodeCreateAuthenticatorMagicLinkOTPSetup:
		var err error
		alternatives, err = h.AlternativeStepsViewModel.CreateAuthenticatorAlternatives(graph, webapp.SessionStepSetupMagicLinkOTP)
		if err != nil {
			return nil, err
		}
	default:
		panic(fmt.Errorf("enter_oob_otp: unexpected node: %T", currentNode))
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)
	viewmodels.Embed(data, alternatives)
	return data, nil
}

func (h *MagicLinkOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		h.Renderer.RenderHTML(w, r, TemplateWebMagicLinkHTML, data)
		return nil
	})

	ctrl.PostAction("resend", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			input = &InputResendCode{}
			return
		})
		if err != nil {
			return err
		}

		if !result.IsInteractionErr {
			h.FlashMessage.Flash(w, string(webapp.FlashMessageTypeResendMagicLinkSuccess))
		}
		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("matched", func() error {
		u := url.URL{}
		u.Path = r.URL.Path
		q := r.URL.Query()
		q.Set(MagicLinkOTPPageQueryStateKey, string(MagicLinkOTPPageQueryStateMatched))
		u.RawQuery = q.Encode()
		result := webapp.Result{
			RedirectURI:      u.String(),
			NavigationAction: "replace",
		}
		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("next", func() error {
		deviceToken := r.Form.Get("x_device_token") == "true"

		getTargetFromGraph := func() (string, error) {
			graph, err := ctrl.InteractionGet()
			if err != nil {
				return "", err
			}
			var target string
			var n MagicLinkOTPNode
			if graph.FindLastNode(&n) {
				target = n.GetMagicLinkOTPTarget()
			} else {
				panic(fmt.Errorf("webapp: unexpected node for magic link: %T", n))
			}

			return target, nil
		}

		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			target, err := getTargetFromGraph()
			if err != nil {
				return
			}

			input = &InputVerifyMagicLinkOTP{
				Target:      target,
				DeviceToken: deviceToken,
			}
			return
		})
		if err != nil {
			return err
		}

		result.RemoveQueries = setutil.Set[string]{
			MagicLinkOTPPageQueryStateKey: struct{}{},
		}
		result.WriteResponse(w, r)
		return nil
	})

	handleAlternativeSteps(ctrl)
}
