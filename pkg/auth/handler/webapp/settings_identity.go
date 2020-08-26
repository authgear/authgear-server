package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

const (
	TemplateItemTypeAuthUISettingsIdentityHTML string = "auth_ui_settings_identity.html"
)

var TemplateAuthUISettingsIdentityHTML = template.T{
	Type:                    TemplateItemTypeAuthUISettingsIdentityHTML,
	IsHTML:                  true,
	TranslationTemplateType: TemplateItemTypeAuthUITranslationJSON,
	Defines:                 defines,
	ComponentTemplateTypes:  components,
}

func ConfigureSettingsIdentityRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/identity")
}

type SettingsIdentityViewModel struct {
	VerificationStatuses map[string]verification.Status
}

type SettingsIdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
	ListCandidates(userID string) ([]identity.Candidate, error)
}

type SettingsVerificationService interface {
	GetVerificationStatuses(is []*identity.Info) (map[string]verification.Status, error)
}

type SettingsIdentityHandler struct {
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
	Identities    SettingsIdentityService
	Verification  SettingsVerificationService
	CSRFCookie    webapp.CSRFCookieDef
}

type SettingsIdentityLinkOAuth struct {
	ProviderAlias    string
	NonceSource      *http.Cookie
	ErrorRedirectURI string
}

var _ nodes.InputUseIdentityOAuthProvider = &SettingsIdentityLinkOAuth{}

func (i *SettingsIdentityLinkOAuth) GetProviderAlias() string {
	return i.ProviderAlias
}

func (i *SettingsIdentityLinkOAuth) GetNonceSource() *http.Cookie {
	return i.NonceSource
}

func (i *SettingsIdentityLinkOAuth) GetErrorRedirectURI() string {
	return i.ErrorRedirectURI
}

type SettingsIdentityUnlinkOAuth struct {
	IdentityID string
}

func (i *SettingsIdentityUnlinkOAuth) GetIdentityType() authn.IdentityType {
	return authn.IdentityTypeOAuth
}

func (i *SettingsIdentityUnlinkOAuth) GetIdentityID() string {
	return i.IdentityID
}

func (h *SettingsIdentityHandler) GetData(r *http.Request, state *webapp.State) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	var anyError interface{}
	if state != nil {
		anyError = state.Error
	}

	baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
	userID := session.GetUserID(r.Context())
	candidates, err := h.Identities.ListCandidates(*userID)
	if err != nil {
		return nil, err
	}
	authenticationViewModel := viewmodels.NewAuthenticationViewModelWithCandidates(candidates)

	viewModel := SettingsIdentityViewModel{
		VerificationStatuses: map[string]verification.Status{},
	}
	identities, err := h.Identities.ListByUser(*userID)
	if err != nil {
		return nil, err
	}
	viewModel.VerificationStatuses, err = h.Verification.GetVerificationStatuses(identities)
	if err != nil {
		return nil, err
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, authenticationViewModel)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

func (h *SettingsIdentityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	redirectURI := httputil.HostRelative(r.URL).String()
	providerAlias := r.Form.Get("x_provider_alias")
	identityID := r.Form.Get("x_identity_id")
	userID := session.GetUserID(r.Context())
	nonceSource, _ := r.Cookie(h.CSRFCookie.Name)

	if r.Method == "GET" {
		err := h.Database.WithTx(func() error {
			state, err := h.WebApp.GetState(StateID(r))
			if err != nil {
				return err
			}

			data, err := h.GetData(r, state)
			if err != nil {
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUISettingsIdentityHTML, data)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	if r.Method == "POST" && r.Form.Get("x_action") == "link_oauth" {
		err := h.Database.WithTx(func() error {
			intent := &webapp.Intent{
				RedirectURI: redirectURI,
				Intent:      intents.NewIntentAddIdentity(*userID),
			}
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = &SettingsIdentityLinkOAuth{
					ProviderAlias:    providerAlias,
					NonceSource:      nonceSource,
					ErrorRedirectURI: httputil.HostRelative(r.URL).String(),
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

	if r.Method == "POST" && r.Form.Get("x_action") == "unlink_oauth" {
		err := h.Database.WithTx(func() error {
			intent := &webapp.Intent{
				RedirectURI: redirectURI,
				Intent:      intents.NewIntentRemoveIdentity(*userID),
			}
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = &SettingsIdentityUnlinkOAuth{
					IdentityID: identityID,
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

	if r.Method == "POST" && r.Form.Get("x_action") == "verify_login_id" {
		err := h.Database.WithTx(func() error {
			intent := &webapp.Intent{
				RedirectURI: redirectURI,
				KeepState:   true,
				Intent:      intents.NewIntentVerifyIdentity(*userID, authn.IdentityTypeLoginID, identityID),
			}
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = nil
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
