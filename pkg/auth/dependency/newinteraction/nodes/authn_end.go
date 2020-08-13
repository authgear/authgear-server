package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/mfa"
)

func init() {
	newinteraction.RegisterNode(&NodeAuthenticationEnd{})
}

type AuthenticationResult string

const (
	// AuthenticationResultRequired is zero value so by default authentication is performed by an authenticator.
	AuthenticationResultRequired AuthenticationResult = ""
	// AuthenticationResultRecoveryCode means the authentication is performed by a recovery code.
	AuthenticationResultRecoveryCode AuthenticationResult = "recovery_code"
	// AuthenticationResultOptional means the authentication is optional.
	// For example, OAuth identity does not require authenticator.
	AuthenticationResultOptional AuthenticationResult = "optional"
	// AuthenticationResultDeviceToken means the authentication is performed by a device token.
	AuthenticationResultDeviceToken AuthenticationResult = "device_token"
)

type EdgeAuthenticationEnd struct {
	Stage                 newinteraction.AuthenticationStage
	Result                AuthenticationResult
	VerifiedAuthenticator *authenticator.Info
	RecoveryCode          *mfa.RecoveryCode
}

func (e *EdgeAuthenticationEnd) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeAuthenticationEnd{
		Stage:                 e.Stage,
		Result:                e.Result,
		VerifiedAuthenticator: e.VerifiedAuthenticator,
		RecoveryCode:          e.RecoveryCode,
	}, nil
}

type NodeAuthenticationEnd struct {
	Stage                 newinteraction.AuthenticationStage `json:"stage"`
	Result                AuthenticationResult               `json:"result"`
	VerifiedAuthenticator *authenticator.Info                `json:"verified_authenticator"`
	RecoveryCode          *mfa.RecoveryCode                  `json:"recovery_code"`
}

func (n *NodeAuthenticationEnd) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeAuthenticationEnd) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeAuthenticationEnd) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	switch n.Result {
	case AuthenticationResultRequired:
		if n.VerifiedAuthenticator == nil {
			return nil, newinteraction.ErrInvalidCredentials
		}
	case AuthenticationResultRecoveryCode:
		if n.RecoveryCode == nil {
			return nil, newinteraction.ErrInvalidCredentials
		}
	case AuthenticationResultOptional:
		break
	case AuthenticationResultDeviceToken:
		break
	default:
		panic("interaction: unknown authentication result: " + n.Result)
	}

	return graph.Intent.DeriveEdgesForNode(graph, n)
}
