package nodes

import (
	"encoding/json"
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeUseIdentityAnonymous{})
}

type InputUseIdentityAnonymous interface {
	GetAnonymousRequestToken() string
}

type EdgeUseIdentityAnonymous struct {
	IsCreating bool
}

func (e *EdgeUseIdentityAnonymous) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	input, ok := rawInput.(InputUseIdentityAnonymous)
	if !ok {
		return nil, interaction.ErrIncompatibleInput
	}

	enabled := false
	for _, t := range ctx.Config.Authentication.Identities {
		if t == authn.IdentityTypeAnonymous {
			enabled = true
			break
		}
	}

	if !enabled {
		return nil, interaction.NewInvariantViolated(
			"AnonymousUserDisallowed",
			"anonymous users are not allowed",
			nil,
		)
	}

	jwt := input.GetAnonymousRequestToken()

	request, err := ctx.AnonymousIdentities.ParseRequestUnverified(jwt)
	if err != nil {
		return nil, interaction.ErrInvalidCredentials
	}

	purpose, err := ctx.Challenges.Consume(request.Challenge)
	if err != nil || *purpose != challenge.PurposeAnonymousRequest {
		return nil, interaction.ErrInvalidCredentials
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
			return nil, interaction.ErrInvalidCredentials
		}
	} else if request.Key == nil {
		// No associated identity => a new key must be provided.
		return nil, interaction.ErrInvalidCredentials
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

func (n *NodeUseIdentityAnonymous) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeUseIdentityAnonymous) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	return nil
}

func (n *NodeUseIdentityAnonymous) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	if n.IsCreating {
		return []interaction.Edge{&EdgeCreateIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
	}
	return []interaction.Edge{&EdgeSelectIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
}
