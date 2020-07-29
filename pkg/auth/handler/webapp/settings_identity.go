package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/intents"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
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

type SettingsIdentityHandler struct {
	ServerConfig  *config.ServerConfig
	Database      *db.Handle
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
	WebApp        WebAppService
}

// FIXME(webapp): implement input interface
type SettingsIdentityLinkOAuth struct {
	UserID           string
	ProviderAlias    string
	Action           string
	NonceSource      *http.Cookie
	ErrorRedirectURI string
}

// FIXME(webapp): implement input interface
type SettingsIdentityUnlinkOAuth struct {
	UserID        string
	ProviderAlias string
}

// FIXME(webapp): implement input interface
type SettingsIdentityAddUpdateRemoveLoginID struct {
	UserID           string
	LoginIDKey       string
	LoginIDType      string
	LoginIDInputType string
	OldLoginIDValue  string
}

func (h *SettingsIdentityHandler) GetData(r *http.Request, state *webapp.State, graph *newinteraction.Graph, edges []newinteraction.Edge) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	var anyError interface{}
	if state != nil {
		anyError = state.Error
	}

	baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
	// FIXME(webapp): derive AuthenticationViewModel with graph and edges
	authenticationViewModel := viewmodels.AuthenticationViewModel{}

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
	oldLoginIDValue := r.Form.Get("x_old_login_id_value")
	sess := auth.GetSession(r.Context())
	userID := sess.AuthnAttrs().UserID
	nonceSource, _ := r.Cookie(webapp.CSRFCookieName)

	if r.Method == "GET" {
		h.Database.WithTx(func() error {
			intent := &webapp.Intent{
				RedirectURI: redirectURI,
				// FIXME(webapp): IntentSettingsIdentity
				// This intent actually does not have any further nodes.
				// Only the edges are useful.
				Intent: intents.NewIntentLogin(),
			}
			state, graph, edges, err := h.WebApp.GetIntent(intent, StateID(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			data, err := h.GetData(r, state, graph, edges)
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
				// FIXME(webapp): IntentLinkOAuth
				Intent: intents.NewIntentLogin(),
			}
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = &SettingsIdentityLinkOAuth{
					ProviderAlias: providerAlias,
					// FIXME(webapp): Use constant
					Action:           "link",
					UserID:           userID,
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
				// FIXME(webapp): IntentUnlinkOAuth
				Intent: intents.NewIntentLogin(),
			}
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = &SettingsIdentityUnlinkOAuth{
					ProviderAlias: providerAlias,
					UserID:        userID,
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
			intent := &webapp.Intent{
				RedirectURI: redirectURI,
				// FIXME(webapp): IntentAddUpdateRemoveLoginID
				Intent: intents.NewIntentLogin(),
			}
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				input = &SettingsIdentityAddUpdateRemoveLoginID{
					UserID:           userID,
					LoginIDKey:       loginIDKey,
					LoginIDType:      loginIDType,
					LoginIDInputType: loginIDInputType,
					OldLoginIDValue:  oldLoginIDValue,
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
}
