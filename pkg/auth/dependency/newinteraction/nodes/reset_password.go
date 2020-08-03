package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeResetPasswordBegin{})
	newinteraction.RegisterNode(&NodeResetPasswordEnd{})
}

type NodeResetPasswordBegin struct{}

func (n *NodeResetPasswordBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeResetPasswordBegin) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{&EdgeResetPassword{}}, nil
}

type InputResetPassword interface {
	GetCode() string
	GetNewPassword() string
}

type EdgeResetPassword struct{}

func (e *EdgeResetPassword) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputResetPassword)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	code := input.GetCode()
	newPassword := input.GetNewPassword()

	codeHash := ctx.ResetPassword.HashCode(code)
	oldInfo, newInfo, updateInfo, err := ctx.ResetPassword.ResetPassword(code, newPassword)
	if err != nil {
		return nil, err
	}

	err = ctx.ResetPassword.AfterResetPassword(codeHash)
	if err != nil {
		return nil, err
	}

	return &NodeResetPasswordEnd{
		OldAuthenticator:    oldInfo,
		NewAuthenticator:    newInfo,
		UpdateAuthenticator: updateInfo,
	}, nil
}

type NodeResetPasswordEnd struct {
	OldAuthenticator    *authenticator.Info `json:"old_authenticator,omitempty"`
	NewAuthenticator    *authenticator.Info `json:"new_authenticator,omitempty"`
	UpdateAuthenticator *authenticator.Info `json:"update_authenticator,omitempty"`
}

func (n *NodeResetPasswordEnd) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeResetPasswordEnd) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	// Password authenticator is always primary for now
	if n.NewAuthenticator != nil {
		return []newinteraction.Edge{
			&EdgeDoCreateAuthenticator{
				Stage:          newinteraction.AuthenticationStagePrimary,
				Authenticators: []*authenticator.Info{n.NewAuthenticator},
			},
		}, nil
	} else if n.UpdateAuthenticator != nil {
		return []newinteraction.Edge{
			&EdgeDoUpdateAuthenticator{
				Stage:                     newinteraction.AuthenticationStagePrimary,
				AuthenticatorBeforeUpdate: n.OldAuthenticator,
				AuthenticatorAfterUpdate:  n.NewAuthenticator,
			},
		}, nil
	} else {
		// Password is not changed, ends the interaction
		return nil, nil
	}
}
