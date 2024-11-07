package declarative

import (
	"context"
	"net/http"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeDoCreateDeviceToken{})
}

type NodeDoCreateDeviceToken struct {
	UserID string       `json:"user_id"`
	Cookie *http.Cookie `json:"cookie,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoCreateDeviceToken{}
var _ authflow.CookieGetter = &NodeDoCreateDeviceToken{}

func NewNodeDoCreateDeviceToken(ctx context.Context, deps *authflow.Dependencies, n *NodeDoCreateDeviceToken) (*NodeDoCreateDeviceToken, error) {
	token := deps.MFA.GenerateDeviceToken(ctx)
	_, err := deps.MFA.CreateDeviceToken(ctx, n.UserID, token)
	if err != nil {
		return nil, err
	}

	cookie := deps.Cookies.ValueCookie(deps.MFADeviceTokenCookie.Def, token)
	n.Cookie = cookie
	return n, nil
}

func (*NodeDoCreateDeviceToken) Kind() string {
	return "authflow.NodeDoCreateDeviceToken"
}

func (n *NodeDoCreateDeviceToken) GetCookies(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]*http.Cookie, error) {
	return []*http.Cookie{n.Cookie}, nil
}
