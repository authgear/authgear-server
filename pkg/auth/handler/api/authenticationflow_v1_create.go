package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureAuthenticationFlowV1CreateRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST").WithPathPattern("/api/v1/authentication_flows")
}

var AuthenticationFlowV1NonRestfulCreateRequestSchema *validation.SimpleSchema

func init() {
	b := validation.SchemaBuilder{}.Type(validation.TypeObject)
	b.Required("type", "name")

	b.Properties().Property("name", validation.SchemaBuilder{}.Type(validation.TypeString))
	b.Properties().Property("url_query", validation.SchemaBuilder{}.Type(validation.TypeString))
	b.Properties().Property("batch_input", validation.SchemaBuilder{}.
		Type(validation.TypeArray).
		Items(validation.SchemaBuilder{}.Type(validation.TypeObject)))
	b.Properties().Property("type", validation.SchemaBuilder{}.
		Type(validation.TypeString).
		Enum(slice.Cast[authflow.FlowType, interface{}](authflow.AllFlowTypes)...))

	AuthenticationFlowV1NonRestfulCreateRequestSchema = b.ToSimpleSchema()
}

type AuthenticationFlowV1NonRestfulCreateRequest struct {
	Type       authflow.FlowType `json:"type,omitempty"`
	Name       string            `json:"name,omitempty"`
	URLQuery   string            `json:"url_query,omitempty"`
	BatchInput []json.RawMessage `json:"batch_input,omitempty"`
}

func (r *AuthenticationFlowV1NonRestfulCreateRequest) GetFlowReference() *authflow.FlowReference {
	return &authflow.FlowReference{
		Type: r.Type,
		Name: r.Name,
	}
}

type AuthenticationFlowV1OAuthSessionService interface {
	Get(ctx context.Context, entryID string) (*oauthsession.Entry, error)
}

type AuthenticationFlowV1UIInfoResolver interface {
	GetOAuthSessionID(req *http.Request, urlQuery string) (string, bool)
	RemoveOAuthSessionID(w http.ResponseWriter, r *http.Request)
	ResolveForUI(ctx context.Context, r protocol.AuthorizationRequest) (*oidc.UIInfo, error)
}

type AuthenticationFlowV1CreateHandler struct {
	LoggerFactory  *log.Factory
	RedisHandle    *appredis.Handle
	Cookies        AuthenticationFlowV1CookieManager
	Workflows      AuthenticationFlowV1WorkflowService
	OAuthSessions  AuthenticationFlowV1OAuthSessionService
	UIInfoResolver AuthenticationFlowV1UIInfoResolver
}

func (h *AuthenticationFlowV1CreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var request AuthenticationFlowV1NonRestfulCreateRequest
	ctx := r.Context()
	err = httputil.BindJSONBody(r, w, AuthenticationFlowV1NonRestfulCreateRequestSchema.Validator(), &request)
	if err != nil {
		httputil.WriteJSONResponse(ctx, w, &api.Response{Error: err})
		return
	}

	h.create(ctx, w, r, request)
}

func (h *AuthenticationFlowV1CreateHandler) create(ctx context.Context, w http.ResponseWriter, r *http.Request, request AuthenticationFlowV1NonRestfulCreateRequest) {
	output, err := h.create0(ctx, w, r, request)
	if err != nil {
		httputil.WriteJSONResponse(ctx, w, &api.Response{Error: err})
		return
	}

	if len(request.BatchInput) > 0 {
		stateToken := output.Flow.StateToken

		output, err = batchInput0(ctx, h.Workflows, w, r, stateToken, request.BatchInput)
		if err != nil {
			apiResp, apiRespErr := prepareErrorResponse(ctx, h.Workflows, stateToken, err)
			if apiRespErr != nil {
				// failed to get the workflow when preparing the error response
				httputil.WriteJSONResponse(ctx, w, &api.Response{Error: apiRespErr})
				return
			}
			httputil.WriteJSONResponse(ctx, w, apiResp)
			return
		}
	}

	for _, c := range output.Cookies {
		httputil.UpdateCookie(w, c)
	}

	result := output.ToFlowResponse()
	httputil.WriteJSONResponse(ctx, w, &api.Response{Result: result})
}

func (h *AuthenticationFlowV1CreateHandler) create0(ctx context.Context, w http.ResponseWriter, r *http.Request, request AuthenticationFlowV1NonRestfulCreateRequest) (*authflow.ServiceOutput, error) {
	flow, err := authflow.InstantiateFlow(*request.GetFlowReference(), jsonpointer.T{})
	if err != nil {
		return nil, err
	}

	var sessionOptionsFromOAuth *authflow.SessionOptions
	if oauthSessionID, ok := h.UIInfoResolver.GetOAuthSessionID(r, request.URLQuery); ok {
		sessionOptionsFromOAuth, err = h.makeSessionOptionsFromOAuth(ctx, oauthSessionID)
		if errors.Is(err, oauthsession.ErrNotFound) {
			// Clear the oauth session if it invalid or expired
			h.UIInfoResolver.RemoveOAuthSessionID(w, r)
		} else if err != nil {
			// Still return error for any other errors.
			return nil, err
		}

		// Do not clear the oauth session so that a new session can be created again.
	}

	// Accept client_id, state, ui_locales from query.
	// This is essential if the templates of some features require these paramenters.
	sessionOptionsFromQuery := h.makeSessionOptionsFromQuery(request.URLQuery)

	// The query overrides the cookie.
	sessionOptions := sessionOptionsFromOAuth.PartiallyMergeFrom(sessionOptionsFromQuery)

	output, err := h.Workflows.CreateNewFlow(ctx, flow, sessionOptions)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (h *AuthenticationFlowV1CreateHandler) makeSessionOptionsFromQuery(urlQuery string) *authflow.SessionOptions {
	q, _ := url.ParseQuery(urlQuery)
	return &authflow.SessionOptions{
		ClientID:  q.Get("client_id"),
		State:     q.Get("state"),
		XState:    q.Get("x_state"),
		UILocales: q.Get("ui_locales"),
		LoginHint: q.Get("login_hint"),
	}
}

func (h *AuthenticationFlowV1CreateHandler) makeSessionOptionsFromOAuth(ctx context.Context, oauthSessionID string) (*authflow.SessionOptions, error) {
	entry, err := h.OAuthSessions.Get(ctx, oauthSessionID)
	if err != nil {
		return nil, err
	}
	req := entry.T.AuthorizationRequest

	uiInfo, err := h.UIInfoResolver.ResolveForUI(ctx, req)
	if err != nil {
		return nil, err
	}

	sessionOptions := &authflow.SessionOptions{
		OAuthSessionID: oauthSessionID,

		ClientID:    uiInfo.ClientID,
		RedirectURI: uiInfo.RedirectURI,
		Prompt:      uiInfo.Prompt,
		State:       uiInfo.State,
		XState:      uiInfo.XState,
		UILocales:   req.UILocalesRaw(),

		IDToken:                  uiInfo.IDTokenHint,
		SuppressIDPSessionCookie: uiInfo.SuppressIDPSessionCookie,
		UserIDHint:               uiInfo.UserIDHint,
		LoginHint:                uiInfo.LoginHint,
	}

	return sessionOptions, nil
}
