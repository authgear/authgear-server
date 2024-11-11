package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCreateAuthenticatorPasskey{})
}

type InputCreateAuthenticatorPasskey interface {
	GetAttestationResponse() []byte
}

type EdgeCreateAuthenticatorPasskey struct {
	NewAuthenticatorID string
	Stage              authn.AuthenticationStage
	IsDefault          bool
}

func (e *EdgeCreateAuthenticatorPasskey) AuthenticatorType() model.AuthenticatorType {
	return model.AuthenticatorTypePasskey
}

func (e *EdgeCreateAuthenticatorPasskey) IsDefaultAuthenticator() bool {
	return false
}

func (e *EdgeCreateAuthenticatorPasskey) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var stageInput InputAuthenticationStage
	if !interaction.Input(rawInput, &stageInput) {
		return nil, interaction.ErrIncompatibleInput
	}
	stage := stageInput.GetAuthenticationStage()
	if stage != e.Stage {
		return nil, interaction.ErrIncompatibleInput
	}

	var input InputCreateAuthenticatorPasskey
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	node := &NodeCreateAuthenticatorPasskey{
		Stage: e.Stage,
	}

	userID := graph.MustGetUserID()
	authenticatorSpec := &authenticator.Spec{
		UserID:    userID,
		IsDefault: e.IsDefault,
		Kind:      stageToAuthenticatorKind(e.Stage),
		Type:      model.AuthenticatorTypePasskey,
		Passkey: &authenticator.PasskeySpec{
			AttestationResponse: input.GetAttestationResponse(),
		},
	}
	authenticatorInfo, err := ctx.Authenticators.NewWithAuthenticatorID(goCtx, e.NewAuthenticatorID, authenticatorSpec)
	if err != nil {
		return nil, err
	}
	node.Authenticator = authenticatorInfo

	if stage == authn.AuthenticationStagePrimary {
		identitySpec := &identity.Spec{
			Type: model.IdentityTypePasskey,
			Passkey: &identity.PasskeySpec{
				AttestationResponse: input.GetAttestationResponse(),
			},
		}
		identityInfo, err := ctx.Identities.New(goCtx, userID, identitySpec, identity.NewIdentityOptions{})
		if err != nil {
			return nil, err
		}
		node.Identity = identityInfo
	}

	return node, nil
}

type NodeCreateAuthenticatorPasskey struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
	Identity      *identity.Info            `json:"identity"`
}

func (n *NodeCreateAuthenticatorPasskey) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorPasskey) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	// If a primary passkey authenticator is being created,
	// we create the passkey identity here instead of using the NodeDoCreateIdentity to do so.
	// NodeDoCreateIdentity does a lot of things that is irrelevant to passkey identity,
	// such as dispatching events.
	return []interaction.Effect{
		interaction.EffectRun(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if n.Identity != nil {
				err := ctx.Identities.Create(goCtx, n.Identity)
				if err != nil {
					return err
				}
			}
			return nil
		}),
		interaction.EffectOnCommit(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			attestationResponse := n.Authenticator.Passkey.AttestationResponse

			err := ctx.Passkey.ConsumeAttestationResponse(goCtx, attestationResponse)
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}

func (n *NodeCreateAuthenticatorPasskey) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeCreateAuthenticatorEnd{
			Stage:          n.Stage,
			Authenticators: []*authenticator.Info{n.Authenticator},
		},
	}, nil
}
