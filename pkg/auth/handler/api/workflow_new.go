package api

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureWorkflowNewRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/api/v1/workflows")
}

var WorkflowNewRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"intent": {
				"type": "object",
				"properties": {
					"kind": { "type": "string" },
					"data": { "type": "object" }
				},
				"required": ["kind", "data"]
			}
		},
		"required": ["intent"]
	}
`)

type WorkflowNewRequest struct {
	Intent workflow.IntentJSON `json:"intent"`
}

type WorkflowNewWorkflowService interface {
	CreateNewWorkflow(intent workflow.Intent, sessionOptions *workflow.SessionOptions) (*workflow.ServiceOutput, error)
}

type WorkflowNewCookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type WorkflowNewOAuthSessionService interface {
	Get(entryID string) (*oauthsession.Entry, error)
}

type WorkflowNewUIInfoResolver interface {
	ResolveForUI(r protocol.AuthorizationRequest) (*oidc.UIInfo, error)
}

type WorkflowNewHandler struct {
	Database       *appdb.Handle
	JSON           JSONResponseWriter
	Cookies        WorkflowNewCookieManager
	Workflows      WorkflowNewWorkflowService
	OAuthSessions  WorkflowNewOAuthSessionService
	UIInfoResolver WorkflowNewUIInfoResolver
}

func (h *WorkflowNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var request WorkflowNewRequest
	err = httputil.BindJSONBody(r, w, WorkflowNewRequestSchema.Validator(), &request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	var output *workflow.ServiceOutput
	err = h.Database.WithTx(func() error {
		output, err = h.handle(w, r, request)
		return err
	})
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	result := WorkflowResponse{
		Action:   output.Action,
		Workflow: output.WorkflowOutput,
	}
	h.JSON.WriteResponse(w, &api.Response{Result: result})
}

func (h *WorkflowNewHandler) handle(w http.ResponseWriter, r *http.Request, request WorkflowNewRequest) (*workflow.ServiceOutput, error) {
	intent, err := workflow.InstantiateIntentFromPublicRegistry(request.Intent)
	if err != nil {
		return nil, err
	}

	var sessionOptions *workflow.SessionOptions
	cookie, err := h.Cookies.GetCookie(r, oauthsession.UICookieDef)
	if err == nil {
		sessionOptions, err = h.makeSessionOptions(cookie)
		if err != nil {
			// Clear the cookie if it invalid or expired
			httputil.UpdateCookie(w, h.Cookies.ClearCookie(oauthsession.UICookieDef))
		}

		// Do not clear the UI cookie so that a new session can be created again.
		// httputil.UpdateCookie(w, h.Cookies.ClearCookie(oauthsession.UICookieDef))
	}
	// Accept client_id, state, ui_locales from query. Override any values found in oauth session.
	// This is essential if the templates of some features require these paramenters.
	if sessionOptions == nil {
		sessionOptions = &workflow.SessionOptions{}
	}
	sessionOptions = h.patchSessionOptions(r, sessionOptions)

	output, err := h.Workflows.CreateNewWorkflow(intent, sessionOptions)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (h *WorkflowNewHandler) patchSessionOptions(
	r *http.Request, opts *workflow.SessionOptions) *workflow.SessionOptions {

	clientID := r.FormValue("client_id")
	state := r.FormValue("state")
	xState := r.FormValue("x_state")
	uiLocales := r.FormValue("ui_locales")

	if clientID != "" {
		opts.ClientID = clientID
	}
	if state != "" {
		opts.State = state
	}
	if xState != "" {
		opts.XState = xState
	}
	if uiLocales != "" {
		opts.UILocales = uiLocales
	}

	return opts
}

func (h *WorkflowNewHandler) makeSessionOptions(cookie *http.Cookie) (*workflow.SessionOptions, error) {
	entry, err := h.OAuthSessions.Get(cookie.Value)
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
