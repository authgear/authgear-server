package webapp

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/setutil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebLoginLinkHTML = template.RegisterHTML(
	"web/login_link_otp.html",
	Components...,
)

func ConfigureLoginLinkOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/login_link_otp")
}

type LoginLinkOTPNode interface {
	GetLoginLinkOTPTarget() string
	GetLoginLinkOTPChannel() string
	GetLoginLinkOTPOOBType() interaction.OOBType
}

type LoginLinkOTPViewModel struct {
	Target              string
	OTPCodeSendCooldown int
	StateQuery          LoginLinkOTPPageQueryState
}

type LoginLinkOTPHandler struct {
	Clock                     clock.Clock
	LoginLinkOTPCodeService   OTPCodeService
	ControllerFactory         ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	AlternativeStepsViewModel *viewmodels.AlternativeStepsViewModeler
	Renderer                  Renderer
	FlashMessage              FlashMessage
	Config                    *config.AppConfig
}

func (h *LoginLinkOTPHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewModel := LoginLinkOTPViewModel{
		StateQuery: GetLoginLinkStateFromQuery(r),
	}
	var alternatives *viewmodels.AlternativeStepsViewModel

	var n LoginLinkOTPNode
	if graph.FindLastNode(&n) {
		channel := model.AuthenticatorOOBChannel(n.GetLoginLinkOTPChannel())
		target := n.GetLoginLinkOTPTarget()

		state, err := h.LoginLinkOTPCodeService.InspectState(
			otp.KindOOBOTPLink(h.Config, channel),
			target,
		)
		if err != nil {
			return nil, err
		}

		cooldown := int(state.CanResendAt.Sub(h.Clock.NowUTC()).Seconds())
		if cooldown < 0 {
			viewModel.OTPCodeSendCooldown = 0
		} else {
			viewModel.OTPCodeSendCooldown = cooldown
		}

		viewModel.Target = mail.MaskAddress(n.GetLoginLinkOTPTarget())
	}

	currentNode := graph.CurrentNode()
	switch currentNode.(type) {
	case *nodes.NodeAuthenticationLoginLinkTrigger:
		var err error
		alternatives, err = h.AlternativeStepsViewModel.AuthenticationAlternatives(graph, webapp.SessionStepVerifyLoginLinkOTPAuthn)
		if err != nil {
			return nil, err
		}
	case *nodes.NodeCreateAuthenticatorLoginLinkOTPSetup:
		var err error
		alternatives, err = h.AlternativeStepsViewModel.CreateAuthenticatorAlternatives(graph, webapp.SessionStepSetupLoginLinkOTP)
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

func (h *LoginLinkOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		h.Renderer.RenderHTML(w, r, TemplateWebLoginLinkHTML, data)
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
			h.FlashMessage.Flash(w, string(webapp.FlashMessageTypeResendLoginLinkSuccess))
		}
		result.WriteResponse(w, r)
		return nil
	})

	getEmailFromGraph := func() (string, error) {
		graph, err := ctrl.InteractionGet()
		if err != nil {
			return "", err
		}
		var code string
		var n LoginLinkOTPNode
		if graph.FindLastNode(&n) {
			code = n.GetLoginLinkOTPTarget()
		} else {
			panic(fmt.Errorf("webapp: unexpected node for login link: %T", n))
		}

		return code, nil
	}

	ctrl.PostAction("dryrun_verify", func() error {
		var state LoginLinkOTPPageQueryState

		email, err := getEmailFromGraph()
		if err != nil {
			return err
		}

		kind := otp.KindOOBOTPLink(h.Config, model.AuthenticatorOOBChannelEmail)
		err = h.LoginLinkOTPCodeService.VerifyOTP(
			kind, email, "", &otp.VerifyOptions{UseSubmittedCode: true, SkipConsume: true},
		)
		if err == nil {
			state = LoginLinkOTPPageQueryStateMatched
		} else if apierrors.IsKind(err, otp.InvalidOTPCode) {
			state = LoginLinkOTPPageQueryStateInvalidCode
		} else {
			return err
		}

		url := url.URL{Path: r.URL.Path}
		query := r.URL.Query()
		query.Set(LoginLinkOTPPageQueryStateKey, string(state))
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
			input = &InputVerifyLoginLinkOTP{
				DeviceToken: deviceToken,
			}
			return
		})
		if err != nil {
			return err
		}

		result.RemoveQueries = setutil.Set[string]{
			LoginLinkOTPPageQueryStateKey: struct{}{},
		}
		result.WriteResponse(w, r)
		return nil
	})

	handleAlternativeSteps(ctrl)
}
