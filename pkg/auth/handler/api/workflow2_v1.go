package api

import (
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type Workflow2V1WorkflowService interface {
	CreateNewWorkflow(intent workflow.Intent, sessionOptions *workflow.SessionOptions) (*workflow.ServiceOutput, error)
	Get(workflowID string, instanceID string, userAgentID string) (*workflow.ServiceOutput, error)
	FeedInput(workflowID string, instanceID string, userAgentID string, rawMessage json.RawMessage) (*workflow.ServiceOutput, error)
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
