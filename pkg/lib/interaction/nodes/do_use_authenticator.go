package nodes

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoUseAuthenticator{})
}

type InputCreateDeviceToken interface {
	CreateDeviceToken() bool
}

type EdgeDoUseAuthenticator struct {
	Stage         authn.AuthenticationStage
	Authenticator *authenticator.Info
}

func (e *EdgeDoUseAuthenticator) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	n := &NodeDoUseAuthenticator{
		Stage:         e.Stage,
		Authenticator: e.Authenticator,
	}

	userID := graph.MustGetUserID()
	var input InputCreateDeviceToken
	if interaction.Input(rawInput, &input) {
		if input.CreateDeviceToken() {
			token := ctx.MFA.GenerateDeviceToken()
			_, err := ctx.MFA.CreateDeviceToken(userID, token)
			if err != nil {
				return nil, err
			}
			cookie := ctx.CookieManager.ValueCookie(ctx.MFADeviceTokenCookie.Def, token)
			n.DeviceTokenCookie = cookie
		}
	}

	return n, nil
}

type NodeDoUseAuthenticator struct {
	Stage             authn.AuthenticationStage `json:"stage"`
	Authenticator     *authenticator.Info       `json:"authenticator"`
	DeviceTokenCookie *http.Cookie              `json:"device_token_cookie"`
}

// GetCookies implements CookiesGetter
func (n *NodeDoUseAuthenticator) GetCookies() []*http.Cookie {
	if n.DeviceTokenCookie == nil {
		return nil
	}
	return []*http.Cookie{n.DeviceTokenCookie}
}

func (n *NodeDoUseAuthenticator) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoUseAuthenticator) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeDoUseAuthenticator) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoUseAuthenticator) UserAuthenticator(stage authn.AuthenticationStage) (*authenticator.Info, bool) {
	if n.Stage == stage && n.Authenticator != nil {
		return n.Authenticator, true
	}
	return nil, false
}

func (n *NodeDoUseAuthenticator) UsedAuthenticationLockoutMethod() (config.AuthenticationLockoutMethod, bool) {
	if n.Authenticator != nil {
		return config.AuthenticationLockoutMethodFromAuthenticatorType(n.Authenticator.Type)
	}
	return "", false
}
