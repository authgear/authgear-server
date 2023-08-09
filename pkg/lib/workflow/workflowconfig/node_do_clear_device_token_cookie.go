package workflowconfig

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDoClearDeviceTokenCookie{})
}

type NodeDoClearDeviceTokenCookie struct {
	Cookie *http.Cookie `json:"cookie,omitempty"`
}

var _ workflow.CookieGetter = &NodeDoClearDeviceTokenCookie{}

func (n *NodeDoClearDeviceTokenCookie) GetCookies(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]*http.Cookie, error) {
	return []*http.Cookie{n.Cookie}, nil
}

var _ workflow.NodeSimple = &NodeDoClearDeviceTokenCookie{}

func (*NodeDoClearDeviceTokenCookie) Kind() string {
	return "workflowconfig.NodeDoClearDeviceTokenCookie"
}

func (*NodeDoClearDeviceTokenCookie) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
}

func (*NodeDoClearDeviceTokenCookie) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoClearDeviceTokenCookie) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (*NodeDoClearDeviceTokenCookie) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
