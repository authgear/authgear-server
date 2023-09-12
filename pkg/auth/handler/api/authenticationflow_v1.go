package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/iawaknahc/originmatcher"

	"github.com/authgear/authgear-server/pkg/api"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/pubsub"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type AuthenticationFlowV1WorkflowService interface {
	CreateNewFlow(intent authflow.PublicFlow, sessionOptions *authflow.SessionOptions) (*authflow.ServiceOutput, error)
	Get(instanceID string, userAgentID string) (*authflow.ServiceOutput, error)
	FeedInput(instanceID string, userAgentID string, rawMessage json.RawMessage) (*authflow.ServiceOutput, error)
}

type AuthenticationFlowV1CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ClearCookie(def *httputil.CookieDef) *http.Cookie
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
}

type AuthenticationFlowV1OAuthSessionService interface {
	Get(entryID string) (*oauthsession.Entry, error)
}

type AuthenticationFlowV1UIInfoResolver interface {
	GetOAuthSessionID(req *http.Request, urlQuery string) (string, bool)
	RemoveOAuthSessionID(w http.ResponseWriter, r *http.Request)
	ResolveForUI(r protocol.AuthorizationRequest) (*oidc.UIInfo, error)
}

type AuthenticationFlowV1WebsocketEventStore interface {
	ChannelName(authenticationFlowID string) (string, error)
}

type AuthenticationFlowV1WebsocketOriginMatcher interface {
	PrepareOriginMatcher(r *http.Request) (*originmatcher.T, error)
}

func authenticationFlowGetOrCreateUserAgentID(cookies AuthenticationFlowV1CookieManager, w http.ResponseWriter, r *http.Request) string {
	var userAgentID string
	userAgentIDCookie, err := cookies.GetCookie(r, authflow.UserAgentIDCookieDef)
	if err == nil {
		userAgentID = userAgentIDCookie.Value
	}
	if userAgentID == "" {
		userAgentID = authflow.NewUserAgentID()
	}
	cookie := cookies.ValueCookie(authflow.UserAgentIDCookieDef, userAgentID)
	httputil.UpdateCookie(w, cookie)
	return userAgentID
}

func ConfigureAuthenticationFlowV1Routes(route httproute.Route) []httproute.Route {
	return []httproute.Route{
		route.WithMethods("OPTIONS", "POST").WithPathPattern("/api/v1/authentication_flows"),
		route.WithMethods("OPTIONS", "GET", "POST").WithPathPattern("/api/v1/authentication_flows/:slug"),
	}
}

var AuthenticationFlowV1RestfulCreateRequestSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"required": ["flow_reference"],
	"properties": {
		"flow_reference": {
			"type": "object",
			"properties": {
				"type": {
					"type": "string",
					"enum": ["signup_flow", "login_flow"]
				},
				"id": {
					"type": "string"
				}
			},
			"required": ["type", "id"]
		},
		"bind_user_agent": { "type": "boolean" },
		"batch_input": {
			"type": "array",
			"items": {
				"type": "object"
			}
		}
	}
}
`)

type AuthenticationFlowV1RestfulCreateRequest struct {
	FlowReference *authflow.FlowReference `json:"flow_reference,omitempty"`
	BindUserAgent *bool                   `json:"bind_user_agent,omitempty"`
	BatchInput    []json.RawMessage       `json:"batch_input,omitempty"`
}

func (r *AuthenticationFlowV1RestfulCreateRequest) SetDefaults() {
	if r.BindUserAgent == nil {
		defaultBindUserAgent := true
		r.BindUserAgent = &defaultBindUserAgent
	}
}

func (r AuthenticationFlowV1RestfulCreateRequest) ToNonRestful(httpReq *http.Request) AuthenticationFlowV1NonRestfulCreateRequest {
	rawQuery := httpReq.URL.RawQuery
	return AuthenticationFlowV1NonRestfulCreateRequest{
		FlowReference: r.FlowReference,
		BindUserAgent: r.BindUserAgent,
		BatchInput:    r.BatchInput,
		URLQuery:      rawQuery,
	}
}

var AuthenticationFlowV1NonRestfulCreateRequestSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"required": ["flow_reference"],
	"properties": {
		"flow_reference": {
			"type": "object",
			"properties": {
				"type": {
					"type": "string",
					"enum": ["signup_flow", "login_flow"]
				},
				"id": {
					"type": "string"
				}
			},
			"required": ["type", "id"]
		},
		"url_query": { "type": "string" },
		"bind_user_agent": { "type": "boolean" },
		"batch_input": {
			"type": "array",
			"items": {
				"type": "object"
			}
		}
	}
}
`)

