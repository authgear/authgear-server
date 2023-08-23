package declarative

import (
	"context"
	"net/http"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeDoClearDeviceTokenCookie{})
}

type NodeDoClearDeviceTokenCookie struct {
	Cookie *http.Cookie `json:"cookie,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoClearDeviceTokenCookie{}
var _ authflow.CookieGetter = &NodeDoClearDeviceTokenCookie{}

func (*NodeDoClearDeviceTokenCookie) Kind() string {
	return "NodeDoClearDeviceTokenCookie"
}

func (n *NodeDoClearDeviceTokenCookie) GetCookies(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]*http.Cookie, error) {
	return []*http.Cookie{n.Cookie}, nil
}
