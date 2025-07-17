package oauth

import (
	"context"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	oauthhandler "github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var errConsentRequiredError = errors.New("consent required")

type Renderer interface {
	RenderHTML(w http.ResponseWriter, r *http.Request, tpl *template.HTML, data interface{})
}

func ConfigureConsentRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST").
		WithPathPattern("/oauth2/consent")
}

var ConsentHandlerLogger = slogutil.NewLogger("handler-from-webapp")

type ProtocolConsentHandler interface {
	HandleConsentWithoutUserConsent(ctx context.Context, req *http.Request) (httputil.Result, *oauthhandler.ConsentRequired)
	HandleConsentWithUserConsent(ctx context.Context, req *http.Request) httputil.Result
	HandleConsentWithUserCancel(ctx context.Context, req *http.Request) httputil.Result
}

type ProtocolIdentityService interface {
	ListByUser(ctx context.Context, userID string) ([]*identity.Info, error)
}

type ConsentUserService interface {
	Get(ctx context.Context, userID string, role accesscontrol.Role) (*model.User, error)
}

type ConsentViewModel struct {
	ClientName          string
	ClientPolicyURI     string
	ClientTOSURI        string
	Scopes              []string
	IdentityDisplayName string
	UserProfile         webapp.UserProfile
}

type ConsentHandler struct {
	Database      *appdb.Handle
	Handler       ProtocolConsentHandler
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	Identities    ProtocolIdentityService
	Users         ConsentUserService
}

func (h *ConsentHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
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
		err = h.Database.WithTx(r.Context(), func(ctx context.Context) error {
			result, consentRequired = h.Handler.HandleConsentWithoutUserConsent(ctx, r)
			if consentRequired != nil {
				err = h.renderConsentPage(ctx, rw, r, consentRequired)
				if err != nil {
					return err
				}
				// return error to rollback transaction
				return errConsentRequiredError
			}
			if result.IsInternalError() {
				return errAuthzInternalError
			}
			return nil
		})
		if err != nil && errors.Is(err, errConsentRequiredError) {
			return
		}
	case http.MethodPost:
		if r.Form.Get("x_action") == "consent" {
			err = h.Database.WithTx(r.Context(), func(ctx context.Context) error {
				result = h.Handler.HandleConsentWithUserConsent(ctx, r)
				if result.IsInternalError() {
					return errAuthzInternalError
				}
				return nil
			})
			break
		} else if r.Form.Get("x_action") == "cancel" {
			err = h.Database.WithTx(r.Context(), func(ctx context.Context) error {
				result = h.Handler.HandleConsentWithUserCancel(ctx, r)
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
		logger := ConsentHandlerLogger.GetLogger(r.Context())
		logger.WithError(err).Error(r.Context(), "oauth consent handler failed")
		http.Error(rw, "Internal Server Error", 500)
	}
}

func (h *ConsentHandler) renderConsentPage(ctx context.Context, rw http.ResponseWriter, r *http.Request, consentRequired *oauthhandler.ConsentRequired) error {
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	data := map[string]interface{}{}
	viewmodels.Embed(data, baseViewModel)

	identities, err := h.Identities.ListByUser(ctx, consentRequired.UserID)
	if err != nil {
		return err
	}
	user, err := h.Users.Get(ctx, consentRequired.UserID, accesscontrol.RoleGreatest)
	if err != nil {
		return err
	}

	displayID := webapp.IdentitiesDisplayName(identities)
	userProfile := webapp.GetUserProfile(user)

	viewModel := ConsentViewModel{}
	viewModel.Scopes = consentRequired.Scopes
	viewModel.ClientName = consentRequired.Client.ClientName
	viewModel.ClientPolicyURI = consentRequired.Client.PolicyURI
	viewModel.ClientTOSURI = consentRequired.Client.TOSURI
	viewModel.IdentityDisplayName = displayID
	viewModel.UserProfile = userProfile
	viewmodels.Embed(data, viewModel)

	h.Renderer.RenderHTML(rw, r, webapp.TemplateWebConsentHTML, data)
	return nil
}
