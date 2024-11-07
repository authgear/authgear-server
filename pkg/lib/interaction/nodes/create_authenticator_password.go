package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCreateAuthenticatorPassword{})
}

type InputCreateAuthenticatorPassword interface {
	GetPassword() string
}

type EdgeCreateAuthenticatorPassword struct {
	NewAuthenticatorID string
	Stage              authn.AuthenticationStage
	IsDefault          bool
}

func (e *EdgeCreateAuthenticatorPassword) AuthenticatorType() model.AuthenticatorType {
	return model.AuthenticatorTypePassword
}

func (e *EdgeCreateAuthenticatorPassword) IsDefaultAuthenticator() bool {
	return false
}

func (e *EdgeCreateAuthenticatorPassword) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var stageInput InputAuthenticationStage
	if !interaction.Input(rawInput, &stageInput) {
		return nil, interaction.ErrIncompatibleInput
	}
	stage := stageInput.GetAuthenticationStage()
	if stage != e.Stage {
		return nil, interaction.ErrIncompatibleInput
	}

	var input InputCreateAuthenticatorPassword
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	userID := graph.MustGetUserID()
	spec := &authenticator.Spec{
		UserID:    userID,
		IsDefault: e.IsDefault,
		Kind:      stageToAuthenticatorKind(e.Stage),
		Type:      model.AuthenticatorTypePassword,
		Password: &authenticator.PasswordSpec{
			PlainPassword: input.GetPassword(),
		},
	}

	info, err := ctx.Authenticators.NewWithAuthenticatorID(goCtx, e.NewAuthenticatorID, spec)
	if err != nil {
		return nil, err
	}

	return &NodeCreateAuthenticatorPassword{Stage: e.Stage, Authenticator: info}, nil
}

type NodeCreateAuthenticatorPassword struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
}

func (n *NodeCreateAuthenticatorPassword) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorPassword) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorPassword) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeCreateAuthenticatorEnd{
			Stage:          n.Stage,
			Authenticators: []*authenticator.Info{n.Authenticator},
		},
	}, nil
}
