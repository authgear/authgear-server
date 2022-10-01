package oauth

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	oauthhandler "github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type Renderer interface {
	RenderHTML(w http.ResponseWriter, r *http.Request, tpl *template.HTML, data interface{})
}

func ConfigureFromWebAppRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST").
		WithPathPattern("/oauth2/_from_webapp")
}

type FromWebAppHandlerLogger struct{ *log.Logger }

func NewFromWebAppHandlerLogger(lf *log.Factory) FromWebAppHandlerLogger {
	return FromWebAppHandlerLogger{lf.New("handler-from-webapp")}
}

type ProtocolFromWebAppHandler interface {
	HandleConsentWithoutUserConsent(req *http.Request) (httputil.Result, *oauthhandler.ConsentRequired)
	HandleConsentWithUserConsent(req *http.Request) httputil.Result
	HandleConsentWithUserCancel(req *http.Request) httputil.Result
}

type FromWebAppViewModel struct {
	ClientName               string
	IsRequestingFullUserInfo bool
}

type FromWebAppHandler struct {
	Logger   FromWebAppHandlerLogger
	Database *appdb.Handle
	Handler  ProtocolFromWebAppHandler

	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
}

func (h *FromWebAppHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var result httputil.Result
	var err error

	err = r.ParseForm()
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}

	switch r.Method {
	case http.MethodGet:
		var consentRequired *oauthhandler.ConsentRequired
		err = h.Database.WithTx(func() error {
			result, consentRequired = h.Handler.HandleConsentWithoutUserConsent(r)
			if consentRequired != nil {
				// return error to rollback transaction
				return errors.New("consent required")
			}
			if result.IsInternalError() {
				return errAuthzInternalError
			}
			return nil
		})
		if consentRequired != nil {
			h.renderConsentPage(rw, r, consentRequired)
			return
		}
	case http.MethodPost:
		if r.Form.Get("x_action") == "consent" {
			err = h.Database.WithTx(func() error {
				result = h.Handler.HandleConsentWithUserConsent(r)
				if result.IsInternalError() {
					return errAuthzInternalError
				}
				return nil
			})
			break
		} else if r.Form.Get("x_action") == "cancel" {
			err = h.Database.WithTx(func() error {
				result = h.Handler.HandleConsentWithUserCancel(r)
				if result.IsInternalError() {
					return errAuthzInternalError
				}
				return nil
			})
			break
		}
		http.Error(rw, "Unknown action", http.StatusBadRequest)
		return
	}

	if err == nil || errors.Is(err, errAuthzInternalError) {
		result.WriteResponse(rw, r)
	} else {
		h.Logger.WithError(err).Error("")
		http.Error(rw, "Internal Server Error", 500)
	}
}

func (h *FromWebAppHandler) renderConsentPage(rw http.ResponseWriter, r *http.Request, consentRequired *oauthhandler.ConsentRequired) {
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	data := map[string]interface{}{}
	viewmodels.Embed(data, baseViewModel)

	// fixme(consent): show current user email
	viewModel := FromWebAppViewModel{}
	viewModel.IsRequestingFullUserInfo = slice.ContainsString(consentRequired.Scopes, oauth.FullUserInfoScope)
	viewModel.ClientName = consentRequired.Client.ClientName
	viewmodels.Embed(data, viewModel)

	h.Renderer.RenderHTML(rw, r, webapp.TemplateWebConsentHTML, data)
}
