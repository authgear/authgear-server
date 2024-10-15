package authflowv2

import (
	"net/http"
	"net/url"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsIdentityListOAuthHTML = template.RegisterHTML(
	"web/authflowv2/settings_identity_list_oauth.html",
	handlerwebapp.SettingsComponents...,
)

func ConfigureAuthflowV2SettingsIdentityListOAuthRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityListOAuth)
}

type AuthflowV2SettingsIdentityListOAuthViewModel struct {
	OAuthCandidates []identity.Candidate
	OAuthIdentities []*identity.OAuth
	Verifications   map[string][]verification.ClaimStatus
	IdentityCount   int
	CreateDisabled  bool
}

type AuthflowV2SettingsIdentityListOAuthHandler struct {
	Database          *appdb.Handle
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          handlerwebapp.Renderer
	Identities        *identityservice.Service
	Verification      handlerwebapp.SettingsVerificationService
	Endpoints         handlerwebapp.SettingsEndpointsProvider
	OAuthStateStore   handlerwebapp.SettingsOAuthStateStore
	AccountManagement accountmanagement.Service
}

func (h *AuthflowV2SettingsIdentityListOAuthHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	userID := session.GetUserID(r.Context())

	candidates, err := h.Identities.ListCandidates(*userID)
	if err != nil {
		return nil, err
	}

	var oauthCandidates []identity.Candidate
	for _, candidate := range candidates {
		typ, _ := candidate[identity.CandidateKeyType].(string)
		if typ == string(model.IdentityTypeOAuth) {
			oauthCandidates = append(oauthCandidates, candidate)
		}
	}

	identities, err := h.Identities.ListByUser(*userID)
	if err != nil {
		return nil, err
	}

	remaining := identity.ApplyFilters(
		identities,
		identity.KeepIdentifiable,
	)

	var oauthIdentities []*identity.OAuth
	var oauthInfos []*identity.Info
	for _, identity := range identities {
		if identity.Type != model.IdentityTypeOAuth {
			continue
		}

		oauthIdentities = append(oauthIdentities, identity.OAuth)
		oauthInfos = append(oauthInfos, identity.OAuth.ToInfo())
	}

	sort.Slice(oauthIdentities, func(i, j int) bool {
		return oauthIdentities[i].UpdatedAt.Before(oauthIdentities[j].UpdatedAt)
	})

	verifications, err := h.Verification.GetVerificationStatuses(oauthInfos)
	if err != nil {
		return nil, err
	}
	vm := AuthflowV2SettingsIdentityListOAuthViewModel{
		OAuthCandidates: oauthCandidates,
		OAuthIdentities: oauthIdentities,
		Verifications:   verifications,
		IdentityCount:   len(remaining),
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func generateAuthorizationURLWithState(authorizationURLString string, stateToken string) (string, error) {
	authorizationURL, err := url.Parse(authorizationURLString)
	if err != nil {
		return "", err
	}

	q := authorizationURL.Query()
	q.Set("state", stateToken)

	authorizationURL.RawQuery = q.Encode()

	return authorizationURL.String(), nil
}

func (h *AuthflowV2SettingsIdentityListOAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx()

	ctrl.Get(func() error {
		var data map[string]interface{}
		err := h.Database.WithTx(func() error {
			data, err = h.GetData(r, w)
			return err
		})
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsIdentityListOAuthHTML, data)
		return nil
	})

	ctrl.PostAction("add", func() error {
		s := session.GetSession(r.Context())

		alias := r.Form.Get("x_provider_alias")
		redirectURI := h.Endpoints.SSOCallbackURL(alias).String()

		output, err := h.AccountManagement.StartAddIdentityOAuth(s, &accountmanagement.StartAddIdentityOAuthInput{
			Alias:       alias,
			RedirectURI: redirectURI,
		})
		if err != nil {
			return err
		}

		state := &webappoauth.WebappOAuthState{
			AccountManagementToken: output.Token,
		}
		stateToken, err := h.OAuthStateStore.GenerateState(state)
		if err != nil {
			return err
		}

		authorizationURLString, err := generateAuthorizationURLWithState(output.AuthorizationURL, stateToken)
		if err != nil {
			return err
		}

		http.Redirect(w, r, authorizationURLString, http.StatusFound)

		return nil
	})

	ctrl.PostAction("remove", func() error {
		s := session.GetSession(r.Context())

		identityID := r.Form.Get("q_identity_id")

		_, err := h.AccountManagement.DeleteIdentityOAuth(s, &accountmanagement.DeleteIdentityOAuthInput{
			IdentityID: identityID,
		})
		if err != nil {
			return err
		}

		redirectURI := httputil.HostRelative(r.URL).String()
		result := webapp.Result{RedirectURI: redirectURI}
		result.WriteResponse(w, r)

		return nil
	})
}
