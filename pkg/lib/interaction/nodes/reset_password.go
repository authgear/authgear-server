package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeResetPasswordBegin{})
	interaction.RegisterNode(&NodeResetPasswordEnd{})
}

type NodeResetPasswordBegin struct{}

func (n *NodeResetPasswordBegin) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeResetPasswordBegin) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	return nil
}

func (n *NodeResetPasswordBegin) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{&EdgeResetPassword{}}, nil
}

type InputResetPassword interface {
	GetCode() string
	GetNewPassword() string
}

type EdgeResetPassword struct{}

func (e *EdgeResetPassword) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputResetPassword
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	code := input.GetCode()
	newPassword := input.GetNewPassword()

	codeHash := ctx.ResetPassword.HashCode(code)
	oldInfo, newInfo, err := ctx.ResetPassword.ResetPassword(code, newPassword)
	if err != nil {
		return nil, err
	}

	err = ctx.ResetPassword.AfterResetPassword(codeHash)
	if err != nil {
		return nil, err
	}

	return &NodeResetPasswordEnd{
		OldAuthenticator: oldInfo,
		NewAuthenticator: newInfo,
	}, nil
}

type NodeResetPasswordEnd struct {
	OldAuthenticator *authenticator.Info `json:"old_authenticator"`
	NewAuthenticator *authenticator.Info `json:"new_authenticator,omitempty"`
}

func (n *NodeResetPasswordEnd) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	return nil
}

func (n *NodeResetPasswordEnd) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeResetPasswordEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	// Password authenticator is always primary for now
	if n.NewAuthenticator != nil {
		return []interaction.Edge{
			&EdgeDoUpdateAuthenticator{
				Stage:                     interaction.AuthenticationStagePrimary,
				AuthenticatorBeforeUpdate: n.OldAuthenticator,
				AuthenticatorAfterUpdate:  n.NewAuthenticator,
			},
		}, nil
	}

	// Password is not changed, ends the interaction
	return nil, nil
}
