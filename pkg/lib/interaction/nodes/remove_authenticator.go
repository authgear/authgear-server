package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeRemoveAuthenticator{})
}

type InputRemoveAuthenticator interface {
	GetAuthenticatorType() authn.AuthenticatorType
	GetAuthenticatorID() string
}

type EdgeRemoveAuthenticator struct{}

func (e *EdgeRemoveAuthenticator) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputRemoveAuthenticator
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	userID := graph.MustGetUserID()
	authenticatorType := input.GetAuthenticatorType()
	authenticatorID := input.GetAuthenticatorID()
	bypassMFARequirement := false
	var bypassInput interface{ BypassMFARequirement() bool }
	if interaction.Input(rawInput, &bypassInput) {
		bypassMFARequirement = bypassInput.BypassMFARequirement()
	}

	info, err := ctx.Authenticators.Get(userID, authenticatorType, authenticatorID)
	if err != nil {
		return nil, err
	}

	return &NodeRemoveAuthenticator{
		AuthenticatorInfo:    info,
		BypassMFARequirement: bypassMFARequirement,
	}, nil
}

type NodeRemoveAuthenticator struct {
	AuthenticatorInfo    *authenticator.Info `json:"authenticator_info"`
	BypassMFARequirement bool                `json:"bypass_mfa_requirement"`
}

func (n *NodeRemoveAuthenticator) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeRemoveAuthenticator) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeRemoveAuthenticator) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
