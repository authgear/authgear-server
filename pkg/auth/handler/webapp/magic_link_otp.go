package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
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
}

type MagicLinkOTPViewModel struct {
	Target     string
	StateQuery MagicLinkOTPPageQueryState
}

type MagicLinkOTPHandler struct {
	ControllerFactory         ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	AlternativeStepsViewModel *viewmodels.AlternativeStepsViewModeler
	Renderer                  Renderer
}

func (h *MagicLinkOTPHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewModel := MagicLinkOTPViewModel{
		StateQuery: GetMagicLinkStateFromQuery(r),
	}
	var n MagicLinkOTPNode
	if graph.FindLastNode(&n) {
		viewModel.Target = mail.MaskAddress(n.GetMagicLinkOTPTarget())
	}
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, viewModel)
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
		// deviceToken := r.Form.Get("x_device_token") == "true"

		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			input = &InputVerifyMagicLinkOTP{
				// DeviceToken: deviceToken,
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
