package nodes

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func init() {
	newinteraction.RegisterNode(&NodeDoUseAuthenticator{})
}

type InputCreateDeviceToken interface {
	CreateDeviceToken() bool
}

type EdgeDoUseAuthenticator struct {
	Stage         newinteraction.AuthenticationStage
	Authenticator *authenticator.Info
}

func (e *EdgeDoUseAuthenticator) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	n := &NodeDoUseAuthenticator{
		Stage:         e.Stage,
		Authenticator: e.Authenticator,
	}

	userID := graph.MustGetUserID()
	if input, ok := rawInput.(InputCreateDeviceToken); ok {
		if input.CreateDeviceToken() {
			token := ctx.MFA.GenerateDeviceToken()
			_, err := ctx.MFA.CreateDeviceToken(userID, token)
			if err != nil {
				return nil, err
			}
			cookie := ctx.CookieFactory.ValueCookie(ctx.MFADeviceTokenCookie.Def, token)
			n.DeviceTokenCookie = cookie
		}
	}

	return n, nil
}

type NodeDoUseAuthenticator struct {
	Stage             newinteraction.AuthenticationStage `json:"stage"`
	Authenticator     *authenticator.Info                `json:"authenticator"`
	DeviceTokenCookie *http.Cookie                       `json:"device_token_cookie"`
}

// GetCookies implements CookiesGetter
func (n *NodeDoUseAuthenticator) GetCookies() []*http.Cookie {
	return []*http.Cookie{n.DeviceTokenCookie}
}

func (n *NodeDoUseAuthenticator) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeDoUseAuthenticator) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeDoUseAuthenticator) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoUseAuthenticator) UserAuthenticator(stage newinteraction.AuthenticationStage) (*authenticator.Info, bool) {
	if n.Stage == stage && n.Authenticator != nil {
		return n.Authenticator, true
	}
	return nil, false
}
