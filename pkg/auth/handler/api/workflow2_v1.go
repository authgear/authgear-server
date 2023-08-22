package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type Workflow2V1WorkflowService interface {
	CreateNewWorkflow(intent workflow.Intent, sessionOptions *workflow.SessionOptions) (*workflow.ServiceOutput, error)
	Get(instanceID string, userAgentID string) (*workflow.ServiceOutput, error)
	FeedInput(instanceID string, userAgentID string, rawMessage json.RawMessage) (*workflow.ServiceOutput, error)
}

type Workflow2V1CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ClearCookie(def *httputil.CookieDef) *http.Cookie
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
}

type Workflow2V1OAuthSessionService interface {
	Get(entryID string) (*oauthsession.Entry, error)
}

type Workflow2V1UIInfoResolver interface {
	ResolveForUI(r protocol.AuthorizationRequest) (*oidc.UIInfo, error)
}

func workflow2getOrCreateUserAgentID(cookies Workflow2V1CookieManager, w http.ResponseWriter, r *http.Request) string {
	var userAgentID string
	userAgentIDCookie, err := cookies.GetCookie(r, workflow.UserAgentIDCookieDef)
	if err == nil {
		userAgentID = userAgentIDCookie.Value
	}
	if userAgentID == "" {
		userAgentID = workflow.NewUserAgentID()
	}
	cookie := cookies.ValueCookie(workflow.UserAgentIDCookieDef, userAgentID)
	httputil.UpdateCookie(w, cookie)
	return userAgentID
}

func ConfigureWorkflow2V1Routes(route httproute.Route) []httproute.Route {
	return []httproute.Route{
		route.WithMethods("OPTIONS", "POST").WithPathPattern("/api/v1/workflow2s"),
		route.WithMethods("OPTIONS", "GET", "POST").WithPathPattern("/api/v1/workflow2s/:slug"),
	}
}

var Workflow2V1RestfulCreateRequestSchema = validation.NewSimpleSchema(`
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

type Workflow2V1RestfulCreateRequest struct {
	FlowReference *workflow.FlowReference `json:"flow_reference,omitempty"`
	BindUserAgent *bool                   `json:"bind_user_agent,omitempty"`
	BatchInput    []json.RawMessage       `json:"batch_input,omitempty"`
}

func (r *Workflow2V1RestfulCreateRequest) SetDefaults() {
	if r.BindUserAgent == nil {
		defaultBindUserAgent := true
		r.BindUserAgent = &defaultBindUserAgent
	}
}

func (r Workflow2V1RestfulCreateRequest) ToNonRestful(httpReq *http.Request) Workflow2V1NonRestfulCreateRequest {
	rawQuery := httpReq.URL.RawQuery
	return Workflow2V1NonRestfulCreateRequest{
		FlowReference: r.FlowReference,
		BindUserAgent: r.BindUserAgent,
		BatchInput:    r.BatchInput,
		URLQuery:      rawQuery,
	}
}

var Workflow2V1NonRestfulCreateRequestSchema = validation.NewSimpleSchema(`
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

type Workflow2V1NonRestfulCreateRequest struct {
	FlowReference *workflow.FlowReference `json:"flow_reference,omitempty"`
	URLQuery      string                  `json:"url_query,omitempty"`
	BindUserAgent *bool                   `json:"bind_user_agent,omitempty"`
	BatchInput    []json.RawMessage       `json:"batch_input,omitempty"`
}

func (r *Workflow2V1NonRestfulCreateRequest) SetDefaults() {
	if r.BindUserAgent == nil {
		defaultBindUserAgent := true
		r.BindUserAgent = &defaultBindUserAgent
	}
}

var Workflow2V1NonRestfulGetRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"id": { "type": "string" }
		},
		"required": ["id"]
	}
`)

type Workflow2V1NonRestfulGetRequest struct {
	ID string `json:"id,omitempty"`
}

var Workflow2V1NonRestfulInputRequestSchema = validation.NewSimpleSchema(`
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

type Workflow2V1NonRestfulInputRequest struct {
	ID         string            `json:"id,omitempty"`
	Input      json.RawMessage   `json:"input,omitempty"`
	BatchInput []json.RawMessage `json:"batch_input,omitempty"`
}