type AuthenticationFlowV1NonRestfulCreateRequest struct {
	FlowReference *authflow.FlowReference `json:"flow_reference,omitempty"`
	URLQuery      string                  `json:"url_query,omitempty"`
	BindUserAgent *bool                   `json:"bind_user_agent,omitempty"`
	BatchInput    []json.RawMessage       `json:"batch_input,omitempty"`
}

func (r *AuthenticationFlowV1NonRestfulCreateRequest) SetDefaults() {
	if r.BindUserAgent == nil {
		defaultBindUserAgent := true
		r.BindUserAgent = &defaultBindUserAgent
	}
}

var AuthenticationFlowV1NonRestfulGetRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"id": { "type": "string" }
		},
		"required": ["id"]
	}
`)

type AuthenticationFlowV1NonRestfulGetRequest struct {
	ID string `json:"id,omitempty"`
}

var AuthenticationFlowV1NonRestfulInputRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"required": ["id"],
		"properties": {
			"id": { "type": "string" }
		},
		"oneOf": [
			{
				"properties": {
					"input": {
						"type": "object"
					}
				},
				"required": ["input"]
			},
			{
				"properties": {
					"batch_input": {
						"type": "array",
						"items": {
							"type": "object"
						},
						"minItems": 1
					}
				},
				"required": ["batch_input"]
			}
		]
	}
`)

type AuthenticationFlowV1NonRestfulInputRequest struct {
	ID         string            `json:"id,omitempty"`
	Input      json.RawMessage   `json:"input,omitempty"`
	BatchInput []json.RawMessage `json:"batch_input,omitempty"`
}

var AuthenticationFlowV1RestfulInputRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"oneOf": [
			{
				"properties": {
					"input": {
						"type": "object"
					}
				},
				"required": ["input"]
			},
			{
				"properties": {
					"batch_input": {
						"type": "array",
						"items": {
							"type": "object"
						},
						"minItems": 1
					}
				},
				"required": ["batch_input"]
			}
		]
	}
