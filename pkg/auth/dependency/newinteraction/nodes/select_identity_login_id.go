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

func (s *EdgeSelectIdentityLoginID) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputSelectIdentityLoginID)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
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

func (n *NodeSelectIdentityLoginID) Apply(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeSelectIdentityLoginID) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
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

	return []newinteraction.Edge{
		&EdgeSelectIdentityEnd{Identity: i},
	}, nil
}
