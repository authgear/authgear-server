package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationWhatsappFallbackSMS{})
}

type InputAuthenticationWhatsappFallbackSMS interface {
	FallbackSMS()
}

type EdgeAuthenticationWhatsappFallbackSMS struct {
	Stage          authn.AuthenticationStage
	Authenticators []*authenticator.Info
}

func (e *EdgeAuthenticationWhatsappFallbackSMS) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputAuthenticationWhatsappFallbackSMS
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	return &NodeAuthenticationWhatsappFallbackSMS{
		Stage:          e.Stage,
		Authenticators: e.Authenticators,
	}, nil
}

type NodeAuthenticationWhatsappFallbackSMS struct {
	Stage          authn.AuthenticationStage `json:"stage"`
	Authenticators []*authenticator.Info     `json:"authenticators"`
}

func (n *NodeAuthenticationWhatsappFallbackSMS) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationWhatsappFallbackSMS) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeAuthenticationWhatsappFallbackSMS) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeAuthenticationOOBTrigger{
			Stage:                n.Stage,
			OOBAuthenticatorType: model.AuthenticatorTypeOOBSMS,
			Authenticators:       n.Authenticators,
		},
	}, nil
}
