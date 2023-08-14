package workflowconfig

import (
	"context"
	"net/http"

	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeDoClearDeviceTokenCookie{})
}

type NodeDoClearDeviceTokenCookie struct {
	Cookie *http.Cookie `json:"cookie,omitempty"`
}

var _ workflow.NodeSimple = &NodeDoClearDeviceTokenCookie{}
var _ workflow.CookieGetter = &NodeDoClearDeviceTokenCookie{}

func (*NodeDoClearDeviceTokenCookie) Kind() string {
	return "workflowconfig.NodeDoClearDeviceTokenCookie"
}

func (n *NodeDoClearDeviceTokenCookie) GetCookies(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]*http.Cookie, error) {
	return []*http.Cookie{n.Cookie}, nil
}
