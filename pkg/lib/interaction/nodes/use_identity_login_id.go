package nodes

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeUseIdentityLoginID{})
}

type InputUseIdentityLoginID interface {
	GetLoginIDKey() string
	GetLoginID() string
}

type UseIdentityLoginIDMode string

const (
	UseIdentityLoginIDModeCreate = "create"
	UseIdentityLoginIDModeSelect = "select"
	UseIdentityLoginIDModeUpdate = "update"
)

type EdgeUseIdentityLoginID struct {
	Mode    UseIdentityLoginIDMode
	Configs []config.LoginIDKeyConfig
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

func (e *EdgeUseIdentityLoginID) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputUseIdentityLoginID
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	loginIDKey := input.GetLoginIDKey()
	loginID := input.GetLoginID()

	claims := map[string]interface{}{
		identity.IdentityClaimLoginIDKey:   loginIDKey,
		identity.IdentityClaimLoginIDValue: loginID,
	}

	// This node is used by signup and login.
	// In login, loginIDKey is empty so it is impossible to derive type.
	// In signup, loginIDKey is given explicitly, and it is required to include
	// IdentityClaimLoginIDType in the claims.
	var typ config.LoginIDKeyType
	if loginIDKey != "" {
		for _, cfg := range e.Configs {
			if cfg.Key == loginIDKey {
				typ = cfg.Type
			}
		}
		if typ == "" {
			return nil, interaction.NewInvariantViolated(
				"InvalidLoginIDKey",
				"invalid login ID key",
				nil,
			)
		}
	}

	if typ != "" {
		claims[identity.IdentityClaimLoginIDType] = string(typ)
	}

	spec := &identity.Spec{
		Type:   authn.IdentityTypeLoginID,
		Claims: claims,
	}

	return &NodeUseIdentityLoginID{
		Mode:         e.Mode,
		IdentitySpec: spec,
	}, nil
}

type NodeUseIdentityLoginID struct {
	Mode         UseIdentityLoginIDMode `json:"mode"`
	IdentitySpec *identity.Spec         `json:"identity_spec"`
}

func (n *NodeUseIdentityLoginID) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeUseIdentityLoginID) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeUseIdentityLoginID) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	switch n.Mode {
	case UseIdentityLoginIDModeCreate:
		return []interaction.Edge{&EdgeCreateIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
	case UseIdentityLoginIDModeSelect:
		return []interaction.Edge{&EdgeSelectIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
	case UseIdentityLoginIDModeUpdate:
		return []interaction.Edge{&EdgeUpdateIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
	default:
		panic(fmt.Errorf("interaction: unexpected use identity mode: %v", n.Mode))
	}
}