`)

type AuthenticationFlowV1RestfulInputRequest struct {
	Input      json.RawMessage   `json:"input,omitempty"`
	BatchInput []json.RawMessage `json:"batch_input,omitempty"`
}

func (r AuthenticationFlowV1RestfulInputRequest) ToNonRestful(instanceID string) AuthenticationFlowV1NonRestfulInputRequest {
	return AuthenticationFlowV1NonRestfulInputRequest{
		ID:         instanceID,
		Input:      r.Input,
		BatchInput: r.BatchInput,
	}
}

type AuthenticationFlowV1Handler struct {
	LoggerFactory  *log.Factory
	RedisHandle    *appredis.Handle
	JSON           JSONResponseWriter
	Cookies        AuthenticationFlowV1CookieManager
	Workflows      AuthenticationFlowV1WorkflowService
	OAuthSessions  AuthenticationFlowV1OAuthSessionService
	UIInfoResolver AuthenticationFlowV1UIInfoResolver
	OriginMatcher  AuthenticationFlowV1WebsocketOriginMatcher
	Events         AuthenticationFlowV1WebsocketEventStore
}

func (h *AuthenticationFlowV1Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slug := httproute.GetParam(r, "slug")
	switch r.Method {
	case "GET":
		switch slug {
		case "ws":
			// websocket
			h.websocket(w, r)
		default:
			// RESTful get
			instanceID := slug
			h.get(w, r, instanceID)
		}
	case "POST":
		switch slug {
		case "get":
			// Non-RESTful get
			var err error
			var request AuthenticationFlowV1NonRestfulGetRequest
			err = httputil.BindJSONBody(r, w, AuthenticationFlowV1NonRestfulGetRequestSchema.Validator(), &request)
			if err != nil {
				h.JSON.WriteResponse(w, &api.Response{Error: err})
				return
			}

			instanceID := request.ID
			h.get(w, r, instanceID)
		case "":
			// RESTful create
			var err error
			var request AuthenticationFlowV1RestfulCreateRequest
			err = httputil.BindJSONBody(r, w, AuthenticationFlowV1RestfulCreateRequestSchema.Validator(), &request)
			if err != nil {
				h.JSON.WriteResponse(w, &api.Response{Error: err})
				return
			}

			h.create(w, r, request.ToNonRestful(r))
		case "create":
			// Non-RESTful create
			var err error
			var request AuthenticationFlowV1NonRestfulCreateRequest
			err = httputil.BindJSONBody(r, w, AuthenticationFlowV1NonRestfulCreateRequestSchema.Validator(), &request)
			if err != nil {
				h.JSON.WriteResponse(w, &api.Response{Error: err})
				return
			}

			h.create(w, r, request)
		case "input":
			// Non-RESTful input
			var err error
			var request AuthenticationFlowV1NonRestfulInputRequest
			err = httputil.BindJSONBody(r, w, AuthenticationFlowV1NonRestfulInputRequestSchema.Validator(), &request)
			if err != nil {
				h.JSON.WriteResponse(w, &api.Response{Error: err})
				return
			}

			if request.Input != nil {
				h.input(w, r, request)
			} else {
				h.batchInput(w, r, request)
			}
		default:
			// RESTful input
			instanceID := slug

			var err error
			var request AuthenticationFlowV1RestfulInputRequest
			err = httputil.BindJSONBody(r, w, AuthenticationFlowV1RestfulInputRequestSchema.Validator(), &request)
			if err != nil {
				h.JSON.WriteResponse(w, &api.Response{Error: err})
				return
			}

			if request.Input != nil {
				h.input(w, r, request.ToNonRestful(instanceID))
			} else {
				h.batchInput(w, r, request.ToNonRestful(instanceID))
			}
		}
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AuthenticationFlowV1Handler) makeSessionOptionsFromQuery(urlQuery string) *authflow.SessionOptions {
	q, _ := url.ParseQuery(urlQuery)
	return &authflow.SessionOptions{
		ClientID:  q.Get("client_id"),
		State:     q.Get("state"),
		XState:    q.Get("x_state"),
		UILocales: q.Get("ui_locales"),
	}
}

func (h *AuthenticationFlowV1Handler) makeSessionOptionsFromOAuth(oauthSessionID string) (*authflow.SessionOptions, error) {
	entry, err := h.OAuthSessions.Get(oauthSessionID)
	if err != nil {
		return nil, err
	}
	req := entry.T.AuthorizationRequest

	uiInfo, err := h.UIInfoResolver.ResolveForUI(req)
	if err != nil {
		return nil, err
	}

	sessionOptions := &authflow.SessionOptions{
		OAuthSessionID:           oauthSessionID,
		ClientID:                 uiInfo.ClientID,
		RedirectURI:              uiInfo.RedirectURI,
		SuppressIDPSessionCookie: uiInfo.SuppressIDPSessionCookie,
		State:                    uiInfo.State,
		XState:                   uiInfo.XState,
		UILocales:                req.UILocalesRaw(),
	}

	return sessionOptions, nil
}

func (h *AuthenticationFlowV1Handler) get(w http.ResponseWriter, r *http.Request, instanceID string) {
	userAgentID := authenticationFlowGetOrCreateUserAgentID(h.Cookies, w, r)

	output, err := h.Workflows.Get(instanceID, userAgentID)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	result := output.ToFlowResponse()
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}

func (h *AuthenticationFlowV1Handler) create(w http.ResponseWriter, r *http.Request, request AuthenticationFlowV1NonRestfulCreateRequest) {
	output, err := h.create0(w, r, request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	if len(request.BatchInput) > 0 {
		instanceID := output.Flow.InstanceID
		userAgentID := output.Session.UserAgentID

		output, err = h.batchInput0(w, r, instanceID, userAgentID, request.BatchInput)
		if err != nil {
			apiResp, apiRespErr := h.prepareErrorResponse(instanceID, userAgentID, err)
			if apiRespErr != nil {
				// failed to get the workflow when preparing the error response
				h.JSON.WriteResponse(w, &api.Response{Error: apiRespErr})
				return
			}
			h.JSON.WriteResponse(w, apiResp)
			return
		}
	}

	for _, c := range output.Cookies {
		httputil.UpdateCookie(w, c)
	}

	result := output.ToFlowResponse()
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}

func (h *AuthenticationFlowV1Handler) create0(w http.ResponseWriter, r *http.Request, request AuthenticationFlowV1NonRestfulCreateRequest) (*authflow.ServiceOutput, error) {
	flow, err := authflow.InstantiateFlow(*request.FlowReference)
	if err != nil {
		return nil, err
	}

	userAgentID := authenticationFlowGetOrCreateUserAgentID(h.Cookies, w, r)

	var sessionOptionsFromOAuth *authflow.SessionOptions
	if oauthSessionID, ok := h.UIInfoResolver.GetOAuthSessionID(r, request.URLQuery); ok {
		sessionOptionsFromOAuth, err = h.makeSessionOptionsFromOAuth(oauthSessionID)
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

	if *request.BindUserAgent {
		sessionOptions.UserAgentID = userAgentID
	}

	output, err := h.Workflows.CreateNewFlow(flow, sessionOptions)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (h *AuthenticationFlowV1Handler) input(w http.ResponseWriter, r *http.Request, request AuthenticationFlowV1NonRestfulInputRequest) {
	instanceID := request.ID
	userAgentID := authenticationFlowGetOrCreateUserAgentID(h.Cookies, w, r)

	output, err := h.input0(w, r, instanceID, userAgentID, request)
	if err != nil {
		apiResp, apiRespErr := h.prepareErrorResponse(instanceID, userAgentID, err)
		if apiRespErr != nil {
			// failed to get the workflow when preparing the error response
			h.JSON.WriteResponse(w, &api.Response{Error: apiRespErr})
			return
		}
		h.JSON.WriteResponse(w, apiResp)
		return
	}

	for _, c := range output.Cookies {
		httputil.UpdateCookie(w, c)
	}

	result := output.ToFlowResponse()
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}

func (h *AuthenticationFlowV1Handler) input0(
	w http.ResponseWriter,
	r *http.Request,
	instanceID string,
	userAgentID string,
	request AuthenticationFlowV1NonRestfulInputRequest,
) (*authflow.ServiceOutput, error) {
	output, err := h.Workflows.FeedInput(instanceID, userAgentID, request.Input)
	if err != nil && !errors.Is(err, authflow.ErrEOF) {
		return nil, err
	}

	return output, nil
}

func (h *AuthenticationFlowV1Handler) batchInput(w http.ResponseWriter, r *http.Request, request AuthenticationFlowV1NonRestfulInputRequest) {
	instanceID := request.ID
	userAgentID := authenticationFlowGetOrCreateUserAgentID(h.Cookies, w, r)

	output, err := h.batchInput0(w, r, instanceID, userAgentID, request.BatchInput)
	if err != nil {
		apiResp, apiRespErr := h.prepareErrorResponse(instanceID, userAgentID, err)
		if apiRespErr != nil {
			// failed to get the workflow when preparing the error response
			h.JSON.WriteResponse(w, &api.Response{Error: apiRespErr})
			return
		}
		h.JSON.WriteResponse(w, apiResp)
		return
	}

	for _, c := range output.Cookies {
		httputil.UpdateCookie(w, c)
	}

	result := output.ToFlowResponse()
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}

func (h *AuthenticationFlowV1Handler) batchInput0(
	w http.ResponseWriter,
	r *http.Request,
	instanceID string,
	userAgentID string,
	rawMessages []json.RawMessage,
) (output *authflow.ServiceOutput, err error) {
	// Collect all cookies
	var cookies []*http.Cookie
	for _, rawMessage := range rawMessages {
		output, err = h.Workflows.FeedInput(instanceID, userAgentID, rawMessage)
		if err != nil && !errors.Is(err, authflow.ErrEOF) {
			return nil, err
		}

		// Feed the next input to the latest instance.
		instanceID = output.Flow.InstanceID
		cookies = append(cookies, output.Cookies...)
	}
	if err != nil && errors.Is(err, authflow.ErrEOF) {
		err = nil
	}
	if err != nil {
		return
	}

	// Return all collected cookies.
	output.Cookies = cookies
	return
}

func (h *AuthenticationFlowV1Handler) prepareErrorResponse(
	instanceID string,
	userAgentID string,
	workflowErr error,
) (*api.Response, error) {
	output, err := h.Workflows.Get(instanceID, userAgentID)
	if err != nil {
		return nil, err
	}

	result := output.ToFlowResponse()
	return &api.Response{
		Error:  workflowErr,
		Result: result,
	}, nil
}

func (h *AuthenticationFlowV1Handler) websocket(w http.ResponseWriter, r *http.Request) {
	matcher, err := h.OriginMatcher.PrepareOriginMatcher(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	handler := &pubsub.HTTPHandler{
		RedisHub:      h.RedisHandle,
		Delegate:      h,
		LoggerFactory: h.LoggerFactory,
		OriginMatcher: matcher,
	}

	handler.ServeHTTP(w, r)
}

func (h *AuthenticationFlowV1Handler) Accept(r *http.Request) (string, error) {
	websocketID := r.FormValue("websocket_id")
	return h.Events.ChannelName(websocketID)
}

func (h *AuthenticationFlowV1Handler) OnRedisSubscribe(r *http.Request) error {
	return nil
}
