package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeSelectIdentityLoginID{})
}

type InputSelectIdentityLoginID interface {
	GetLoginIDKey() string
	GetLoginID() string
}

type EdgeSelectIdentityLoginID struct {
	Configs []config.LoginIDKeyConfig
}

// GetIdentityCandidates implements IdentityCandidatesGetter.
func (e *EdgeSelectIdentityLoginID) GetIdentityCandidates() []identity.Candidate {
	candidates := make([]identity.Candidate, len(e.Configs))
	for i, c := range e.Configs {
		conf := c
		candidates[i] = identity.NewLoginIDCandidate(&conf)
	}
	return candidates
}

func (e *EdgeSelectIdentityLoginID) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputSelectIdentityLoginID)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}
	spec := &identity.Spec{
		Type: authn.IdentityTypeLoginID,
		Claims: map[string]interface{}{
			identity.IdentityClaimLoginIDValue: input.GetLoginID(),
		},
	}
	if key := input.GetLoginIDKey(); key != "" {
		spec.Claims[identity.IdentityClaimLoginIDKey] = key
	}

	info, err := ctx.Identities.GetBySpec(spec)
	if errors.Is(err, identity.ErrIdentityNotFound) {
		info = nil
	} else if err != nil {
		return nil, err
	}

	return &NodeSelectIdentityLoginID{RequestedIdentity: spec, ExistingIdentity: info}, nil
}

type NodeSelectIdentityLoginID struct {
	RequestedIdentity *identity.Spec `json:"requested_identity"`
	ExistingIdentity  *identity.Info `json:"existing_identity"`
}

func (n *NodeSelectIdentityLoginID) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeSelectIdentityLoginID) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{
		&EdgeSelectIdentityEnd{RequestedIdentity: n.RequestedIdentity, ExistingIdentity: n.ExistingIdentity},
	}, nil
}
