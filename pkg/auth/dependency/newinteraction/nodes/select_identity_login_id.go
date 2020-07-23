package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

type InputSelectIdentityLoginID interface {
	GetLoginID() string
}

type EdgeSelectIdentityLoginID struct {
	Config config.LoginIDKeyConfig
}

func (e *EdgeSelectIdentityLoginID) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputSelectIdentityLoginID)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	return &NodeSelectIdentityLoginID{
		Config:  e.Config,
		LoginID: input.GetLoginID(),
	}, nil
}

type NodeSelectIdentityLoginID struct {
	Config  config.LoginIDKeyConfig `json:"login_id_config"`
	LoginID string                  `json:"login_id"`
}

func (n *NodeSelectIdentityLoginID) Apply(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeSelectIdentityLoginID) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	spec := &identity.Spec{
		Type: authn.IdentityTypeLoginID,
		Claims: map[string]interface{}{
			identity.IdentityClaimLoginIDKey:   n.Config.Key,
			identity.IdentityClaimLoginIDValue: n.LoginID,
		},
	}

	_, info, err := ctx.Identities.GetByClaims(spec.Type, spec.Claims)
	if errors.Is(err, identity.ErrIdentityNotFound) {
		info = nil
	} else if err != nil {
		return nil, err
	}

	return []newinteraction.Edge{
		&EdgeSelectIdentityEnd{RequestedIdentity: spec, ExistingIdentity: info},
	}, nil
}
