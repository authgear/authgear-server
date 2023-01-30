package webapp

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
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
			"x_oob_otp_code": { "type": "string" }
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
	StateQuery MagicLinkOTPPageQueryState
}

func NewVerifyMagicLinkOTPViewModel(r *http.Request) VerifyMagicLinkOTPViewModel {
	target := r.URL.Query().Get("target")
	code := r.URL.Query().Get("token")

	return VerifyMagicLinkOTPViewModel{
		Target:     target,
		Code:       code,
		StateQuery: GetMagicLinkStateFromQuery(r),
	}
}

type VerifyMagicLinkOTPHandler struct {
	MagicLinkOTPCodeService     otp.Service
	GlobalSessionServiceFactory *GlobalSessionServiceFactory
	ControllerFactory           ControllerFactory
	BaseViewModel               *viewmodels.BaseViewModeler
	AuthenticationViewModel     *viewmodels.AuthenticationViewModeler
	Renderer                    Renderer
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

	ctrl.PostAction("", func() error {
		err := VerifyMagicLinkOTPSchema.Validator().ValidateValue(FormToJSON(r.Form))
		var state MagicLinkOTPPageQueryState = MagicLinkOTPPageQueryStateInitial

		if err != nil {
			return err
		}

		code := r.Form.Get("x_oob_otp_code")
		target := r.Form.Get("x_oob_otp_target")

		codeModel, err := h.MagicLinkOTPCodeService.SetUserInputtedCode(target, code)
		if err != nil {
			if errors.Is(err, otp.ErrCodeNotFound) {
				state = MagicLinkOTPPageQueryStateInvalidCode
			} else {
				return err
			}
		}

		_, err = h.MagicLinkOTPCodeService.VerifyMagicLinkCode(target, false)
		if err != nil {
			state = MagicLinkOTPPageQueryStateInvalidCode
		} else if state == MagicLinkOTPPageQueryStateInitial {
			state = MagicLinkOTPPageQueryStateMatched
		}

		if state == MagicLinkOTPPageQueryStateMatched {
			// Update the web session and trigger the refresh event
			webSessionProvider := h.GlobalSessionServiceFactory.NewGlobalSessionService(
				config.AppID(codeModel.AppID),
			)
			webSession, err := webSessionProvider.GetSession(codeModel.WebSessionID)
			if err != nil {
				return err
			}
			err = webSessionProvider.UpdateSession(webSession)
			if err != nil {
				return err
			}
		}

		url := url.URL{}
		url.Path = r.URL.Path
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
}
