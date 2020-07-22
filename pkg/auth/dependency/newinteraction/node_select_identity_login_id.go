package newinteraction

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

type InputSelectIdentityLoginID interface {
	GetLoginID() string
}

type EdgeSelectIdentityLoginID struct {
	Config config.LoginIDKeyConfig
}

func (s *EdgeSelectIdentityLoginID) Instantiate(ctx *Context, graph *Graph, rawInput interface{}) (Node, error) {
	input, ok := rawInput.(InputSelectIdentityLoginID)
	if !ok {
		return nil, ErrIncompatibleInput
	}

	return &NodeSelectIdentityLoginID{
		Config:  s.Config,
		LoginID: input.GetLoginID(),
	}, nil
}

type NodeSelectIdentityLoginID struct {
	Config  config.LoginIDKeyConfig `json:"login_id_config"`
	LoginID string                  `json:"login_id"`
}

func (n *NodeSelectIdentityLoginID) Apply(ctx *Context, graph *Graph) error {
	return nil
}

func (n *NodeSelectIdentityLoginID) DeriveEdges(ctx *Context, graph *Graph) ([]Edge, error) {
	_, i, err := ctx.Identities.GetByClaims(
		authn.IdentityTypeLoginID,
		map[string]interface{}{
			identity.IdentityClaimLoginIDValue: n.LoginID,
		},
	)
	if errors.Is(err, identity.ErrIdentityNotFound) {
		i = nil
	} else if err != nil {
		return nil, err
	}

	return []Edge{
		&EdgeSelectIdentityEnd{Identity: i},
	}, nil
}
