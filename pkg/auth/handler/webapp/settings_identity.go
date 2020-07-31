package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/intents"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/httputil"
	"github.com/authgear/authgear-server/pkg/template"
)

const (
	TemplateItemTypeAuthUISettingsIdentityHTML config.TemplateItemType = "auth_ui_settings_identity.html"
)

var TemplateAuthUISettingsIdentityHTML = template.Spec{
	Type:        TemplateItemTypeAuthUISettingsIdentityHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
}

func ConfigureSettingsIdentityRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/identity")
}

type SettingsIdentityService interface {
	ListCandidates(userID string) ([]identity.Candidate, error)
}

type SettingsIdentityHandler struct {
	ServerConfig  *config.ServerConfig
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
	Identities    SettingsIdentityService
}

type SettingsIdentityLinkOAuth struct {
	ProviderAlias    string
	State            string
	NonceSource      *http.Cookie
	ErrorRedirectURI string
}

var _ nodes.InputUseIdentityOAuthProvider = &SettingsIdentityLinkOAuth{}

func (i *SettingsIdentityLinkOAuth) GetProviderAlias() string {
	return i.ProviderAlias
}

func (i *SettingsIdentityLinkOAuth) GetState() string {
	return i.State
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
	userID := auth.GetSession(r.Context()).AuthnAttrs().UserID
	candidates, err := h.Identities.ListCandidates(userID)
	if err != nil {
		return nil, err
	}
	authenticationViewModel := viewmodels.NewAuthenticationViewModelWithCandidates(candidates)

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, authenticationViewModel)

	return data, nil
}

func (h *SettingsIdentityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redirectURI := httputil.HostRelative(r.URL).String()
	providerAlias := r.Form.Get("x_provider_alias")
	loginIDKey := r.Form.Get("x_login_id_key")
	loginIDType := r.Form.Get("x_login_id_type")
	loginIDInputType := r.Form.Get("x_login_id_input_type")
	identityID := r.Form.Get("x_identity_id")
	sess := auth.GetSession(r.Context())
	userID := sess.AuthnAttrs().UserID
	nonceSource, _ := r.Cookie(webapp.CSRFCookieName)

	if r.Method == "GET" {
		h.Database.WithTx(func() error {
			state, err := h.WebApp.GetState(StateID(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			data, err := h.GetData(r, state)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			h.Renderer.Render(w, r, TemplateItemTypeAuthUISettingsIdentityHTML, data)
			return nil
		})
	}

	if r.Method == "POST" && r.Form.Get("x_action") == "link_oauth" {
		h.Database.WithTx(func() error {
			intent := &webapp.Intent{
				RedirectURI: redirectURI,
				Intent:      intents.NewIntentAddIdentity(userID),
			}
			stateID := webapp.NewID()
			intent.StateID = stateID
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = &SettingsIdentityLinkOAuth{
					ProviderAlias:    providerAlias,
					State:            stateID,
					NonceSource:      nonceSource,
					ErrorRedirectURI: httputil.HostRelative(r.URL).String(),
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

	if r.Method == "POST" && r.Form.Get("x_action") == "unlink_oauth" {
		h.Database.WithTx(func() error {
			intent := &webapp.Intent{
				RedirectURI: redirectURI,
				Intent:      intents.NewIntentRemoveIdentity(userID),
			}
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = &SettingsIdentityUnlinkOAuth{
					IdentityID: identityID,
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

	if r.Method == "POST" && r.Form.Get("x_action") == "login_id" {
		h.Database.WithTx(func() error {
			var intent *webapp.Intent
			if identityID == "" {
				intent = &webapp.Intent{
					RedirectURI: redirectURI,
					Intent:      intents.NewIntentAddIdentity(userID),
					StateExtra: map[string]interface{}{
						"x_login_id_key":        loginIDKey,
						"x_login_id_type":       loginIDType,
						"x_login_id_input_type": loginIDInputType,
						"x_identity_id":         identityID,
					},
				}
			} else {
				intent = &webapp.Intent{
					RedirectURI: redirectURI,
					Intent:      intents.NewIntentUpdateIdentity(userID, identityID),
					StateExtra: map[string]interface{}{
						"x_login_id_key":        loginIDKey,
						"x_login_id_type":       loginIDType,
						"x_login_id_input_type": loginIDInputType,
						"x_identity_id":         identityID,
					},
				}
			}

			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = struct{}{}
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
