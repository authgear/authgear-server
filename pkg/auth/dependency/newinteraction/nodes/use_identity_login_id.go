package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeUseIdentityLoginID{})
}

type InputUseIdentityLoginID interface {
	GetLoginIDKey() string
	GetLoginID() string
}

type EdgeUseIdentityLoginID struct {
	IsCreating bool
	Configs    []config.LoginIDKeyConfig
}

// GetIdentityCandidates implements IdentityCandidatesGetter.
func (e *EdgeUseIdentityLoginID) GetIdentityCandidates() []identity.Candidate {
	candidates := make([]identity.Candidate, len(e.Configs))
	for i, c := range e.Configs {
		conf := c
		candidates[i] = identity.NewLoginIDCandidate(&conf)
	}
	return candidates
}

func (e *EdgeUseIdentityLoginID) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputUseIdentityLoginID)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	loginIDKey := input.GetLoginIDKey()
	loginID := input.GetLoginID()
	claims := map[string]interface{}{
		identity.IdentityClaimLoginIDKey:   loginIDKey,
		identity.IdentityClaimLoginIDValue: loginID,
	}
	spec := &identity.Spec{
		Type:   authn.IdentityTypeLoginID,
		Claims: claims,
	}

	return &NodeUseIdentityLoginID{
		IsCreating:   e.IsCreating,
		IdentitySpec: spec,
	}, nil

}

type NodeUseIdentityLoginID struct {
	IsCreating   bool           `json:"is_creating"`
	IdentitySpec *identity.Spec `json:"identity_spec"`
}

func (n *NodeUseIdentityLoginID) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeUseIdentityLoginID) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	if n.IsCreating {
		return []newinteraction.Edge{&EdgeCreateIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
	}
	return []newinteraction.Edge{&EdgeSelectIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
}
