package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCreateAuthenticatorWhatsappFallbackSMS{})
}

type InputCreateAuthenticatorWhatsappFallbackSMS interface {
	FallbackSMS()
}

type EdgeCreateAuthenticatorWhatsappFallbackSMS struct {
	Stage     authn.AuthenticationStage
	IsDefault bool
}

func (e *EdgeCreateAuthenticatorWhatsappFallbackSMS) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputCreateAuthenticatorWhatsappFallbackSMS
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	return &NodeCreateAuthenticatorWhatsappFallbackSMS{
		Stage:     e.Stage,
		IsDefault: e.IsDefault,
	}, nil
}

type NodeCreateAuthenticatorWhatsappFallbackSMS struct {
	Stage     authn.AuthenticationStage `json:"stage"`
	IsDefault bool                      `json:"is_default"`
}

func (n *NodeCreateAuthenticatorWhatsappFallbackSMS) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorWhatsappFallbackSMS) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorWhatsappFallbackSMS) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeCreateAuthenticatorOOBSetup{
			Stage:                n.Stage,
			IsDefault:            n.IsDefault,
			OOBAuthenticatorType: model.AuthenticatorTypeOOBSMS,
		},
	}, nil
}
