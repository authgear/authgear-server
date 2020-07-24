package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeCreateIdentityLoginID{})
}

type InputCreateIdentityLoginID interface {
	GetLoginID() *loginid.LoginID
}

type EdgeCreateIdentityLoginID struct {
	RequestedIdentity *identity.Spec
	Config            config.LoginIDKeyConfig
}

func (e *EdgeCreateIdentityLoginID) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	var claims map[string]interface{}
	if e.RequestedIdentity != nil {
		claims = e.RequestedIdentity.Claims
	} else {
		input, ok := rawInput.(InputCreateIdentityLoginID)
		if !ok {
			return nil, newinteraction.ErrIncompatibleInput
		}

		loginID := input.GetLoginID()
		if loginID.Key != e.Config.Key {
			return nil, newinteraction.ErrIncompatibleInput
		}

		claims = map[string]interface{}{
			identity.IdentityClaimLoginIDKey:   loginID.Key,
			identity.IdentityClaimLoginIDValue: loginID.Value,
		}
	}

	newIdentity, err := ctx.Identities.New(graph.MustGetUserID(), authn.IdentityTypeLoginID, claims)
	if err != nil {
		return nil, err
	}

	return &NodeCreateIdentityLoginID{Identity: newIdentity}, nil
}

type NodeCreateIdentityLoginID struct {
	Identity *identity.Info `json:"identity"`
}

func (n *NodeCreateIdentityLoginID) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateIdentityLoginID) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{&EdgeCreateIdentityEnd{NewIdentity: n.Identity}}, nil
}
