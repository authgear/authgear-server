package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationEnd{})
}

type AuthenticationType string

const (
	AuthenticationTypeNone         AuthenticationType = "none"
	AuthenticationTypePassword     AuthenticationType = "password"
	AuthenticationTypeOTP          AuthenticationType = "otp"
	AuthenticationTypeRecoveryCode AuthenticationType = "recovery_code"
	AuthenticationTypeDeviceToken  AuthenticationType = "device_token"
)

type EdgeAuthenticationEnd struct {
	Stage                 interaction.AuthenticationStage
	AuthenticationType    AuthenticationType
	VerifiedAuthenticator *authenticator.Info
	RecoveryCode          *mfa.RecoveryCode
}

func (e *EdgeAuthenticationEnd) Instantiate(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	return &NodeAuthenticationEnd{
		Stage:                 e.Stage,
		AuthenticationType:    e.AuthenticationType,
		VerifiedAuthenticator: e.VerifiedAuthenticator,
		RecoveryCode:          e.RecoveryCode,
	}, nil
}

type NodeAuthenticationEnd struct {
	Stage                 interaction.AuthenticationStage `json:"stage"`
	AuthenticationType    AuthenticationType              `json:"authentication_type"`
	VerifiedAuthenticator *authenticator.Info             `json:"verified_authenticator"`
	RecoveryCode          *mfa.RecoveryCode               `json:"recovery_code"`
}

func (n *NodeAuthenticationEnd) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationEnd) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeAuthenticationEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	switch n.AuthenticationType {
	case AuthenticationTypeNone:
		break
	case AuthenticationTypePassword:
		if n.VerifiedAuthenticator == nil {
			return nil, interaction.ErrInvalidCredentials
		}
	case AuthenticationTypeOTP:
		if n.VerifiedAuthenticator == nil {
			return nil, interaction.ErrInvalidCredentials
		}
	case AuthenticationTypeRecoveryCode:
		if n.RecoveryCode == nil {
			return nil, interaction.ErrInvalidCredentials
		}
	case AuthenticationTypeDeviceToken:
		break
	default:
		panic("interaction: unknown authentication type: " + n.AuthenticationType)
	}

	return graph.Intent.DeriveEdgesForNode(graph, n)
}
