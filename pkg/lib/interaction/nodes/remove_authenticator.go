package nodes

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeRemoveAuthenticator{})
}

type InputRemoveAuthenticator interface {
	GetAuthenticatorType() model.AuthenticatorType
	GetAuthenticatorID() string
}

type EdgeRemoveAuthenticator struct{}

func (e *EdgeRemoveAuthenticator) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputRemoveAuthenticator
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	authenticatorID := input.GetAuthenticatorID()
	bypassMFARequirement := false
	var bypassInput interface{ BypassMFARequirement() bool }
	if interaction.Input(rawInput, &bypassInput) {
		bypassMFARequirement = bypassInput.BypassMFARequirement()
	}

	info, err := ctx.Authenticators.Get(authenticatorID)
	if err != nil {
		return nil, err
	}

	if info.UserID != graph.MustGetUserID() {
		return nil, fmt.Errorf("authenticator does not belong to the user")
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
