package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
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
	GetMagicLinkOTP() string
	GetMagicLinkOTPTarget() string
	GetMagicLinkOTPChannel() string
	GetMagicLinkOTPOOBType() interaction.OOBType
}

type MagicLinkOTPViewModel struct {
	Target              string
	OTPCodeSendCooldown int
	StateQuery          MagicLinkOTPPageQueryState
}

type MagicLinkOTPHandler struct {
	MagicLinkOTPCodeService   otp.Service
	ControllerFactory         ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	AlternativeStepsViewModel *viewmodels.AlternativeStepsViewModeler
	Renderer                  Renderer
	RateLimiter               RateLimiter
	FlashMessage              FlashMessage
	AntiSpamOTPCodeBucket     AntiSpamOTPCodeBucketMaker
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
		channel := model.AuthenticatorOOBChannel(n.GetMagicLinkOTPChannel())
		target := n.GetMagicLinkOTPTarget()
		bucket := h.AntiSpamOTPCodeBucket.MakeBucket(channel, target)
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

	getCodeFromGraph := func() (string, error) {
		graph, err := ctrl.InteractionGet()
		if err != nil {
			return "", err
		}
		var code string
		var n MagicLinkOTPNode
		if graph.FindLastNode(&n) {
			code = n.GetMagicLinkOTP()
		} else {
			panic(fmt.Errorf("webapp: unexpected node for magic link: %T", n))
		}

		return code, nil
	}

	ctrl.PostAction("dryrun_verify", func() error {
		// nolint: ineffassign
		var state MagicLinkOTPPageQueryState = MagicLinkOTPPageQueryStateInitial

		code, err := getCodeFromGraph()
		if err != nil {
			return err
		}

		_, err = h.MagicLinkOTPCodeService.VerifyMagicLinkCode(code, false)
		if err == nil {
			state = MagicLinkOTPPageQueryStateMatched
		} else if errors.Is(err, otp.ErrInvalidCode) {
			state = MagicLinkOTPPageQueryStateInvalidCode
		} else {
			return err
		}

		url := url.URL{Path: r.URL.Path}
		query := r.URL.Query()
		query.Set(MagicLinkOTPPageQueryStateKey, string(state))
		url.RawQuery = query.Encode()

		result := webapp.Result{
			RedirectURI:      url.String(),
			NavigationAction: "replace",
		}
		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("next", func() error {
		deviceToken := r.Form.Get("x_device_token") == "true"

		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			code, err := getCodeFromGraph()
			if err != nil {
				return
			}

			input = &InputVerifyMagicLinkOTP{
				Code:        code,
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
