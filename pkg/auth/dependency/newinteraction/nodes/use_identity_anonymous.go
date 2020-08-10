package nodes

import (
	"encoding/json"
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/dependency/challenge"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeUseIdentityAnonymous{})
}

type InputUseIdentityAnonymous interface {
	GetAnonymousRequestToken() string
}

type EdgeUseIdentityAnonymous struct {
	IsCreating bool
}

func (e *EdgeUseIdentityAnonymous) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputUseIdentityAnonymous)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	enabled := false
	for _, t := range ctx.Config.Authentication.Identities {
		if t == authn.IdentityTypeAnonymous {
			enabled = true
			break
		}
	}

	if !enabled {
		return nil, newinteraction.ConfigurationViolated.New("anonymous users are not allowed")
	}

	jwt := input.GetAnonymousRequestToken()

	request, err := ctx.AnonymousIdentities.ParseRequestUnverified(jwt)
	if err != nil {
		return nil, newinteraction.ErrInvalidCredentials
	}

	purpose, err := ctx.Challenges.Consume(request.Challenge)
	if err != nil || *purpose != challenge.PurposeAnonymousRequest {
		return nil, newinteraction.ErrInvalidCredentials
	}

	anonIdentity, err := ctx.AnonymousIdentities.GetByKeyID(request.KeyID)
	if errors.Is(err, identity.ErrIdentityNotFound) {
		anonIdentity = nil
	} else if err != nil {
		return nil, err
	}

	if anonIdentity != nil {
		// Key ID has associated identity =>
		// verify the JWT signature before proceeding to use the key ID.
		request, err = ctx.AnonymousIdentities.ParseRequest(jwt, anonIdentity)
		if err != nil {
			return nil, newinteraction.ErrInvalidCredentials
		}
	} else if request.Key == nil {
		// No associated identity => a new key must be provided.
		return nil, newinteraction.ErrInvalidCredentials
	}

	key, err := json.Marshal(request.Key)
	if err != nil {
		return nil, err
	}

	spec := &identity.Spec{
		Type: authn.IdentityTypeAnonymous,
		Claims: map[string]interface{}{
			identity.IdentityClaimAnonymousKeyID: request.KeyID,
			identity.IdentityClaimAnonymousKey:   string(key),
		},
	}

	return &NodeUseIdentityAnonymous{
		IdentitySpec: spec,
		Action:       request.Action,
	}, nil
}

type NodeUseIdentityAnonymous struct {
	IsCreating   bool                    `json:"is_creating"`
	IdentitySpec *identity.Spec          `json:"identity_spec"`
	Action       anonymous.RequestAction `json:"action"`
}

func (n *NodeUseIdentityAnonymous) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeUseIdentityAnonymous) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeUseIdentityAnonymous) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	if n.IsCreating {
		return []newinteraction.Edge{&EdgeCreateIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
	}
	return []newinteraction.Edge{&EdgeSelectIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
}