var Workflow2V1RestfulInputRequestSchema = validation.NewSimpleSchema(`
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

type Workflow2V1RestfulInputRequest struct {
	Input      json.RawMessage   `json:"input,omitempty"`
	BatchInput []json.RawMessage `json:"batch_input,omitempty"`
}

func (r Workflow2V1RestfulInputRequest) ToNonRestful(instanceID string) Workflow2V1NonRestfulInputRequest {
	return Workflow2V1NonRestfulInputRequest{
		ID:         instanceID,
		Input:      r.Input,
		BatchInput: r.BatchInput,
	}
}

type Workflow2V1Handler struct {
	JSON           JSONResponseWriter
	Cookies        Workflow2V1CookieManager
	Workflows      Workflow2V1WorkflowService
	OAuthSessions  Workflow2V1OAuthSessionService
	UIInfoResolver Workflow2V1UIInfoResolver
}

func (h *Workflow2V1Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slug := httproute.GetParam(r, "slug")
	switch r.Method {
	case "GET":
		// RESTful get
		instanceID := slug
		h.get(w, r, instanceID)
	case "POST":
		switch slug {
		case "get":
			// Non-RESTful get
			var err error
			var request Workflow2V1NonRestfulGetRequest
			err = httputil.BindJSONBody(r, w, Workflow2V1NonRestfulGetRequestSchema.Validator(), &request)
			if err != nil {
				h.JSON.WriteResponse(w, &api.Response{Error: err})
				return
			}

			instanceID := request.ID
			h.get(w, r, instanceID)
		case "":
			// RESTful create
			var err error
			var request Workflow2V1RestfulCreateRequest
			err = httputil.BindJSONBody(r, w, Workflow2V1RestfulCreateRequestSchema.Validator(), &request)
			if err != nil {
				h.JSON.WriteResponse(w, &api.Response{Error: err})
				return
			}

			h.create(w, r, request.ToNonRestful(r))
		case "create":
			// Non-RESTful create
			var err error
			var request Workflow2V1NonRestfulCreateRequest
			err = httputil.BindJSONBody(r, w, Workflow2V1NonRestfulCreateRequestSchema.Validator(), &request)
			if err != nil {
				h.JSON.WriteResponse(w, &api.Response{Error: err})
				return
			}

			h.create(w, r, request)
		case "input":
			// Non-RESTful input
			var err error
			var request Workflow2V1NonRestfulInputRequest
			err = httputil.BindJSONBody(r, w, Workflow2V1NonRestfulInputRequestSchema.Validator(), &request)
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
			var request Workflow2V1RestfulInputRequest
			err = httputil.BindJSONBody(r, w, Workflow2V1RestfulInputRequestSchema.Validator(), &request)
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

func (h *Workflow2V1Handler) makeSessionOptionsFromQuery(urlQuery string) *workflow.SessionOptions {
	q, _ := url.ParseQuery(urlQuery)
	return &workflow.SessionOptions{
		ClientID:  q.Get("client_id"),
		State:     q.Get("state"),
		XState:    q.Get("x_state"),
		UILocales: q.Get("ui_locales"),
	}
}

func (h *Workflow2V1Handler) makeSessionOptionsFromCookie(oauthSessionCookie *http.Cookie) (*workflow.SessionOptions, error) {
	entry, err := h.OAuthSessions.Get(oauthSessionCookie.Value)
	if err != nil {
		return nil, err
	}
	req := entry.T.AuthorizationRequest

	uiInfo, err := h.UIInfoResolver.ResolveForUI(req)
	if err != nil {
		return nil, err
	}

	sessionOptions := &workflow.SessionOptions{
		ClientID:                 uiInfo.ClientID,
		RedirectURI:              uiInfo.RedirectURI,
		SuppressIDPSessionCookie: uiInfo.SuppressIDPSessionCookie,
		State:                    uiInfo.State,
		XState:                   uiInfo.XState,
		UILocales:                req.UILocalesRaw(),
	}

	return sessionOptions, nil
}

func (h *Workflow2V1Handler) get(w http.ResponseWriter, r *http.Request, instanceID string) {
	userAgentID := workflow2getOrCreateUserAgentID(h.Cookies, w, r)

	output, err := h.Workflows.Get(instanceID, userAgentID)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	result := workflow.FlowResponse{
		ID:         output.Workflow.InstanceID,
		Data:       output.Data,
		JSONSchema: output.SchemaBuilder,
		Finished:   output.Finished,
	}
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}

func (h *Workflow2V1Handler) create(w http.ResponseWriter, r *http.Request, request Workflow2V1NonRestfulCreateRequest) {
	output, err := h.create0(w, r, request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	if len(request.BatchInput) > 0 {
		instanceID := output.Workflow.InstanceID
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

	result := workflow.FlowResponse{
		ID:         output.Workflow.InstanceID,
		Data:       output.Data,
		Finished:   output.Finished,
		JSONSchema: output.SchemaBuilder,
	}
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}

func (h *Workflow2V1Handler) create0(w http.ResponseWriter, r *http.Request, request Workflow2V1NonRestfulCreateRequest) (*workflow.ServiceOutput, error) {
	flow, err := workflow.InstantiateFlow(*request.FlowReference)
	if err != nil {
		return nil, err
	}

	userAgentID := workflow2getOrCreateUserAgentID(h.Cookies, w, r)

	var sessionOptionsFromCookie *workflow.SessionOptions
	oauthCookie, err := h.Cookies.GetCookie(r, oauthsession.UICookieDef)
	if err == nil {
		sessionOptionsFromCookie, err = h.makeSessionOptionsFromCookie(oauthCookie)
		if errors.Is(err, oauthsession.ErrNotFound) {
			// Clear the cookie if it invalid or expired
			httputil.UpdateCookie(w, h.Cookies.ClearCookie(oauthsession.UICookieDef))
		} else if err != nil {
			// Still return error for any other errors.
			return nil, err
		}

		// Do not clear the UI cookie so that a new session can be created again.
		// httputil.UpdateCookie(w, h.Cookies.ClearCookie(oauthsession.UICookieDef))
	}

	// Accept client_id, state, ui_locales from query.
	// This is essential if the templates of some features require these paramenters.
	sessionOptionsFromQuery := h.makeSessionOptionsFromQuery(request.URLQuery)

	// The query overrides the cookie.
	sessionOptions := sessionOptionsFromCookie.PartiallyMergeFrom(sessionOptionsFromQuery)

	if *request.BindUserAgent {
		sessionOptions.UserAgentID = userAgentID
	}

	output, err := h.Workflows.CreateNewWorkflow(flow, sessionOptions)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (h *Workflow2V1Handler) input(w http.ResponseWriter, r *http.Request, request Workflow2V1NonRestfulInputRequest) {
	instanceID := request.ID
	userAgentID := workflow2getOrCreateUserAgentID(h.Cookies, w, r)

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

	result := workflow.FlowResponse{
		ID:         output.Workflow.InstanceID,
		Data:       output.Data,
		JSONSchema: output.SchemaBuilder,
		Finished:   output.Finished,
	}
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}

func (h *Workflow2V1Handler) input0(
	w http.ResponseWriter,
	r *http.Request,
	instanceID string,
	userAgentID string,
	request Workflow2V1NonRestfulInputRequest,
) (*workflow.ServiceOutput, error) {
	output, err := h.Workflows.FeedInput(instanceID, userAgentID, request.Input)
	if err != nil && !errors.Is(err, workflow.ErrEOF) {
		return nil, err
	}

	return output, nil
}

func (h *Workflow2V1Handler) batchInput(w http.ResponseWriter, r *http.Request, request Workflow2V1NonRestfulInputRequest) {
	instanceID := request.ID
	userAgentID := workflow2getOrCreateUserAgentID(h.Cookies, w, r)

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

	result := workflow.FlowResponse{
		ID:         output.Workflow.InstanceID,
		Data:       output.Data,
		JSONSchema: output.SchemaBuilder,
		Finished:   output.Finished,
	}
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}

func (h *Workflow2V1Handler) batchInput0(
	w http.ResponseWriter,
	r *http.Request,
	instanceID string,
	userAgentID string,
	rawMessages []json.RawMessage,
) (output *workflow.ServiceOutput, err error) {
	// Collect all cookies
	var cookies []*http.Cookie
	for _, rawMessage := range rawMessages {
		output, err = h.Workflows.FeedInput(instanceID, userAgentID, rawMessage)
		if err != nil && !errors.Is(err, workflow.ErrEOF) {
			return nil, err
		}

		// Feed the next input to the latest instance.
		instanceID = output.Workflow.InstanceID
		cookies = append(cookies, output.Cookies...)
	}
	if err != nil && errors.Is(err, workflow.ErrEOF) {
		err = nil
	}
	if err != nil {
		return
	}

	// Return all collected cookies.
	output.Cookies = cookies
	return
}

func (h *Workflow2V1Handler) prepareErrorResponse(
	instanceID string,
	userAgentID string,
	workflowErr error,
) (*api.Response, error) {
	output, err := h.Workflows.Get(instanceID, userAgentID)
	if err != nil {
		return nil, err
	}

	result := workflow.FlowResponse{
		ID:         output.Workflow.InstanceID,
		Data:       output.Data,
		Finished:   output.Finished,
		JSONSchema: output.SchemaBuilder,
	}
	return &api.Response{
		Error:  workflowErr,
		Result: result,
	}, nil
}
