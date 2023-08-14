package workflowconfig

import (
	"context"
	"net/http"

	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeDoCreateDeviceToken{})
}

type NodeDoCreateDeviceToken struct {
	UserID string       `json:"user_id"`
	Cookie *http.Cookie `json:"cookie,omitempty"`
}

var _ workflow.NodeSimple = &NodeDoCreateDeviceToken{}
var _ workflow.CookieGetter = &NodeDoCreateDeviceToken{}

func NewNodeDoCreateDeviceToken(deps *workflow.Dependencies, n *NodeDoCreateDeviceToken) (*NodeDoCreateDeviceToken, error) {
	token := deps.MFA.GenerateDeviceToken()
	_, err := deps.MFA.CreateDeviceToken(n.UserID, token)
	if err != nil {
		return nil, err
	}

	cookie := deps.Cookies.ValueCookie(deps.MFADeviceTokenCookie.Def, token)
	n.Cookie = cookie
	return n, nil
}

func (*NodeDoCreateDeviceToken) Kind() string {
	return "workflow.NodeDoCreateDeviceToken"
}

func (n *NodeDoCreateDeviceToken) GetCookies(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]*http.Cookie, error) {
	return []*http.Cookie{n.Cookie}, nil
}
