package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodePromptCreatePasskeyBegin{})
	interaction.RegisterNode(&NodePromptCreatePasskeyEnd{})
}

type EdgePromptCreatePasskeyBegin struct{}

func (e *EdgePromptCreatePasskeyBegin) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	userID := graph.MustGetUserID()
	ais, err := ctx.Authenticators.List(goCtx,
		userID,
		authenticator.KeepKind(authenticator.KindPrimary),
		authenticator.KeepType(model.AuthenticatorTypePasskey),
	)
	if err != nil {
		return nil, err
	}

	// Check if the identity being used needs passkey.
	// We MUST NOT check all identities because some identity
	// is not used interactively, e.g. anonymous and biometric.
	// And some identity like oauth does not use passkey at all.
	needPasskey := false
	iden, ok := graph.GetUserLastIdentity()
	if ok {
		types := iden.PrimaryAuthenticatorTypes()
		for _, typ := range types {
			if typ == model.AuthenticatorTypePasskey {
				needPasskey = true
			}
		}
	}

	passkeyEnabled := false
	for _, typ := range *ctx.Config.Authentication.PrimaryAuthenticators {
		if typ == model.AuthenticatorTypePasskey {
			passkeyEnabled = true
		}
	}

	if !passkeyEnabled || len(ais) > 0 || !needPasskey {
		return &NodePromptCreatePasskeyEnd{}, nil
	}

	return &NodePromptCreatePasskeyBegin{}, nil
}

type NodePromptCreatePasskeyBegin struct{}

func (n *NodePromptCreatePasskeyBegin) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodePromptCreatePasskeyBegin) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodePromptCreatePasskeyBegin) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{&EdgePromptCreatePasskey{}}, nil
}

type InputPromptCreatePasskey interface {
	IsSkipped() bool
	GetAttestationResponse() []byte
}

type EdgePromptCreatePasskey struct{}

func (e *EdgePromptCreatePasskey) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputPromptCreatePasskey
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	node := &NodePromptCreatePasskeyEnd{}
	if input.IsSkipped() {
		return node, nil
	}

	userID := graph.MustGetUserID()
	authenticatorSpec := &authenticator.Spec{
		UserID:    userID,
		IsDefault: false,
		Kind:      authenticator.KindPrimary,
		Type:      model.AuthenticatorTypePasskey,
		Passkey: &authenticator.PasskeySpec{
			AttestationResponse: input.GetAttestationResponse(),
		},
	}
	authenticatorInfo, err := ctx.Authenticators.New(goCtx, authenticatorSpec)
	if err != nil {
		return nil, err
	}
	node.Authenticator = authenticatorInfo

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

	return node, nil
}

type NodePromptCreatePasskeyEnd struct {
	Authenticator *authenticator.Info `json:"authenticator"`
	Identity      *identity.Info      `json:"identity"`
}

func (n *NodePromptCreatePasskeyEnd) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodePromptCreatePasskeyEnd) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	// If a primary passkey authenticator is being created,
	// we create the passkey identity here instead of using the NodeDoCreateIdentity to do so.
	// NodeDoCreateIdentity does a lot of things that is irrelevant to passkey identity,
	// such as dispatching events.
	return []interaction.Effect{
		interaction.EffectRun(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if n.Authenticator != nil {
				err := ctx.Authenticators.Create(goCtx, n.Authenticator, true)
				if err != nil {
					return err
				}
			}
			if n.Identity != nil {
				err := ctx.Identities.Create(goCtx, n.Identity)
				if err != nil {
					return err
				}
			}
			return nil
		}),
		interaction.EffectOnCommit(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if n.Authenticator != nil {
				attestationResponse := n.Authenticator.Passkey.AttestationResponse

				err := ctx.Passkey.ConsumeAttestationResponse(goCtx, attestationResponse)
				if err != nil {
					return err
				}
			}

			return nil
		}),
	}, nil
}

func (n *NodePromptCreatePasskeyEnd) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}
