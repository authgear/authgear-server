package workflowconfig

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDoCreateDeviceToken{})
}

type NodeDoCreateDeviceToken struct {
	UserID string       `json:"user_id"`
	Cookie *http.Cookie `json:"cookie,omitempty"`
}

var _ workflow.CookieGetter = &NodeDoCreateDeviceToken{}

func (n *NodeDoCreateDeviceToken) GetCookies(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]*http.Cookie, error) {
	return []*http.Cookie{n.Cookie}, nil
}

var _ workflow.NodeSimple = &NodeDoCreateDeviceToken{}

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

func (n *NodeDoCreateDeviceToken) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*NodeDoCreateDeviceToken) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoCreateDeviceToken) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeDoCreateDeviceToken) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
